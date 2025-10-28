import type { AppBskyFeedDefs } from "@atproto/api";
import type { Mutual } from "./graph";

/**
 * Types of computed feeds supported by SkyPanel.
 */
export type ComputedFeedKind = "mutuals" | "quiet";

/**
 * Request to fetch or compute the mutuals feed.
 *
 * Shows posts only from accounts where the follow is reciprocal.
 */
export type MutualsFeedRequest = { kind: "mutuals"; cursor?: string; limit?: number; forceRefresh?: boolean };

/**
 * Request to fetch or compute the quiet posters feed.
 *
 * Shows posts from accounts that post infrequently, to avoid missing sparse posters.
 *
 * NOTE: Future user control - threshold for "quiet" (e.g., posts per day)
 */
export type QuietPostersFeedRequest = { kind: "quiet"; cursor?: string; limit?: number; forceRefresh?: boolean };

export type ComputedFeedRequest = MutualsFeedRequest | QuietPostersFeedRequest;

type ComputedFeedResultBase = { cursor?: string; feed: AppBskyFeedDefs.FeedViewPost[]; computedAt: number };

export type ComputedFeedResult =
  | (ComputedFeedResultBase & { kind: "mutuals"; mutuals: Mutual[] })
  | (ComputedFeedResultBase & { kind: "quiet"; quietPosters: QuietPoster[] });

/**
 * Metadata about a quiet poster.
 *
 * Tracks post rate to identify low-volume accounts.
 */
export type QuietPoster = {
  did: string;
  handle: string;
  displayName?: string;
  avatar?: string;
  postsPerDay: number;
  lastPostAt?: number;
};

/**
 * Cached computed feed data with TTL.
 *
 * Stored in chrome.storage.local to avoid recomputing on every request.
 */
export type CachedComputedFeed = {
  kind: ComputedFeedKind;
  data: ComputedFeedResult;
  computedAt: number;
  expiresAt: number;
};

/**
 * Storage schema for all cached computed feeds.
 */
export type ComputedFeedCache = { mutuals?: CachedComputedFeed; quiet?: CachedComputedFeed };
