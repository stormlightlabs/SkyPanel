import type { AppBskyFeedDefs } from "@atproto/api";
import { backgroundClient } from "$lib/client/background-client";
import type { FeedRequest } from "$lib/types/feed";

type LoadingState = "idle" | "initial" | "next";

let items = $state<AppBskyFeedDefs.FeedViewPost[]>([]);
let activeRequest = $state<FeedRequest>({ kind: "timeline" });
let cursor = $state<string | null>(null);
let loading = $state<LoadingState>("idle");
let errorMessage = $state<string | null>(null);
let inflight = false;

export const feedItems = items;
export const feedCursor = cursor;
export const feedLoading = loading;
export const feedError = errorMessage;
export const currentFeed = activeRequest;
export const feedEmpty = $derived(loading === "idle" && items.length === 0);
export const feedHasMore = $derived(typeof cursor === "string" && cursor.length > 0);

export async function selectFeed(request: FeedRequest): Promise<void> {
  activeRequest = request;
  await fetchFeed({ request, mode: "replace" });
}

export async function reloadActiveFeed(): Promise<void> {
  await fetchFeed({ request: activeRequest, mode: "replace" });
}

export async function loadMore(): Promise<void> {
  const nextCursor = cursor;
  if (!nextCursor || inflight) {
    return;
  }
  await fetchFeed({ request: { ...activeRequest, cursor: nextCursor }, mode: "append" });
}

export function resetFeed(): void {
  items = [];
  cursor = null;
  errorMessage = null;
  loading = "idle";
}

async function fetchFeed({ request, mode }: { request: FeedRequest; mode: "replace" | "append" }): Promise<void> {
  if (inflight) {
    return;
  }

  inflight = true;
  loading = mode === "replace" ? "initial" : "next";
  errorMessage = null;

  try {
    const response = await backgroundClient.fetchFeed(request);
    if (!response.ok) {
      errorMessage = response.error;
      return;
    }

    const { result } = response;
    cursor = result.cursor ?? null;

    if (mode === "replace") {
      items = result.feed;
    } else {
      items = [...items, ...result.feed];
    }

    const { cursor: _ignoredCursor, ...rest } = request;
    activeRequest = rest as FeedRequest;
  } catch (error) {
    console.error("[feed-store] fetch failed", error);
    errorMessage = error instanceof Error ? error.message : "Unable to load feed";
  } finally {
    inflight = false;
    loading = "idle";
  }
}
