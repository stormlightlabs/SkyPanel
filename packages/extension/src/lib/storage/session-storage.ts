import { browser } from "wxt/browser";
import type { PersistedSession } from "$lib/types/session";

const STORAGE_KEY = "skypanel.session";

/**
 * Type-safe wrapper around chrome.storage.local for session persistence.
 *
 * Provides graceful degradation when storage is unavailable (e.g., in tests).
 * Uses a single namespaced key to avoid conflicts with other extensions.
 */
export class SessionStorage {
  constructor(private readonly storageKey: string = STORAGE_KEY) {}

  /**
   * Access chrome.storage.local with availability check.
   */
  private get storage() {
    const storage = browser.storage?.local;
    if (!storage) {
      console.warn("[SessionStorage] storage.local is unavailable; skipping persistence.");
    }
    return storage;
  }

  /**
   * Load persisted session from storage.
   *
   * @returns The persisted session, or null if not found or storage unavailable
   */
  async load(): Promise<PersistedSession | null> {
    const storage = this.storage;
    if (!storage) {
      return null;
    }
    const result = await storage.get(this.storageKey);
    const raw = result[this.storageKey] as PersistedSession | undefined;
    if (!raw) {
      return null;
    }
    return raw;
  }

  /**
   * Save session to storage.
   *
   * Silently fails if storage is unavailable.
   */
  async save(session: PersistedSession): Promise<void> {
    const storage = this.storage;
    if (!storage) {
      return;
    }
    await storage.set({ [this.storageKey]: session });
  }

  /**
   * Remove session from storage.
   *
   * Silently fails if storage is unavailable.
   */
  async clear(): Promise<void> {
    const storage = this.storage;
    if (!storage) {
      return;
    }
    await storage.remove(this.storageKey);
  }
}

/**
 * Singleton instance for session storage operations.
 */
export const sessionStorage = new SessionStorage();
