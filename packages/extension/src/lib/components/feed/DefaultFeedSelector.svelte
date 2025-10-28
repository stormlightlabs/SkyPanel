<script lang="ts">
  import {
    computedFeedComputedAt,
    computedFeedError,
    computedFeedItems,
    computedFeedLoading,
    computedFeedMutuals,
    computedFeedQuietPosters,
    currentComputedFeed,
    getIsComputing,
    refreshActiveComputedFeed,
    selectComputedFeed,
  } from "$lib/state/computed-feed.svelte";
  import type { ComputedFeedKind } from "$lib/types/computed-feed";
  import { formatDistanceToNow } from "$lib/utils/time";
  import FeedPostCard from "./FeedPostCard.svelte";

  let { disabled = false } = $props();
  let selectedKind = $state<ComputedFeedKind>();
  const isComputing = $derived.by(getIsComputing);

  $effect(() => {
    if (currentComputedFeed) {
      selectedKind = currentComputedFeed.kind;
    }
  });

  const selectKind = (kind: ComputedFeedKind) => {
    if (disabled) return;
    selectedKind = kind;
    selectComputedFeed({ kind });
  };

  const refresh = () => {
    if (disabled || !selectedKind) return;
    refreshActiveComputedFeed();
  };

  const formatComputedAt = (timestamp?: number) => {
    if (!timestamp) return "Never";
    return formatDistanceToNow(timestamp);
  };

  const options = [
    { kind: "mutuals" as const, label: "Mutuals" },
    { kind: "quiet" as const, label: "Quiet Posters" },
  ];
</script>

<div class="space-y-4">
  <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-4 shadow-lg shadow-slate-950/40">
    <header class="flex items-center justify-between gap-2">
      <nav
        class="inline-flex rounded-full border border-slate-800 bg-slate-950/60 p-1 text-xs font-semibold uppercase tracking-wide text-slate-400">
        {#each options as option (option.kind)}
          <button
            type="button"
            class="rounded-full px-4 py-1 transition"
            class:bg-sky-600={selectedKind === option.kind}
            class:text-slate-950={selectedKind === option.kind}
            class:text-slate-300={selectedKind !== option.kind}
            class:hover:text-sky-200={selectedKind !== option.kind}
            {disabled}
            onclick={() => selectKind(option.kind)}>
            {option.label}
          </button>
        {/each}
      </nav>

      <button
        class="rounded-full border border-slate-800 bg-slate-900 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-300 transition hover:border-slate-700 hover:text-sky-200 disabled:cursor-not-allowed disabled:opacity-60"
        onclick={refresh}
        disabled={isComputing || !selectedKind || disabled}>
        {isComputing ? "…" : "Refresh"}
      </button>
    </header>

    {#if selectedKind === "mutuals"}
      <div class="mt-4 space-y-2">
        <p class="text-xs text-slate-400">
          Shows posts from accounts where you follow each other. This feed is computed from your social graph and cached
          for 30 minutes.
        </p>
        {#if computedFeedMutuals.length > 0}
          <p class="text-xs text-slate-500">
            Found {computedFeedMutuals.length} mutual{computedFeedMutuals.length === 1 ? "" : "s"}. Last computed:{" "}
            {formatComputedAt(computedFeedComputedAt)}.
          </p>
        {/if}
      </div>
    {:else if selectedKind === "quiet"}
      <div class="mt-4 space-y-2">
        <p class="text-xs text-slate-400">
          Shows posts from accounts that post infrequently (less than 1 post/day). Helps you avoid missing sparse
          posters. This feed is computed from post rates and cached for 2 hours.
        </p>
        {#if computedFeedQuietPosters.length > 0}
          <p class="text-xs text-slate-500">
            Found {computedFeedQuietPosters.length} quiet poster{computedFeedQuietPosters.length === 1 ? "" : "s"}. Last
            computed: {formatComputedAt(computedFeedComputedAt)}.
          </p>
        {/if}
      </div>
    {:else}
      <p class="mt-4 text-xs text-slate-400">
        Select a default feed to get started. These feeds are computed from your social graph and post data.
      </p>
    {/if}
  </section>

  {#if computedFeedError}
    <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
      {computedFeedError}
    </div>
  {/if}

  <section class="space-y-3">
    {#if !computedFeedItems.length && isComputing}
      <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
        {computedFeedLoading === "computing" ? "Computing feed…" : "Refreshing feed…"}
        <p class="mt-2 text-xs text-slate-500">This may take a moment for accounts with large follow lists.</p>
      </div>
    {:else if !computedFeedItems.length}
      <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
        {selectedKind ? "No posts found in this feed." : "Select a default feed to get started."}
      </div>
    {:else}
      {#each computedFeedItems as item (item.post?.cid ?? item.post?.uri ?? Math.random())}
        <FeedPostCard {item} />
      {/each}
    {/if}
  </section>
</div>
