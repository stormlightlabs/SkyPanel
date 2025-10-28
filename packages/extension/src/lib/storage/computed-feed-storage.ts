import { browser } from "wxt/browser";
import type { ComputedFeedCache, ComputedFeedKind, CachedComputedFeed } from "$lib/types/computed-feed";

const STORAGE_KEY = "skypanel.computedFeeds";

/**
 * TTL (time-to-live) for mutuals feed cache in milliseconds.
 *
 * Mutuals computation is expensive (fetches all follows and followers),
 * but the data changes infrequently. 30 minutes balances accuracy and performance.
 *
 * NOTE: This could be made user-configurable in settings.
 */
const MUTUALS_TTL_MS = 30 * 60 * 1000;

/**
 * TTL (time-to-live) for quiet posters feed cache in milliseconds.
 *
 * Quiet posters computation is very expensive (fetches author feeds for all follows),
 * and the data changes slowly. 2 hours balances accuracy and performance.
 *
 * NOTE: This could be made user-configurable in settings.
 */
const QUIET_TTL_MS = 2 * 60 * 60 * 1000;

/**
 * Type-safe wrapper around chrome.storage.local for computed feed caching.
 *
 * Provides TTL-based expiration to balance accuracy (fresh data) with
 * performance (avoiding expensive recomputation).
 *
 * Cache invalidation strategy:
 * - Automatic expiration based on TTL
 * - Manual refresh via forceRefresh flag in requests
 * - Clear on logout (handled by parent service)
 */
export class ComputedFeedStorage {
  constructor(private readonly storageKey: string = STORAGE_KEY) {}

  /**
   * Access chrome.storage.local with availability check.
   */
  private get storage() {
    const storage = browser.storage?.local;
    if (!storage) {
      console.warn("[ComputedFeedStorage] storage.local is unavailable; skipping persistence.");
    }
    return storage;
  }

  /**
   * Get TTL for a specific feed kind.
   */
  private getTTL(kind: ComputedFeedKind): number {
    switch (kind) {
      case "mutuals":
        return MUTUALS_TTL_MS;
      case "quiet":
        return QUIET_TTL_MS;
    }
  }

  /**
   * Load cached computed feed from storage.
   *
   * Returns undefined if:
   * - Storage is unavailable
   * - Feed not found in cache
   * - Cache has expired (past TTL)
   *
   * @param kind - Type of computed feed to load
   * @returns Cached feed data, or undefined if not found or expired
   */
  async load(kind: ComputedFeedKind): Promise<CachedComputedFeed | undefined> {
    const storage = this.storage;
    if (!storage) {
      return undefined;
    }

    try {
      const result = await storage.get(this.storageKey);
      const cache = result[this.storageKey] as ComputedFeedCache | undefined;

      if (!cache) {
        return undefined;
      }

      const cached = cache[kind];
      if (!cached) {
        return undefined;
      }

      if (Date.now() > cached.expiresAt) {
        console.log(`[ComputedFeedStorage] Cache expired for ${kind} feed`);
        return undefined;
      }

      const remainingMs = cached.expiresAt - Date.now();
      const remainingMin = Math.floor(remainingMs / 60000);
      console.log(`[ComputedFeedStorage] Cache hit for ${kind} feed (expires in ${remainingMin}m)`);

      return cached;
    } catch (error) {
      console.error("[ComputedFeedStorage] Failed to load cache", { kind, error });
      return undefined;
    }
  }

  /**
   * Save computed feed to cache with TTL-based expiration.
   *
   * @param cached - Computed feed data to cache
   */
  async save(cached: CachedComputedFeed): Promise<void> {
    const storage = this.storage;
    if (!storage) {
      return;
    }

    try {
      const result = await storage.get(this.storageKey);
      const cache = (result[this.storageKey] as ComputedFeedCache | undefined) ?? {};

      cache[cached.kind] = cached;

      await storage.set({ [this.storageKey]: cache });

      const ttlMin = Math.floor((cached.expiresAt - cached.computedAt) / 60000);
      console.log(`[ComputedFeedStorage] Cached ${cached.kind} feed (TTL: ${ttlMin}m)`);
    } catch (error) {
      console.error("[ComputedFeedStorage] Failed to save cache", { kind: cached.kind, error });
    }
  }

  /**
   * Create a cached feed entry with automatic TTL calculation.
   *
   * @param kind - Type of computed feed
   * @param data - Computed feed result
   * @returns Cached feed with computed expiration time
   */
  createCached(kind: ComputedFeedKind, data: CachedComputedFeed["data"]): CachedComputedFeed {
    const now = Date.now();
    const ttl = this.getTTL(kind);

    return { kind, data, computedAt: now, expiresAt: now + ttl };
  }

  /**
   * Clear specific computed feed from cache.
   *
   * @param kind - Type of computed feed to clear
   */
  async clear(kind: ComputedFeedKind): Promise<void> {
    const storage = this.storage;
    if (!storage) {
      return;
    }

    try {
      const result = await storage.get(this.storageKey);
      const cache = result[this.storageKey] as ComputedFeedCache | undefined;

      if (!cache) {
        return;
      }

      delete cache[kind];

      await storage.set({ [this.storageKey]: cache });
      console.log(`[ComputedFeedStorage] Cleared ${kind} feed from cache`);
    } catch (error) {
      console.error("[ComputedFeedStorage] Failed to clear cache", { kind, error });
    }
  }

  /**
   * Clear all computed feeds from cache.
   *
   * Useful on logout or when user explicitly requests cache refresh.
   */
  async clearAll(): Promise<void> {
    const storage = this.storage;
    if (!storage) {
      return;
    }

    try {
      await storage.remove(this.storageKey);
      console.log("[ComputedFeedStorage] Cleared all computed feed caches");
    } catch (error) {
      console.error("[ComputedFeedStorage] Failed to clear all caches", error);
    }
  }
}

/**
 * Singleton instance for computed feed storage operations.
 */
export const computedFeedStorage = new ComputedFeedStorage();
