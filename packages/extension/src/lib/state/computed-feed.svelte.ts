import type { AppBskyFeedDefs } from "@atproto/api";
import { backgroundClient } from "$lib/client/background-client";
import type { ComputedFeedRequest, ComputedFeedResult, QuietPoster } from "$lib/types/computed-feed";
import type { Mutual } from "$lib/types/graph";

/**
 * State management for computed feeds (Mutuals, Quiet Posters).
 *
 * Computed feeds are expensive to generate and are cached aggressively.
 * Each feed type has specific metadata (list of mutuals, quiet poster stats, etc.).
 */

type LoadingState = "idle" | "computing" | "refreshing";

let items = $state<AppBskyFeedDefs.FeedViewPost[]>([]);
let activeRequest = $state<ComputedFeedRequest>();
let loading = $state<LoadingState>("idle");
let errorMessage = $state<string>();
let mutuals = $state<Mutual[]>([]);
let quietPosters = $state<QuietPoster[]>([]);
let computedAt = $state<number>();
let inflight = false;

export const computedFeedItems = items;
export const computedFeedLoading = loading;
export const computedFeedError = errorMessage;
export const currentComputedFeed = activeRequest;
export const computedFeedMutuals = mutuals;
export const computedFeedQuietPosters = quietPosters;
export const computedFeedComputedAt = computedAt;

export const getComputedFeedEmpty = () => loading === "idle" && items.length === 0;
export const getIsComputing = () => loading === "computing" || loading === "refreshing";

/**
 * Select and fetch a computed feed.
 *
 * Uses cached result if available, unless forceRefresh is true.
 *
 * @param request - Computed feed request (mutuals or quiet)
 */
export async function selectComputedFeed(request: ComputedFeedRequest): Promise<void> {
  activeRequest = request;
  await fetchComputedFeed({ request, forceRefresh: false });
}

/**
 * Refresh the currently active computed feed.
 *
 * Forces recomputation, bypassing cache.
 */
export async function refreshActiveComputedFeed(): Promise<void> {
  if (!activeRequest) {
    return;
  }
  await fetchComputedFeed({ request: activeRequest, forceRefresh: true });
}

/**
 * Reset computed feed state to initial values.
 */
export function resetComputedFeed(): void {
  items = [];
  activeRequest = undefined;
  errorMessage = undefined;
  loading = "idle";
  mutuals = [];
  quietPosters = [];
  computedAt = undefined;
}

/**
 * Fetch computed feed from background service.
 *
 * The background service handles caching and expensive computation.
 * This function just manages UI state.
 */
async function fetchComputedFeed({
  request,
  forceRefresh,
}: {
  request: ComputedFeedRequest;
  forceRefresh: boolean;
}): Promise<void> {
  if (inflight) {
    return;
  }

  inflight = true;
  loading = forceRefresh ? "refreshing" : "computing";
  errorMessage = undefined;

  try {
    const response = await backgroundClient.fetchComputedFeed({ ...request, forceRefresh });

    if (!response.ok) {
      errorMessage = response.error;
      return;
    }

    const { result } = response;
    items = result.feed;
    computedAt = result.computedAt;

    switch (result.kind) {
      case "mutuals":
        mutuals = result.mutuals;
        quietPosters = [];
        break;
      case "quiet":
        quietPosters = result.quietPosters;
        mutuals = [];
        break;
    }

    activeRequest = { kind: result.kind, forceRefresh: false };
  } catch (error) {
    console.error("[computed-feed-store] fetch failed", error);
    errorMessage = error instanceof Error ? error.message : "Unable to compute feed";
  } finally {
    inflight = false;
    loading = "idle";
  }
}
