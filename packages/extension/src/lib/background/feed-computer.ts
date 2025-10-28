import type { AppBskyFeedDefs } from "@atproto/api";
import type { SessionManager } from "./session-manager";
import type { GraphService } from "./graph-service";
import type { FeedService } from "./feed-service";
import type { ComputedFeedResult, QuietPoster } from "$lib/types/computed-feed";
import type { Mutual } from "$lib/types/graph";

/**
 * Number of posts to fetch per author when computing quiet posters.
 *
 * NOTE: This could be made user-configurable.
 */
const QUIET_POSTER_SAMPLE_SIZE = 30;

/**
 * Threshold for considering an account a "quiet poster" (posts per day).
 *
 * Accounts posting less frequently than this are included in the quiet feed.
 *
 * NOTE: This should be user-configurable in settings.
 */
const QUIET_POSTER_THRESHOLD = 1.0;

/**
 * Number of days to look back when computing post rate.
 *
 * NOTE: This could be made user-configurable.
 */
const QUIET_POSTER_LOOKBACK_DAYS = 30;

/**
 * Service for computing specialized feeds from Bluesky social graph and post data.
 *
 * Computes:
 * - Mutuals feed: Posts from accounts with reciprocal follows
 * - Quiet Posters feed: Posts from low-volume accounts to avoid missing sparse posters
 *
 * These computations are expensive and should be cached aggressively.
 */
export class FeedComputer {
  constructor(
    private readonly sessions: SessionManager,
    private readonly graphService: GraphService,
    private readonly feedService: FeedService,
  ) {}

  /**
   * Compute the mutuals feed for the current user.
   *
   * Process:
   * 1. Compute mutual follows (reciprocal relationships)
   * 2. Fetch timeline
   * 3. Filter to only posts from mutuals
   *
   * NOTE: This is expensive - the mutuals computation fetches all follows and followers.
   * Results should be cached with a reasonable TTL (e.g., 30 minutes).
   *
   * NOTE: Future enhancement - cursor pagination for mutuals feed.
   * Currently fetches full timeline and filters in memory.
   *
   * @param cursor - Pagination cursor (not currently used, for future enhancement)
   * @param limit - Number of posts to return (not currently used, returns all matches)
   * @returns Feed of posts from mutual follows
   * @throws Error if not authenticated or computation fails
   */
  async computeMutualsFeed(cursor?: string, _limit?: number): Promise<ComputedFeedResult> {
    const session = this.sessions.snapshot;
    if (!session) {
      throw new Error("Not authenticated - please log in to compute mutuals feed");
    }

    try {
      console.log("[FeedComputer] Computing mutuals feed...");

      const mutuals = await this.graphService.computeMutuals(session.did);
      const mutualDids = new Set(mutuals.map((m) => m.did));

      console.log(`[FeedComputer] Found ${mutuals.length} mutuals, fetching timeline...`);

      const timelineResult = await this.feedService.fetch({ kind: "timeline", cursor, limit: 100 });

      const filteredFeed = timelineResult.feed.filter((post) => {
        const authorDid = post.post.author.did;
        return mutualDids.has(authorDid);
      });

      console.log(
        `[FeedComputer] Filtered timeline to ${filteredFeed.length} posts from mutuals (out of ${timelineResult.feed.length} total)`,
      );

      return { kind: "mutuals", cursor: timelineResult.cursor, feed: filteredFeed, mutuals, computedAt: Date.now() };
    } catch (error) {
      console.error("[FeedComputer] Mutuals feed computation failed", error);
      throw error;
    }
  }

