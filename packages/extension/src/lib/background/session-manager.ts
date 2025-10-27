import { AtpAgent, type AtpPersistSessionHandler, type AtpSessionData } from "@atproto/api";
import { sessionStorage } from "$lib/storage/session-storage";
import { DEFAULT_SERVICE_URL, type PersistedSession, type SessionSnapshot } from "$lib/types/session";

type SessionListener = (snapshot?: SessionSnapshot) => void;

export class SessionManager {
  private agentRef: AtpAgent;
  private current?: SessionSnapshot;
  private listeners = new Set<SessionListener>();

  constructor(private readonly serviceUrl: string = DEFAULT_SERVICE_URL) {
    this.agentRef = this.createAgent(serviceUrl);
  }

  get agent(): AtpAgent {
    return this.agentRef;
  }

  get snapshot(): SessionSnapshot | undefined {
    return this.current;
  }

  get authenticated(): boolean {
    return this.agentRef.hasSession && !!this.current;
  }

  subscribe(listener: SessionListener): () => void {
    this.listeners.add(listener);
    listener(this.current);
    return () => {
      this.listeners.delete(listener);
    };
  }

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
      console.warn("[SessionManager] resume failed, clearing persisted session", error);
      await sessionStorage.clear();
      this.current = undefined;
      this.notify();
      return;
    }
  }

  async login(identifier: string, password: string): Promise<SessionSnapshot> {
    const response = await this.agentRef.login({ identifier, password });
    const session = response.data as AtpSessionData;
    const snapshot = this.snapshotFrom(session);
    this.current = snapshot;
    await this.persist({ session, serviceUrl: this.agentRef.serviceUrl.toString(), storedAt: Date.now() });
    this.notify(snapshot);
    return snapshot;
  }

  async logout(): Promise<void> {
    if (this.agentRef.hasSession) {
      try {
        await this.agentRef.logout();
      } catch (error) {
        console.warn("[SessionManager] logout failed", error);
      }
    }
    this.current = undefined;
    await sessionStorage.clear();
    this.notify();
  }

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

  private createAgent(url: string): AtpAgent {
    return new AtpAgent({ service: url, persistSession: this.handlePersistSession });
  }

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

  private async persist(persisted: PersistedSession): Promise<void> {
    await sessionStorage.save(persisted);
  }

  private notify(snapshot?: SessionSnapshot): void {
    for (const listener of this.listeners) {
      listener(snapshot);
    }
  }
}
