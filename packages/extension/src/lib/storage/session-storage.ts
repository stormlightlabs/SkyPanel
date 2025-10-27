import { browser } from "wxt/browser";
import type { PersistedSession } from "$lib/types/session";

const STORAGE_KEY = "skypanel.session";

export class SessionStorage {
  constructor(private readonly storageKey: string = STORAGE_KEY) {}

  async load(): Promise<PersistedSession | null> {
    const result = await browser.storage.local.get(this.storageKey);
    const raw = result[this.storageKey] as PersistedSession | undefined;
    if (!raw) {
      return null;
    }
    return raw;
  }

  async save(session: PersistedSession): Promise<void> {
    await browser.storage.local.set({ [this.storageKey]: session });
  }

  async clear(): Promise<void> {
    await browser.storage.local.remove(this.storageKey);
  }
}

export const sessionStorage = new SessionStorage();
