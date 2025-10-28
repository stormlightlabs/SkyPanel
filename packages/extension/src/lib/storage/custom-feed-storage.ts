import { browser } from 'wxt/browser';
import type { CustomFeedDefinition } from '$lib/types/custom-feed';

const STORAGE_KEY = 'skypanel.customFeeds';
const SCHEMA_VERSION = 1;

/**
 * Storage wrapper for persisting custom feed definitions.
 */
type CustomFeedStorageData = { version: number; feeds: Record<string, CustomFeedDefinition> };

/**
 * Type-safe wrapper around chrome.storage.local for custom feed persistence.
 *
 * Provides CRUD operations for user-created feed definitions.
 * Unlike computed feeds which are cached temporarily, custom feeds are permanently stored and managed by the user.
 *
 * Storage strategy:
 * - All feeds stored in a single object keyed by feed ID
 * - Versioned schema for future migrations
 * - No TTL - feeds persist until explicitly deleted
 * - Cleared on logout (optional, could be preserved)
 */
export class CustomFeedStorage {
	constructor(private readonly storageKey: string = STORAGE_KEY) {}

	/**
	 * Access chrome.storage.local with availability check.
	 */
	private get storage() {
		const storage = browser.storage?.local;
		if (!storage) {
			console.warn('[CustomFeedStorage] storage.local is unavailable; skipping persistence.');
		}
		return storage;
	}

	/**
	 * Load all custom feed definitions from storage.
	 *
	 * Returns an empty map if storage is unavailable or no feeds exist.
	 * Handles version migrations automatically.
	 */
	async loadAll(): Promise<Map<string, CustomFeedDefinition>> {
		const storage = this.storage;
		if (!storage) {
			return new Map();
		}

		try {
			const result = await storage.get(this.storageKey);
			const data = result[this.storageKey] as CustomFeedStorageData | undefined;

			if (!data || !data.feeds) {
				return new Map();
			}

			if (data.version !== SCHEMA_VERSION) {
				console.log(`[CustomFeedStorage] Migrating from version ${data.version} to ${SCHEMA_VERSION}`);
				const migrated = this.migrate(data);
				await this.saveData(migrated);
				return new Map(Object.entries(migrated.feeds));
			}

			return new Map(Object.entries(data.feeds));
		} catch (error) {
			console.error('[CustomFeedStorage] Failed to load feeds', error);
			return new Map();
		}
	}

	/**
	 * Load a single custom feed definition by ID.
	 */
	async load(feedId: string): Promise<CustomFeedDefinition | undefined> {
		const feeds = await this.loadAll();
		return feeds.get(feedId);
	}

	/**
	 * Save a custom feed definition to storage.
	 *
	 * Creates a new feed if the ID doesn't exist, otherwise updates existing.
	 */
	async save(feed: CustomFeedDefinition): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		try {
			const feeds = await this.loadAll();
			feeds.set(feed.id, feed);

			const data: CustomFeedStorageData = { version: SCHEMA_VERSION, feeds: Object.fromEntries(feeds) };

			await this.saveData(data);
			console.log(`[CustomFeedStorage] Saved feed: ${feed.name} (${feed.id})`);
		} catch (error) {
			console.error('[CustomFeedStorage] Failed to save feed', { feedId: feed.id, error });
			throw error;
		}
	}

	/**
	 * Delete a custom feed definition from storage.
	 */
	async delete(feedId: string): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		try {
			const feeds = await this.loadAll();
			const deleted = feeds.delete(feedId);

			if (!deleted) {
				console.warn(`[CustomFeedStorage] Feed not found: ${feedId}`);
				return;
			}

			const data: CustomFeedStorageData = { version: SCHEMA_VERSION, feeds: Object.fromEntries(feeds) };

			await this.saveData(data);
			console.log(`[CustomFeedStorage] Deleted feed: ${feedId}`);
		} catch (error) {
			console.error('[CustomFeedStorage] Failed to delete feed', { feedId, error });
			throw error;
		}
	}

	/**
	 * Clear all custom feed definitions from storage.
	 */
	async clearAll(): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		try {
			await storage.remove(this.storageKey);
			console.log('[CustomFeedStorage] Cleared all custom feeds');
		} catch (error) {
			console.error('[CustomFeedStorage] Failed to clear all feeds', error);
		}
	}

	/**
	 * Save storage data to chrome.storage.local.
	 */
	private async saveData(data: CustomFeedStorageData): Promise<void> {
		const storage = this.storage;
		if (!storage) {
			return;
		}

		await storage.set({ [this.storageKey]: data });
	}

	/**
	 * Migrate storage data from old version to current version.
	 *
	 * Currently a no-op since we're at version 1.
	 * Future versions should add migration logic here.
	 */
	private migrate(data: CustomFeedStorageData): CustomFeedStorageData {
		return { version: SCHEMA_VERSION, feeds: data.feeds };
	}
}

/**
 * Singleton instance for custom feed storage operations.
 */
export const customFeedStorage = new CustomFeedStorage();
