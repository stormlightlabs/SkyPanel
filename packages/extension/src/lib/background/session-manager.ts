import { AtpAgent, type AtpPersistSessionHandler, type AtpSessionData } from "@atproto/api";
import { sessionStorage } from "$lib/storage/session-storage";
import { DEFAULT_SERVICE_URL, type PersistedSession, type SessionSnapshot } from "$lib/types/session";

/**
 * Callback function invoked when the session state changes.
 * Receives the current session snapshot or undefined if logged out.
 */
type SessionListener = (snapshot?: SessionSnapshot) => void;

/**
 * Manages authentication session lifecycle for the Bluesky ATP agent.
 *
 * Responsibilities:
 * - Wraps AtpAgent with session persistence to chrome.storage.local
 * - Handles login, logout, and automatic session resume on startup
 * - Notifies subscribers of session state changes via pub/sub pattern
 * - Maintains session consistency between agent state and storage
 *
 * Session Lifecycle:
 * 1. On startup: resumeFromStorage() attempts to restore previous session
 * 2. On login: Credentials → AtpAgent → Storage → Notify subscribers
 * 3. On logout: Clear agent → Clear storage → Notify subscribers
 * 4. On token refresh: AtpAgent auto-refreshes → handlePersistSession → Update storage
 */
export class SessionManager {
  private agentRef: AtpAgent;
  private current?: SessionSnapshot;
  private listeners = new Set<SessionListener>();

  constructor(serviceUrl: string = DEFAULT_SERVICE_URL) {
    this.agentRef = this.createAgent(serviceUrl);
  }

  /**
   * Access the underlying AtpAgent for making authenticated API calls.
   */
  get agent(): AtpAgent {
    return this.agentRef;
  }

  /**
   * Current session snapshot containing user identity and status.
   * Returns undefined if not authenticated.
   */
  get snapshot(): SessionSnapshot | undefined {
    return this.current;
  }

  /**
   * Whether the user is currently authenticated with a valid session.
   */
  get authenticated(): boolean {
    return this.agentRef.hasSession && !!this.current;
  }

  /**
   * Subscribe to session state changes.
   *
   * The listener is invoked immediately with the current state,
   * and subsequently whenever the session changes (login, logout, token refresh).
   *
   * @returns Unsubscribe function to remove the listener
   */
  subscribe(listener: SessionListener): () => void {
    this.listeners.add(listener);
    listener(this.current);
    return () => {
      this.listeners.delete(listener);
    };
  }

  /**
   * Attempt to resume a previous session from chrome.storage.local.
   *
   * Called on background service worker startup to restore authentication state.
   * If the stored session is expired or invalid, it is cleared automatically.
   *
   * @returns The restored session snapshot, or undefined if no valid session exists
   */
  async resumeFromStorage(): Promise<SessionSnapshot | undefined> {
    const persisted = await sessionStorage.load();
    if (!persisted) {
      return;
    }

    if (persisted.serviceUrl && persisted.serviceUrl !== this.agentRef.serviceUrl.toString()) {
      this.agentRef = this.createAgent(persisted.serviceUrl);
    }

    try {
      await this.agentRef.resumeSession(persisted.session);
      const session = this.agentRef.session ?? persisted.session;
      const snapshot = this.snapshotFrom(session);
      this.current = snapshot;
      await this.persist({ session, serviceUrl: this.agentRef.serviceUrl.toString(), storedAt: Date.now() });
      this.notify(snapshot);
      return snapshot;
    } catch (error) {
      console.warn("[SessionManager] Resume failed, clearing persisted session", error);
      await sessionStorage.clear();
      this.current = undefined;
      this.notify();
      return;
    }
  }

  /**
   * Authenticate with Bluesky using handle/email and app password.
   *
   * On success, stores the session to chrome.storage.local and notifies subscribers.
   *
   * @param identifier - User handle (e.g., "alice.bsky.social") or email
   * @param password - App password (not main account password)
   * @returns Session snapshot with user identity
   * @throws Error if login fails (invalid credentials, network error, etc.)
   */
  async login(identifier: string, password: string): Promise<SessionSnapshot> {
    if (!identifier || !password) {
      throw new Error("Identifier and password are required");
    }

    try {
      const response = await this.agentRef.login({ identifier, password });
      const session = response.data as AtpSessionData;
      const snapshot = this.snapshotFrom(session);
      this.current = snapshot;
      await this.persist({ session, serviceUrl: this.agentRef.serviceUrl.toString(), storedAt: Date.now() });
      this.notify(snapshot);
      return snapshot;
    } catch (error) {
      console.error("[SessionManager] Login failed", error);
      throw error;
    }
  }

  /**
   * Log out the current user and clear stored session.
   *
   * Attempts to invalidate the session on the server, then clears local state
   * regardless of whether the server request succeeds (best-effort cleanup).
   */
  async logout(): Promise<void> {
    if (this.agentRef.hasSession) {
      try {
        await this.agentRef.logout();
      } catch (error) {
        console.warn("[SessionManager] Server logout failed, clearing local session anyway", error);
      }
    }
    this.current = undefined;
    await sessionStorage.clear();
    this.notify();
  }

  /**
   * Callback invoked by AtpAgent when the session changes.
   *
   * This happens during token refresh or when the agent detects session expiry.
   * Updates local state and storage to stay in sync with the agent.
   */
  private readonly handlePersistSession: AtpPersistSessionHandler = async (_, session) => {
    if (!session) {
      this.current = undefined;
      await sessionStorage.clear();
      this.notify();
      return;
    }

    const snapshot = this.snapshotFrom(session);
    this.current = snapshot;
    await this.persist({ session, serviceUrl: this.agentRef.serviceUrl.toString(), storedAt: Date.now() });
    this.notify(snapshot);
  };

  /**
   * Create a new AtpAgent configured to persist session changes.
   */
  private createAgent(url: string): AtpAgent {
    return new AtpAgent({ service: url, persistSession: this.handlePersistSession });
  }

  /**
   * Convert AtpSessionData to a lightweight SessionSnapshot for UI consumption.
   * Excludes sensitive tokens while preserving user identity and status.
   */
  private snapshotFrom(session: AtpSessionData): SessionSnapshot {
    return {
      did: session.did,
      handle: session.handle,
      email: session.email,
      emailConfirmed: session.emailConfirmed,
      emailAuthFactor: session.emailAuthFactor,
      active: session.active,
      status: session.status,
      serviceUrl: this.agentRef.serviceUrl.toString(),
    };
  }

  /**
   * Persist session to chrome.storage.local with timestamp.
   */
  private async persist(persisted: PersistedSession): Promise<void> {
    await sessionStorage.save(persisted);
  }

  /**
   * Notify all subscribers of the current session state.
   */
  private notify(snapshot?: SessionSnapshot): void {
    for (const listener of this.listeners) {
      listener(snapshot);
    }
  }
}
