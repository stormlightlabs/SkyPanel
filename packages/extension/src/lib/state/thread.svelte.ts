import { backgroundClient } from '$lib/client/background-client';
import type { ThreadStatus } from '$lib/types/feed';
import type { AppBskyFeedGetPostThread } from '@atproto/api';

/**
 * Manages thread state
 *
 * Provides reactive state management for post threads, including loading states and error handling.
 * Coordinates with background script via BackgroundClient to fetch thread data with parent chain and nested replies.
 *
 * Thread structure includes:
 * - Parent posts (ancestors) up to root
 * - Target post itself
 * - Reply posts (descendants) nested by depth
 */
class ThreadStore {
	private static instance: ThreadStore;

	private thread = $state<AppBskyFeedGetPostThread.OutputSchema['thread']>();
	private currentUri = $state<string>();
	private status = $state<ThreadStatus>('idle');
	private errorMessage = $state<string>();
	private inflight = false;

	isLoading = $derived(this.status === 'loading');

	private constructor() {}

	static getInstance(): ThreadStore {
		if (!ThreadStore.instance) {
			ThreadStore.instance = new ThreadStore();
		}
		return ThreadStore.instance;
	}

	get currentThread() {
		return this.thread;
	}

	get uri() {
		return this.currentUri;
	}

	get currentStatus() {
		return this.status;
	}

	get error() {
		return this.errorMessage;
	}

	/**
	 * Fetches a thread for the specified post URI.
	 *
	 * Loads the full thread context including parent chain (up to root) and nested replies (configurable depth).
	 *
	 * @param uri - AT-URI of the post to fetch thread for
	 * @param depth - How many levels of replies to fetch (default: 6)
	 * @param parentHeight - How many levels of parents to fetch (default: 80)
	 */
	async load(uri: string, depth?: number, parentHeight?: number): Promise<void> {
		if (this.inflight) {
			console.warn('[thread-store] load already in progress, skipping');
			return;
		}

		this.inflight = true;
		this.currentUri = uri;
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			const response = await backgroundClient.fetchThread({ uri, depth, parentHeight });

			if (response.ok) {
				this.thread = response.result.thread;
				this.status = 'idle';
			} else {
				this.errorMessage = response.error;
				this.status = 'error';
			}
		} catch (error) {
			console.error('[thread-store] load failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to load thread';
			this.status = 'error';
		} finally {
			this.inflight = false;
		}
	}

	/**
	 * Clears the current thread and resets state.
	 */
	reset(): void {
		this.thread = undefined;
		this.currentUri = undefined;
		this.status = 'idle';
		this.errorMessage = undefined;
		this.inflight = false;
	}
}

export const threadStore = ThreadStore.getInstance();
