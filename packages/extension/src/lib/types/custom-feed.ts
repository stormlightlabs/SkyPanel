/**
 * Type definitions for custom/private feed configurations.
 *
 * Custom feeds are locally stored feed definitions that allow users to
 * create personalized feeds with specific sources, filters, and rules.
 * These feeds are private by design and never published to remote services.
 */

import type { AppBskyFeedDefs } from '@atproto/api';

/**
 * Source types for custom feeds.
 */
export type CustomFeedSource =
	| { type: 'timeline' }
	| { type: 'author'; actor: string }
	| { type: 'list'; list: string };

/**
 * Filter to include or exclude posts based on author.
 */
export type AuthorFilter = { actors: string[]; mode: 'include' | 'exclude' };

/**
 * Rate-based rules to filter posts based on posting frequency.
 */
export type RateBasedRule = { enabled: boolean; maxPostsPerDay?: number; minPostsPerDay?: number };

/**
 * Label-based filters for content moderation.
 */
export type LabelFilter = { labels: string[]; mode: 'include' | 'exclude' };

/**
 * Keyword/text filters for post content.
 */
export type KeywordFilter = { keywords: string[]; mode: 'include' | 'exclude'; caseSensitive: boolean };

/**
 * Complete custom feed definition schema.
 */
export type CustomFeedDefinition = {
	id: string;
	name: string;
	description?: string;
	createdAt: string;
	updatedAt: string;
	sources: CustomFeedSource[];
	authorFilter?: AuthorFilter;
	rateBasedRule?: RateBasedRule;
	labelFilter?: LabelFilter;
	keywordFilter?: KeywordFilter;
};

/**
 * Custom feed state for loading and display.
 */
export type CustomFeedState = {
	definitions: Map<string, CustomFeedDefinition>;
	selectedFeedId?: string;
	posts: Map<string, AppBskyFeedDefs.FeedViewPost>;
	cursor?: string;
	status: 'idle' | 'loading' | 'error';
	errorMessage?: string;
};

/**
 * Request to load a custom feed.
 */
export type CustomFeedRequest = { feedId: string; cursor?: string; limit?: number };

/**
 * Result from loading a custom feed.
 */
export type CustomFeedResult = { feedId: string; posts: AppBskyFeedDefs.FeedViewPost[]; cursor?: string };