  /**
   * Compute the quiet posters feed for the current user.
   *
   * Process:
   * 1. Fetch all follows
   * 2. Sample recent posts from each follow
   * 3. Calculate post rate (posts per day)
   * 4. Identify accounts below the quiet threshold
   * 5. Fetch recent posts from quiet posters, prioritized by recency
   *
   * NOTE: This is extremely expensive - fetches author feeds for every follow.
   * Consider these optimizations:
   * - Cache post rate calculations per author with long TTL
   * - Only recompute for new follows
   * - Limit to sampling a subset of follows
   * - Show progress indicator in UI
   *
   * NOTE: User control opportunities:
   * - Configurable quiet threshold (posts per day)
   * - Configurable lookback window (days)
   * - Configurable sample size (posts per author)
   * - Option to exclude/include specific accounts
   *
   * @param cursor - Pagination cursor (not currently used, for future enhancement)
   * @param limit - Number of posts to return
   * @returns Feed of posts from quiet posters
   * @throws Error if not authenticated or computation fails
   */
  async computeQuietPostersFeed(_cursor?: string, limit = 50): Promise<ComputedFeedResult> {
    const session = this.sessions.snapshot;
    if (!session) {
      throw new Error("Not authenticated - please log in to compute quiet posters feed");
    }

    try {
      console.log("[FeedComputer] Computing quiet posters feed...");

      const followsResult = await this.graphService.getFollows({ actor: session.did, limit: 100 });

      console.log(`[FeedComputer] Found ${followsResult.follows.length} follows, computing post rates...`);

      const quietPosters = await this.identifyQuietPosters(followsResult.follows);

      console.log(
        `[FeedComputer] Identified ${quietPosters.length} quiet posters (threshold: ${QUIET_POSTER_THRESHOLD} posts/day)`,
      );

      const feed = await this.fetchQuietPostersPosts(quietPosters, limit);

      console.log(`[FeedComputer] Fetched ${feed.length} posts from quiet posters`);

      return { kind: "quiet", cursor: undefined, feed, quietPosters, computedAt: Date.now() };
    } catch (error) {
      console.error("[FeedComputer] Quiet posters feed computation failed", error);
      throw error;
    }
  }

  /**
   * Identify quiet posters from a list of follows by computing post rates.
   *
   * Samples recent posts for each follow and calculates posts per day.
   */
  private async identifyQuietPosters(
    follows: Array<{ did: string; handle: string; displayName?: string; avatar?: string }>,
  ): Promise<QuietPoster[]> {
    const quietPosters: QuietPoster[] = [];
    const lookbackMs = QUIET_POSTER_LOOKBACK_DAYS * 24 * 60 * 60 * 1000;
    const cutoffTime = Date.now() - lookbackMs;

    for (const follow of follows) {
      try {
        const authorFeed = await this.feedService.fetch({
          kind: "author",
          actor: follow.did,
          limit: QUIET_POSTER_SAMPLE_SIZE,
        });

        if (authorFeed.feed.length === 0) {
          quietPosters.push({
            did: follow.did,
            handle: follow.handle,
            displayName: follow.displayName,
            avatar: follow.avatar,
            postsPerDay: 0,
            lastPostAt: undefined,
          });
          continue;
        }

        const recentPosts = authorFeed.feed.filter((post) => {
          const indexedAt = new Date(post.post.indexedAt).getTime();
          return indexedAt >= cutoffTime;
        });

        const lastPostAt = new Date(authorFeed.feed[0].post.indexedAt).getTime();
        const postsPerDay = (recentPosts.length / QUIET_POSTER_LOOKBACK_DAYS) * 1.0;

        if (postsPerDay <= QUIET_POSTER_THRESHOLD) {
          quietPosters.push({
            did: follow.did,
            handle: follow.handle,
            displayName: follow.displayName,
            avatar: follow.avatar,
            postsPerDay,
            lastPostAt,
          });
        }
      } catch (error) {
        console.warn(`[FeedComputer] Failed to compute post rate for ${follow.handle}`, error);
      }
    }

    return quietPosters.sort((a, b) => (b.lastPostAt ?? 0) - (a.lastPostAt ?? 0));
  }

  /**
   * Fetch recent posts from quiet posters, prioritized by recency.
   *
   * NOTE: This could be optimized by fetching in parallel with a concurrency limit.
   */
  private async fetchQuietPostersPosts(
    quietPosters: QuietPoster[],
    limit: number,
  ): Promise<AppBskyFeedDefs.FeedViewPost[]> {
    const allPosts: AppBskyFeedDefs.FeedViewPost[] = [];

    for (const poster of quietPosters) {
      if (allPosts.length >= limit) {
        break;
      }

      try {
        const authorFeed = await this.feedService.fetch({ kind: "author", actor: poster.did, limit: 5 });

        allPosts.push(...authorFeed.feed);
      } catch (error) {
        console.warn(`[FeedComputer] Failed to fetch posts for ${poster.handle}`, error);
      }
    }

    return allPosts.slice(0, limit);
  }
}
