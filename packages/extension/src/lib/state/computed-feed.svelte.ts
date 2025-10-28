import type { AppBskyFeedDefs } from "@atproto/api";
import { backgroundClient, type BackgroundClient } from "$lib/client/background-client";
import type { ComputedFeedRequest, QuietPoster } from "$lib/types/computed-feed";
import type { Mutual } from "$lib/types/graph";

type LoadingState = "idle" | "computing" | "refreshing";

/**
 * Manages computed feed state for the extension UI.
 *
 * Handles loading and caching of expensive computed feeds (Mutuals, Quiet Posters).
 * Each feed type includes specific metadata like mutual lists or quiet poster statistics.
 * Coordinates with {@link BackgroundClient} to fetch computed data and manages loading states.
 */
class ComputedFeedStore {
  private static instance: ComputedFeedStore;

  private items = $state<AppBskyFeedDefs.FeedViewPost[]>([]);
  private activeRequest = $state<ComputedFeedRequest>();
  private loading = $state<LoadingState>("idle");
  private errorMessage = $state<string>();
  private mutuals = $state<Mutual[]>([]);
  private quietPosters = $state<QuietPoster[]>([]);
  private computedAt = $state<number>();
  private inflight = false;

  private constructor() {}

  static getInstance(): ComputedFeedStore {
    if (!ComputedFeedStore.instance) {
      ComputedFeedStore.instance = new ComputedFeedStore();
    }
    return ComputedFeedStore.instance;
  }

  get currentItems() {
    return this.items;
  }

  get currentLoading() {
    return this.loading;
  }

  get error() {
    return this.errorMessage;
  }

  get currentFeed() {
    return this.activeRequest;
  }

  get currentMutuals() {
    return this.mutuals;
  }

  get currentQuietPosters() {
    return this.quietPosters;
  }

  get currentComputedAt() {
    return this.computedAt;
  }

  get isEmpty() {
    return this.loading === "idle" && this.items.length === 0;
  }

  get isComputing() {
    return this.loading === "computing" || this.loading === "refreshing";
  }

  async select(request: ComputedFeedRequest): Promise<void> {
    this.activeRequest = request;
    await this.fetch({ request, forceRefresh: false });
  }

  async refresh(): Promise<void> {
    if (!this.activeRequest) {
      return;
    }
    await this.fetch({ request: this.activeRequest, forceRefresh: true });
  }

  reset(): void {
    this.items = [];
    this.activeRequest = undefined;
    this.errorMessage = undefined;
    this.loading = "idle";
    this.mutuals = [];
    this.quietPosters = [];
    this.computedAt = undefined;
  }

  private async fetch({
    request,
    forceRefresh,
  }: {
    request: ComputedFeedRequest;
    forceRefresh: boolean;
  }): Promise<void> {
    if (this.inflight) {
      return;
    }

    this.inflight = true;
    this.loading = forceRefresh ? "refreshing" : "computing";
    this.errorMessage = undefined;

    try {
      const response = await backgroundClient.fetchComputedFeed({ ...request, forceRefresh });

      if (!response.ok) {
        this.errorMessage = response.error;
        return;
      }

      const { result } = response;
      this.items = result.feed;
      this.computedAt = result.computedAt;

      switch (result.kind) {
        case "mutuals":
          this.mutuals = result.mutuals;
          this.quietPosters = [];
          break;
        case "quiet":
          this.quietPosters = result.quietPosters;
          this.mutuals = [];
          break;
      }

      this.activeRequest = { kind: result.kind, forceRefresh: false };
    } catch (error) {
      console.error("[computed-feed-store] fetch failed", error);
      this.errorMessage = error instanceof Error ? error.message : "Unable to compute feed";
    } finally {
      this.inflight = false;
      this.loading = "idle";
    }
  }
}

export const computedFeedStore = ComputedFeedStore.getInstance();
