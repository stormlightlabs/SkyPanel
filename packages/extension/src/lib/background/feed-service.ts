import type { FeedResult, FeedRequest } from "$lib/types/feed";
import type { SessionManager } from "./session-manager";

/**
 * Default number of posts to fetch per page in feed requests.
 *
 * NOTE: This could be made user-configurable in settings.
 */
const DEFAULT_LIMIT = 30;

/**
 * Service for fetching Bluesky feeds via authenticated AtpAgent.
 *
 * Supports three feed types:
 * - Timeline: Reverse-chronological home feed (following + algorithm)
 * - Author: Posts from a specific user (by handle or DID)
 * - List: Posts from members of a curated list (by AT-URI)
 *
 * All feeds support cursor-based pagination for infinite scroll.
 */
export class FeedService {
  constructor(private readonly sessions: SessionManager) {}

  /**
   * Fetch a page of posts from the requested feed.
   *
   * @param request - Feed type and pagination parameters
   * @returns Feed result with posts and optional cursor for next page
   * @throws Error if not authenticated or API request fails
   */
  async fetch(request: FeedRequest): Promise<FeedResult> {
    const agent = this.sessions.agent;
    if (!agent.hasSession) {
      throw new Error("Not authenticated - please log in to fetch feeds");
    }

    try {
      switch (request.kind) {
        case "timeline": {
          const response = await agent.app.bsky.feed.getTimeline({
            cursor: request.cursor,
            limit: request.limit ?? DEFAULT_LIMIT,
          });
          return { kind: "timeline", cursor: response.data.cursor, feed: response.data.feed };
        }
        case "author": {
          const response = await agent.app.bsky.feed.getAuthorFeed({
            actor: request.actor,
            cursor: request.cursor,
            limit: request.limit ?? DEFAULT_LIMIT,
          });
          return { kind: "author", actor: request.actor, cursor: response.data.cursor, feed: response.data.feed };
        }
        case "list": {
          const response = await agent.app.bsky.feed.getListFeed({
            list: request.list,
            cursor: request.cursor,
            limit: request.limit ?? DEFAULT_LIMIT,
          });
          return { kind: "list", list: request.list, cursor: response.data.cursor, feed: response.data.feed };
        }
        default: {
          const exhaustive: never = request;
          throw new Error(`Unsupported feed request: ${(exhaustive as { kind: string }).kind}`);
        }
      }
    } catch (error) {
      console.error("[FeedService] Feed fetch failed", { request, error });
      throw error;
    }
  }
}

/**
 * Future feed types can be added here:
 * - "mutuals": Computed feed of mutual follows
 * - "quiet": Computed feed of low-volume posters
 * - "search": Search results feed
 * - "custom": User-defined feed with filters
 */
