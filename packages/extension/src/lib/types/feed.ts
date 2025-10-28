import type {
	AppBskyFeedDefs,
	AppBskyFeedGetAuthorFeed,
	AppBskyFeedGetListFeed,
	AppBskyFeedGetTimeline,
	AppBskyFeedGetPostThread
} from '@atproto/api';

export type FeedKind = 'timeline' | 'author' | 'list';
export type ThreadStatus = 'idle' | 'loading' | 'error';
export type TimelineFeedRequest = { kind: 'timeline'; cursor?: string; limit?: number };
export type AuthorFeedRequest = { kind: 'author'; actor: string; cursor?: string; limit?: number };
export type ListFeedRequest = { kind: 'list'; list: string; cursor?: string; limit?: number };
export type FeedRequest = TimelineFeedRequest | AuthorFeedRequest | ListFeedRequest;

export type FeedResponseData =
	| AppBskyFeedGetTimeline.Response['data']
	| AppBskyFeedGetAuthorFeed.Response['data']
	| AppBskyFeedGetListFeed.Response['data'];

type FeedResultBase = { cursor?: string; feed: AppBskyFeedDefs.FeedViewPost[] };

export type FeedResult =
	| (FeedResultBase & { kind: 'timeline' })
	| (FeedResultBase & { kind: 'author'; actor: string })
	| (FeedResultBase & { kind: 'list'; list: string });

/**
 * Request to fetch a thread for a specific post.
 *
 * @param uri - AT-URI of the post to fetch thread for
 * @param depth - How many levels of replies to fetch (default: 6)
 * @param parentHeight - How many levels of parents to fetch (default: 80)
 */
export type ThreadRequest = { uri: string; depth?: number; parentHeight?: number };

/**
 * Result of fetching a thread.
 *
 * Contains the full thread structure with parent chain and nested replies.
 */
export type ThreadResult = { thread: AppBskyFeedGetPostThread.OutputSchema['thread'] };
