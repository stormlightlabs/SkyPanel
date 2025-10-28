import { browser } from 'wxt/browser';
import type { ProfileViewDetailed } from '$lib/types/profile';

const STORAGE_KEY = 'skypanel.profiles';

/**
 * TTL (time-to-live) for profile cache in milliseconds.
 *
 * Profile data changes infrequently (avatar, bio, follower counts).
 */
const PROFILE_TTL_MS = 10 * 60 * 1000;

/**
 * Cached profile entry with TTL metadata.
 */
export type CachedProfile = { actor: string; profile: ProfileViewDetailed; fetchedAt: number; expiresAt: number };

/**
 * Profile cache storage structure.
 *
 * Maps actor DID/handle to cached profile data.
 */
type ProfileCache = Record<string, CachedProfile>;

/**
 * Type-safe wrapper around chrome.storage.local for profile caching.
 *
 * Provides TTL-based expiration to reduce API calls while keeping
 * profile data reasonably fresh. Profiles are keyed by actor DID/handle
 * to support caching multiple profiles.
 *
 * Cache invalidation strategy:
 * - Automatic expiration based on 10-minute TTL
 * - Manual refresh via forceRefresh flag in requests
 * - Clear all on logout/session start
 */
export class ProfileStorage {
	constructor(private readonly storageKey: string = STORAGE_KEY) {}

	/**
	 * Access {@link chrome.storage.local} with availability check.
	 */
	private get storage() {
		const storage = browser.storage?.local;
		if (!storage) {
			console.warn('[ProfileStorage] storage.local is unavailable; skipping persistence.');
		}
		return storage;
	}

	/**
	 * Load cached profile from storage.
	 *
	 * Returns undefined if:
	 * - Storage is unavailable
	 * - Profile not found in cache
	 * - Cache has expired (past TTL)
	 *
	 * @param actor - DID or handle of the profile to load
	 * @returns Cached profile data, or undefined if not found or expired
	 */
	async load(actor: string): Promise<CachedProfile | undefined> {
		const storage = this.storage;
		if (!storage) {
			return undefined;
		}

		try {
			const result = await storage.get(this.storageKey);
			const cache = result[this.storageKey] as ProfileCache | undefined;

			if (!cache) {
				return undefined;
			}

			const cached = cache[actor];
			if (!cached) {
				return undefined;
			}

			if (Date.now() > cached.expiresAt) {
				console.log(`[ProfileStorage] Cache expired for ${actor}`);
				return undefined;
			}

			const remainingMs = cached.expiresAt - Date.now();
			const remainingMin = Math.floor(remainingMs / 60_000);
			console.log(`[ProfileStorage] Cache hit for ${actor} (expires in ${remainingMin}m)`);

			return cached;
		} catch (error) {
			console.error('[ProfileStorage] Failed to load cache', { actor, error });
			return undefined;
		}
	}

	/**
	 * Save profile to cache with TTL-based expiration.
	 *
	 * @param cached - Profile data to cache
	 */
	async save(cached: CachedProfile): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		try {
			const result = await storage.get(this.storageKey);
			const cache = (result[this.storageKey] as ProfileCache | undefined) ?? {};

			cache[cached.actor] = cached;

			await storage.set({ [this.storageKey]: cache });

			const ttlMin = Math.floor((cached.expiresAt - cached.fetchedAt) / 60_000);
			console.log(`[ProfileStorage] Cached ${cached.actor} profile (TTL: ${ttlMin}m)`);
		} catch (error) {
			console.error('[ProfileStorage] Failed to save cache', { actor: cached.actor, error });
		}
	}

	/**
	 * Create a cached profile entry with automatic TTL calculation.
	 *
	 * @param actor - DID or handle of the profile
	 * @param profile - Profile data from API
	 * @returns Cached profile with computed expiration time
	 */
	createCached(actor: string, profile: ProfileViewDetailed): CachedProfile {
		const now = Date.now();

		return { actor, profile, fetchedAt: now, expiresAt: now + PROFILE_TTL_MS };
	}

	/**
	 * Clear specific profile from cache.
	 *
	 * @param actor - DID or handle of the profile to clear
	 */
	async clear(actor: string): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		try {
			const result = await storage.get(this.storageKey);
			const cache = result[this.storageKey] as ProfileCache | undefined;

			if (!cache) {
				return;
			}

			delete cache[actor];

			await storage.set({ [this.storageKey]: cache });
			console.log(`[ProfileStorage] Cleared ${actor} from cache`);
		} catch (error) {
			console.error('[ProfileStorage] Failed to clear cache', { actor, error });
		}
	}

	/**
	 * Clear all profiles from cache.
	 *
	 * Useful on logout or session start to ensure fresh data.
	 */
	async clearAll(): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		try {
			await storage.remove(this.storageKey);
			console.log('[ProfileStorage] Cleared all profile caches');
		} catch (error) {
			console.error('[ProfileStorage] Failed to clear all caches', error);
		}
	}
}

export const profileStorage = new ProfileStorage();
