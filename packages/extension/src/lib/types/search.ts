/**
 * Type definitions for post search functionality.
 *
 * Provides types for searching posts via app.bsky.feed.searchPosts with
 * various filters including author, hashtags, domains, language, and date ranges.
 */

import type { AppBskyFeedDefs, AppBskyFeedSearchPosts } from '@atproto/api';

/**
 * Search request parameters matching Bluesky's searchPosts API.
 */
export type SearchRequest = {
	query: string;
	author?: string;
	lang?: string;
	domain?: string;
	url?: string;
	tag?: string[];
	since?: string;
	until?: string;
	cursor?: string;
	limit?: number;
};

/**
 * Search result from app.bsky.feed.searchPosts.
 */
export type SearchResult = { posts: AppBskyFeedDefs.PostView[]; cursor?: string; hitsTotal?: number };

/**
 * Type alias for the API response structure.
 */
export type SearchPostsResponse = AppBskyFeedSearchPosts.Response['data'];
