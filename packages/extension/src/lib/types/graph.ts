import type {
  AppBskyGraphDefs,
  AppBskyGraphGetFollows,
  AppBskyGraphGetFollowers,
  AppBskyActorDefs,
} from "@atproto/api";

/**
 * A follow relationship from the Bluesky social graph.
 *
 * Represents an account that a user follows.
 */
export type Follow = AppBskyGraphDefs.Relationship;

/**
 * A follower relationship from the Bluesky social graph.
 *
 * Represents an account that follows the user.
 */
export type Follower = AppBskyGraphDefs.Relationship;

/**
 * Request to fetch follows for an actor.
 */
export type GetFollowsRequest = { actor: string; cursor?: string; limit?: number };

/**
 * Request to fetch followers for an actor.
 */
export type GetFollowersRequest = { actor: string; cursor?: string; limit?: number };

/**
 * Response containing follows with pagination cursor.
 */
export type GetFollowsResult = { follows: AppBskyGraphGetFollows.Response["data"]["follows"]; cursor?: string };

/**
 * Response containing followers with pagination cursor.
 */
export type GetFollowersResult = { followers: AppBskyGraphGetFollowers.Response["data"]["followers"]; cursor?: string };

/**
 * A mutual follow relationship.
 *
 * Represents an account where both users follow each other.
 */
export type Mutual = {
  did: string;
  handle: string;
  displayName?: string;
  avatar?: string;
  viewer?: AppBskyActorDefs.ViewerState;
};
