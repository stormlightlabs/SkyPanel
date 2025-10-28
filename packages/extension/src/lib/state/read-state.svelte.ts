import { readStateStorage, type ReadStateMap } from '$lib/storage/read-state-storage';

class ReadStateStore {
	private readStates = $state<ReadStateMap>({});
	private initialized = $state(false);

	/**
	 * Initialize the store by loading persisted read states, called once on app startup.
	 */
	async init(): Promise<void> {
		if (this.initialized) {
			return;
		}
		this.readStates = await readStateStorage.load();
		this.initialized = true;
	}

	/**
	 * Get the last-seen timestamp for an author.
	 */
	getLastSeen(authorDid: string): number | undefined {
		return this.readStates[authorDid];
	}

	/**
	 * Check if a post is unread based on its author and timestamp.
	 */
	isUnread(authorDid: string, postTimestamp: string | number): boolean {
		const lastSeen = this.readStates[authorDid];
		if (lastSeen === undefined) {
			return false;
		}

		const postTime = typeof postTimestamp === 'string' ? new Date(postTimestamp).getTime() : postTimestamp;

		return postTime > lastSeen;
	}

	/**
	 * Mark an author as read at the given timestamp. If no timestamp provided, uses current time.
	 * This marks all posts from this author up to the timestamp as read.
	 */
	async markAuthorRead(authorDid: string, timestamp?: number): Promise<void> {
		const ts = timestamp ?? Date.now();
		this.readStates[authorDid] = ts;
		await readStateStorage.markSeen(authorDid, ts);
	}

	/**
	 * Mark multiple authors as read at once.
	 *
	 * More efficient than calling {@link markAuthorRead} multiple times.
	 */
	async bulkMarkRead(updates: Array<{ authorDid: string; timestamp: number }>): Promise<void> {
		for (const { authorDid, timestamp } of updates) {
			this.readStates[authorDid] = timestamp;
		}
		await readStateStorage.bulkUpdate(updates);
	}

	/**
	 * Get all read states.
	 */
	getAllStates(): ReadStateMap {
		return { ...this.readStates };
	}

	/**
	 * Clear all read states.
	 */
	async clearAll(): Promise<void> {
		this.readStates = {};
		await readStateStorage.clear();
	}
}

export const readStateStore = new ReadStateStore();
