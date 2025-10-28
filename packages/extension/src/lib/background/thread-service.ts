import type { AppBskyFeedGetPostThread } from '@atproto/api';
import type { SessionManager } from './session-manager';

/**
 * Default depth for fetching nested replies in a thread.
 *
 * Controls how many levels of replies to fetch below the target post.
 * Higher values load more data but may be slower.
 */
const DEFAULT_DEPTH = 6;

/**
 * Default parent height for fetching ancestor posts in a thread.
 *
 * Controls how many levels of parent posts to fetch above the target post.
 * Set to a high value to ensure we always get the full thread context.
 */
const DEFAULT_PARENT_HEIGHT = 80;

/**
 * Thread result type wrapping the AT Protocol response.
 */
export type ThreadResult = { thread: AppBskyFeedGetPostThread.OutputSchema['thread'] };

/**
 * Service for fetching post threads from Bluesky via authenticated AtpAgent.
 *
 * Supports fetching full thread context including:
 * - Parent posts (ancestors) up to the root
 * - The target post itself
 * - Reply posts (descendants) nested to specified depth
 *
 * Threads can include blocked or deleted posts represented as NotFoundPost or BlockedPost.
 */
export class ThreadService {
	constructor(private readonly sessions: SessionManager) {}

	/**
	 * Fetch a complete thread for the specified post.
	 *
	 * Uses the app.bsky.feed.getPostThread API to retrieve the full conversation context around a post, including parent chain and nested replies.
	 *
	 * @param uri - AT-URI of the post to fetch thread for (e.g., "at://did:plc:abc123/app.bsky.feed.post/xyz789")
	 * @param depth - How many levels of replies to fetch (default: 6)
	 * @param parentHeight - How many levels of parents to fetch (default: 80)
	 * @returns Thread result with nested post structure
	 * @throws Error if not authenticated or API request fails
	 */
	async fetchThread(uri: string, depth?: number, parentHeight?: number): Promise<ThreadResult> {
		const agent = this.sessions.agent;
		if (!agent.hasSession) {
			throw new Error('Not authenticated - please log in to fetch threads');
		}

		try {
			const response = await agent.app.bsky.feed.getPostThread({
				uri,
				depth: depth ?? DEFAULT_DEPTH,
				parentHeight: parentHeight ?? DEFAULT_PARENT_HEIGHT
			});

			return { thread: response.data.thread };
		} catch (error) {
			console.error('[ThreadService] Thread fetch failed', { uri, error });
			throw error;
		}
	}
}
