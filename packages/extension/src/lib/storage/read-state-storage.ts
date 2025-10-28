/**
 * Storage layer for tracking read state of posts by author.
 *
 * Uses chrome.storage.local to persist global per-author last-seen timestamps.
 * Read state is author-scoped (not feed-scoped), so marking an author as read
 * applies across all feeds (timeline, lists, author feeds).
 */

const STORAGE_KEY = 'read_state';

/**
 * Map of author DIDs to Unix timestamps (milliseconds).
 * Represents the last time a post from this author was marked as read.
 */
export type ReadStateMap = Record<string, number>;

export class ReadStateStorage {
	/**
	 * Load all read states from {@link chrome.storage.local}.
	 */
	async load(): Promise<ReadStateMap> {
		const result = await chrome.storage.local.get(STORAGE_KEY);
		return (result[STORAGE_KEY] as ReadStateMap) ?? {};
	}

	/**
	 * Get the last-seen timestamp for a specific author.
	 */
	async getLastSeen(authorDid: string): Promise<number | undefined> {
		const readStates = await this.load();
		return readStates[authorDid];
	}

	/**
	 * Mark an author as read at the given timestamp.
	 *
	 * If timestamp is not provided, uses current time.
	 */
	async markSeen(authorDid: string, timestamp?: number): Promise<void> {
		const readStates = await this.load();
		readStates[authorDid] = timestamp ?? Date.now();
		await chrome.storage.local.set({ [STORAGE_KEY]: readStates });
	}

	/**
	 * Bulk update multiple author read states.
	 */
	async bulkUpdate(updates: Array<{ authorDid: string; timestamp: number }>): Promise<void> {
		const readStates = await this.load();
		for (const { authorDid, timestamp } of updates) {
			readStates[authorDid] = timestamp;
		}
		await chrome.storage.local.set({ [STORAGE_KEY]: readStates });
	}

	/**
	 * Clear all read states
	 */
	async clear(): Promise<void> {
		await chrome.storage.local.remove(STORAGE_KEY);
	}
}

export const readStateStorage = new ReadStateStorage();
