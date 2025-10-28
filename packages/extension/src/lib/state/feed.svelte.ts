import type { AppBskyFeedDefs } from '@atproto/api';
import { backgroundClient } from '$lib/client/background-client';
import type { FeedRequest } from '$lib/types/feed';
import { SvelteMap } from 'svelte/reactivity';
import { groupConsecutivePosts, type FeedItem } from '$lib/utils/post-grouping';
import { readStateStore } from '$lib/state/read-state.svelte';

type LoadingState = 'idle' | 'initial' | 'next';

/**
 * Manages feed state for the extension UI.
 *
 * Handles loading, pagination, and caching of feed posts from various sources (timeline, author feeds, list feeds).
 * Coordinates with BackgroundClient to fetch feed data and manages loading states and error handling.
 * Uses a Map internally to deduplicate posts by CID while preserving insertion order.
 */
class FeedStore {
	private static instance: FeedStore;

	private itemsMap = $state(new SvelteMap<string, AppBskyFeedDefs.FeedViewPost>());
	private activeRequest = $state<FeedRequest>({ kind: 'timeline' });
	private cursor = $state<string>();
	private loading = $state<LoadingState>('idle');
	private errorMessage = $state<string>();
	private inflight = false;
	isEmpty = $derived(this.loading === 'idle' && this.itemsMap.size === 0);
	hasMore = $derived(typeof this.cursor === 'string' && this.cursor.length > 0);

	/**
	 * Grouped feed items with consecutive posts by the same author collapsed.
	 *
	 * Uses read state to determine which groups should be auto-collapsed.
	 */
	groupedItems = $derived.by<FeedItem[]>(() => {
		const isUnreadFn = (authorDid: string, timestamp: string) => readStateStore.isUnread(authorDid, timestamp);
		const posts = [...this.itemsMap.values()];
		return groupConsecutivePosts(posts, isUnreadFn);
	});

	private constructor() {}

	static getInstance(): FeedStore {
		if (!FeedStore.instance) {
			FeedStore.instance = new FeedStore();
		}
		return FeedStore.instance;
	}

	get currentItems() {
		return this.itemsMap;
	}

	get currentCursor() {
		return this.cursor;
	}

	get currentLoading() {
		return this.loading;
	}

	get error() {
		return this.errorMessage;
	}

	get currentFeed() {
		return this.activeRequest;
	}

	async select(request: FeedRequest): Promise<void> {
		this.activeRequest = request;
		await this.fetch({ request, mode: 'replace' });
	}

	async reload(): Promise<void> {
		await this.fetch({ request: this.activeRequest, mode: 'replace' });
	}

	async loadMore(): Promise<void> {
		const nextCursor = this.cursor;
		if (!nextCursor || this.inflight) {
			return;
		}
		await this.fetch({ request: { ...this.activeRequest, cursor: nextCursor }, mode: 'append' });
	}

	reset(): void {
		this.itemsMap.clear();
		this.cursor = undefined;
		this.errorMessage = undefined;
		this.loading = 'idle';
	}

	private async fetch({ request, mode }: { request: FeedRequest; mode: 'replace' | 'append' }): Promise<void> {
		if (this.inflight) {
			return;
		}

		this.inflight = true;
		this.loading = mode === 'replace' ? 'initial' : 'next';
		this.errorMessage = undefined;

		try {
			const response = await backgroundClient.fetchFeed(request);
			if (!response.ok) {
				this.errorMessage = response.error;
				return;
			}

			const { result } = response;
			this.cursor = result.cursor;

			if (mode === 'replace') {
				this.itemsMap.clear();
				for (const item of result.feed) {
					this.itemsMap.set(item.post.cid, item);
				}
			} else {
				for (const item of result.feed) {
					this.itemsMap.set(item.post.cid, item);
				}
			}

			const { cursor: _ignoredCursor, ...rest } = request;
			this.activeRequest = rest as FeedRequest;
		} catch (error) {
			console.error('[feed-store] fetch failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to load feed';
		} finally {
			this.inflight = false;
			this.loading = 'idle';
		}
	}
}

export const feedStore = FeedStore.getInstance();
