import type { SessionManager } from "./session-manager";
import type {
  GetFollowsRequest,
  GetFollowsResult,
  GetFollowersRequest,
  GetFollowersResult,
  Mutual,
} from "$lib/types/graph";

/**
 * Default number of follows/followers to fetch per page.
 *
 * NOTE: This could be made user-configurable in settings.
 */
const DEFAULT_LIMIT = 100;

/**
 * Service for fetching social graph data from Bluesky via authenticated AtpAgent.
 *
 * Supports:
 * - Fetching follows (accounts the user follows)
 * - Fetching followers (accounts that follow the user)
 * - Computing mutual follows (reciprocal relationships)
 *
 * All operations support pagination for large follow lists.
 */
export class GraphService {
  constructor(private readonly sessions: SessionManager) {}

  /**
   * Fetch accounts that the specified actor follows.
   *
   * @param request - Actor identifier and pagination parameters
   * @returns List of follows with optional cursor for next page
   * @throws Error if not authenticated or API request fails
   */
  async getFollows(request: GetFollowsRequest): Promise<GetFollowsResult> {
    const agent = this.sessions.agent;
    if (!agent.hasSession) {
      throw new Error("Not authenticated - please log in to fetch follows");
    }

    try {
      const response = await agent.app.bsky.graph.getFollows({
        actor: request.actor,
        cursor: request.cursor,
        limit: request.limit ?? DEFAULT_LIMIT,
      });
      return { follows: response.data.follows, cursor: response.data.cursor };
    } catch (error) {
      console.error("[GraphService] Get follows failed", { request, error });
      throw error;
    }
  }

  /**
   * Fetch accounts that follow the specified actor.
   *
   * @param request - Actor identifier and pagination parameters
   * @returns List of followers with optional cursor for next page
   * @throws Error if not authenticated or API request fails
   */
  async getFollowers(request: GetFollowersRequest): Promise<GetFollowersResult> {
    const agent = this.sessions.agent;
    if (!agent.hasSession) {
      throw new Error("Not authenticated - please log in to fetch followers");
    }

    try {
      const response = await agent.app.bsky.graph.getFollowers({
        actor: request.actor,
        cursor: request.cursor,
        limit: request.limit ?? DEFAULT_LIMIT,
      });
      return { followers: response.data.followers, cursor: response.data.cursor };
    } catch (error) {
      console.error("[GraphService] Get followers failed", { request, error });
      throw error;
    }
  }

  /**
   * Compute mutual follows for the specified actor.
   *
   * A mutual is defined as an account where:
   * - The actor follows them, AND
   * - They follow the actor back
   *
   * This method fetches all follows and followers (handling pagination automatically),
   * then computes the intersection.
   *
   * NOTE: This can be expensive for accounts with large follow counts.
   * Consider showing a progress indicator in the UI during computation.
   *
   * @param actor - Actor identifier (handle or DID)
   * @returns Array of mutual follow relationships
   * @throws Error if not authenticated or API request fails
   */
  async computeMutuals(actor: string): Promise<Mutual[]> {
    const agent = this.sessions.agent;
    if (!agent.hasSession) {
      throw new Error("Not authenticated - please log in to compute mutuals");
    }

    try {
      const [allFollows, allFollowers] = await Promise.all([
        this.fetchAllFollows(actor),
        this.fetchAllFollowers(actor),
      ]);

      const followerDids = new Set(allFollowers.map((f) => f.did));

      const mutuals: Mutual[] = allFollows
        .filter((follow) => followerDids.has(follow.did))
        .map((follow) => ({
          did: follow.did,
          handle: follow.handle,
          displayName: follow.displayName,
          avatar: follow.avatar,
          viewer: follow.viewer,
        }));

      console.log(`[GraphService] Computed ${mutuals.length} mutuals for ${actor}`);
      return mutuals;
    } catch (error) {
      console.error("[GraphService] Compute mutuals failed", { actor, error });
      throw error;
    }
  }

  /**
   * Fetch all follows for an actor, handling pagination automatically.
   *
   * NOTE: This can make multiple API calls for accounts with many follows.
   * Consider rate limit implications.
   */
  private async fetchAllFollows(actor: string) {
    const allFollows: GetFollowsResult["follows"] = [];
    let cursor: string | undefined;

    do {
      const result = await this.getFollows({ actor, cursor, limit: DEFAULT_LIMIT });
      allFollows.push(...result.follows);
      cursor = result.cursor;
    } while (cursor);

    return allFollows;
  }

  /**
   * Fetch all followers for an actor, handling pagination automatically.
   *
   * NOTE: This can make multiple API calls for accounts with many followers.
   * TODO: Consider rate limit implications.
   */
  private async fetchAllFollowers(actor: string) {
    const allFollowers: GetFollowersResult["followers"] = [];
    let cursor: string | undefined;

    do {
      const result = await this.getFollowers({ actor, cursor, limit: DEFAULT_LIMIT });
      allFollowers.push(...result.followers);
      cursor = result.cursor;
    } while (cursor);

    return allFollowers;
  }
}
