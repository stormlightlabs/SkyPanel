import { browser } from "wxt/browser";
import { backgroundClient } from "$lib/client/background-client";
import { isSessionChangedEvent } from "$lib/messaging/messages";
import type { SessionSnapshot } from "$lib/types/session";

/**
 * Manages session state for the extension UI.
 *
 * Provides reactive state management for authentication, coordinating between
 * the UI and background script via BackgroundClient. Handles session hydration,
 * login, logout, and automatic updates from background session changes.
 */
class SessionStore {
  private static instance: SessionStore;

  private session = $state<SessionSnapshot>();
  private status = $state<"idle" | "loading" | "error">("idle");
  private errorMessage = $state<string>();
  private hydrated = $state(false);
  private hydratePromise = $state<Promise<void>>();

  private constructor() {
    browser.runtime.onMessage.addListener(this.handleSessionChanged);
  }

  static getInstance(): SessionStore {
    if (!SessionStore.instance) {
      SessionStore.instance = new SessionStore();
    }
    return SessionStore.instance;
  }

  get currentSession() {
    return this.session;
  }

  get currentStatus() {
    return this.status;
  }

  get error() {
    return this.errorMessage;
  }

  get isHydrated() {
    return this.hydrated;
  }

  get isAuthenticated() {
    return !!this.session;
  }

  private handleSessionChanged = (message: unknown) => {
    if (!isSessionChangedEvent(message)) {
      return;
    }
    console.log("[session-store] session changed event:", message);
    this.session = message.session ?? undefined;
    this.errorMessage = undefined;
    this.status = "idle";
  };

  async hydrate(): Promise<void> {
    console.log("[session-store] hydrate called, hydrated:", this.hydrated, "hydratePromise:", !!this.hydratePromise);

    if (this.hydratePromise) {
      console.log("[session-store] returning existing hydrate promise");
      return this.hydratePromise;
    }

    if (this.hydrated) {
      console.log("[session-store] already hydrated, skipping");
      return;
    }

    console.log("[session-store] starting hydration");
    this.status = "loading";
    this.hydratePromise = (async () => {
      try {
        console.log("[session-store] fetching session from background");
        const response = await backgroundClient.getSession();
        console.log("[session-store] received response:", response);
        this.session = response.session ?? undefined;
        this.errorMessage = undefined;
        this.status = "idle";
        console.log("[session-store] hydration complete, session:", !!this.session);
      } catch (error) {
        console.error("[session-store] hydrate failed", error);
        this.errorMessage = error instanceof Error ? error.message : "Unable to load session";
        this.status = "error";
      } finally {
        console.log("[session-store] setting hydrated = true");
        this.hydrated = true;
        this.hydratePromise = undefined;
      }
    })();

    await this.hydratePromise;
    console.log("[session-store] hydrate finished");
  }

  async login(identifier: string, password: string): Promise<boolean> {
    console.log("[session-store] login called with identifier:", identifier);
    this.status = "loading";
    this.errorMessage = undefined;
    try {
      console.log("[session-store] calling backgroundClient.login");
      const response = await backgroundClient.login(identifier, password);
      console.log("[session-store] login response:", response);
      if (!response.ok) {
        console.error("[session-store] login failed:", response.error);
        this.errorMessage = response.error;
        this.status = "idle";
        return false;
      }
      console.log("[session-store] login succeeded, setting session");
      this.session = response.session;
      this.status = "idle";
      return true;
    } catch (error) {
      console.error("[session-store] login exception:", error);
      this.errorMessage = error instanceof Error ? error.message : "Unable to login";
      this.status = "idle";
      return false;
    }
  }

  async logout(): Promise<void> {
    this.status = "loading";
    this.errorMessage = undefined;
    try {
      await backgroundClient.logout();
      this.session = undefined;
    } catch (error) {
      console.error("[session-store] logout failed", error);
      this.errorMessage = error instanceof Error ? error.message : "Unable to logout";
    } finally {
      this.status = "idle";
    }
  }

  async refresh(): Promise<void> {
    this.status = "loading";
    try {
      const response = await backgroundClient.getSession();
      this.session = response.session ?? undefined;
      this.errorMessage = undefined;
      this.status = "idle";
    } catch (error) {
      console.error("[session-store] refresh failed", error);
      this.errorMessage = error instanceof Error ? error.message : "Unable to refresh session";
      this.status = "error";
    }
  }
}

export const sessionStore = SessionStore.getInstance();
