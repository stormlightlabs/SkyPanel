import type {
  AppBskyFeedDefs,
  AppBskyFeedGetAuthorFeed,
  AppBskyFeedGetListFeed,
  AppBskyFeedGetTimeline,
} from "@atproto/api";

export type FeedKind = "timeline" | "author" | "list";

export type TimelineFeedRequest = { kind: "timeline"; cursor?: string; limit?: number };
export type AuthorFeedRequest = { kind: "author"; actor: string; cursor?: string; limit?: number };
export type ListFeedRequest = { kind: "list"; list: string; cursor?: string; limit?: number };
export type FeedRequest = TimelineFeedRequest | AuthorFeedRequest | ListFeedRequest;

export type FeedResponseData =
  | AppBskyFeedGetTimeline.Response["data"]
  | AppBskyFeedGetAuthorFeed.Response["data"]
  | AppBskyFeedGetListFeed.Response["data"];

type FeedResultBase = { cursor?: string; feed: AppBskyFeedDefs.FeedViewPost[] };

export type FeedResult =
  | (FeedResultBase & { kind: "timeline" })
  | (FeedResultBase & { kind: "author"; actor: string })
  | (FeedResultBase & { kind: "list"; list: string });
