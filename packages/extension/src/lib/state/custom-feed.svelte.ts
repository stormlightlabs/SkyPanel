import { customFeedStorage } from '$lib/storage/custom-feed-storage';
import type { CustomFeedDefinition } from '$lib/types/custom-feed';

/**
 * Manages custom feed definitions using Svelte 5 runes.
 *
 * Provides reactive state management for user-created feed definitions,
 * coordinating with CustomFeedStorage for persistence. Handles CRUD
 * operations, loading states, and error handling.
 */
class CustomFeedStore {
	private static instance: CustomFeedStore;

	private definitions = $state(new Map<string, CustomFeedDefinition>());
	private selectedFeedId = $state<string>();
	private status = $state<'idle' | 'loading' | 'error'>('idle');
	private errorMessage = $state<string>();
	private hydrated = $state(false);
	isLoading = $derived(this.status === 'loading');
	selectedFeed: CustomFeedDefinition | undefined;
	private constructor() {
		this.selectedFeed = $derived.by(() => this.deriveSelectedFeed());
	}

	static getInstance(): CustomFeedStore {
		if (!CustomFeedStore.instance) {
			CustomFeedStore.instance = new CustomFeedStore();
		}
		return CustomFeedStore.instance;
	}

	get allDefinitions(): Map<string, CustomFeedDefinition> {
		return this.definitions;
	}

	private deriveSelectedFeed() {
		if (!this.selectedFeedId) return void 0;
		return this.definitions.get(this.selectedFeedId);
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

	/**
	 * Load all custom feed definitions from storage.
	 *
	 * Should be called once on app initialization to hydrate the store.
	 */
	async hydrate(): Promise<void> {
		if (this.hydrated) {
			return;
		}

		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			const feeds = await customFeedStorage.loadAll();
			this.definitions = feeds;
			this.hydrated = true;
			this.status = 'idle';
			console.log(`[custom-feed-store] Loaded ${feeds.size} custom feed(s)`);
		} catch (error) {
			console.error('[custom-feed-store] hydrate failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to load custom feeds';
			this.status = 'error';
		}
	}

	/**
	 * Create a new custom feed definition.
	 *
	 * Generates a unique ID and timestamps, then persists to storage.
	 */
	async create(feed: Omit<CustomFeedDefinition, 'id' | 'createdAt' | 'updatedAt'>): Promise<string> {
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			const id = `custom-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
			const now = new Date().toISOString();

			const definition: CustomFeedDefinition = { ...feed, id, createdAt: now, updatedAt: now };

			await customFeedStorage.save(definition);
			this.definitions.set(id, definition);
			this.status = 'idle';

			console.log(`[custom-feed-store] Created feed: ${definition.name} (${id})`);
			return id;
		} catch (error) {
			console.error('[custom-feed-store] create failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to create feed';
			this.status = 'error';
			throw error;
		}
	}

	/**
	 * Update an existing custom feed definition.
	 */
	async update(feedId: string, updates: Partial<Omit<CustomFeedDefinition, 'id' | 'createdAt'>>): Promise<void> {
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			const existing = this.definitions.get(feedId);
			if (!existing) {
				throw new Error(`Feed not found: ${feedId}`);
			}

			const updated: CustomFeedDefinition = {
				...existing,
				...updates,
				id: feedId,
				createdAt: existing.createdAt,
				updatedAt: new Date().toISOString()
			};

			await customFeedStorage.save(updated);
			this.definitions.set(feedId, updated);
			this.status = 'idle';

			console.log(`[custom-feed-store] Updated feed: ${updated.name} (${feedId})`);
		} catch (error) {
			console.error('[custom-feed-store] update failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to update feed';
			this.status = 'error';
			throw error;
		}
	}

	/**
	 * Delete a custom feed definition.
	 */
	async delete(feedId: string): Promise<void> {
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			await customFeedStorage.delete(feedId);
			this.definitions.delete(feedId);

			if (this.selectedFeedId === feedId) {
				this.selectedFeedId = undefined;
			}

			this.status = 'idle';
			console.log(`[custom-feed-store] Deleted feed: ${feedId}`);
		} catch (error) {
			console.error('[custom-feed-store] delete failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to delete feed';
			this.status = 'error';
			throw error;
		}
	}

	/**
	 * Clone an existing feed definition with a new name.
	 */
	async clone(feedId: string, newName: string): Promise<string> {
		const existing = this.definitions.get(feedId);
		if (!existing) {
			throw new Error(`Feed not found: ${feedId}`);
		}

		return this.create({
			name: newName,
			description: existing.description,
			sources: existing.sources,
			authorFilter: existing.authorFilter,
			rateBasedRule: existing.rateBasedRule,
			labelFilter: existing.labelFilter,
			keywordFilter: existing.keywordFilter
		});
	}

	/**
	 * Select a feed to display.
	 */
	select(feedId: string): void {
		if (!this.definitions.has(feedId)) {
			console.warn(`[custom-feed-store] Feed not found: ${feedId}`);
			return;
		}

		this.selectedFeedId = feedId;
	}

	/**
	 * Clear all custom feed definitions.
	 */
	async clearAll(): Promise<void> {
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			await customFeedStorage.clearAll();
			this.definitions.clear();
			this.selectedFeedId = undefined;
			this.status = 'idle';
			console.log('[custom-feed-store] Cleared all feeds');
		} catch (error) {
			console.error('[custom-feed-store] clearAll failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to clear feeds';
			this.status = 'error';
			throw error;
		}
	}

	/**
	 * Reset the store to initial state.
	 */
	reset(): void {
		this.definitions.clear();
		this.selectedFeedId = undefined;
		this.status = 'idle';
		this.errorMessage = undefined;
		this.hydrated = false;
	}
}

export const customFeedStore = CustomFeedStore.getInstance();
