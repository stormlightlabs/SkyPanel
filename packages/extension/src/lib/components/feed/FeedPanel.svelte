<script lang="ts">
  import { feedStore } from "$lib/state/feed.svelte";
  import type { FeedKind } from "$lib/types/feed";
  import FeedPostCard from "./FeedPostCard.svelte";
  import { infiniteScroll } from "./infinite-scroll";

  let { disabled = false } = $props();
  let tab = $state<FeedKind>("timeline");
  let authorHandle = $state("");
  let listUri = $state("");

  const authorInputId = "feed-author";
  const listInputId = "feed-list";

  const items = $derived(feedStore.currentItems);
  const loading = $derived(feedStore.currentLoading);
  const error = $derived(feedStore.error);
  const hasMore = $derived(feedStore.hasMore);
  const current = $derived(feedStore.currentFeed);

  $effect(() => {
    if (current.kind === "author" && current.actor) {
      authorHandle = current.actor;
    } else if (current.kind === "list" && current.list) {
      listUri = current.list;
    }
  });

  const limitInput = (value: string) => value.trim();

  const selectTab = (kind: FeedKind) => {
    tab = kind;
    if (disabled) return;

    if (kind === "timeline") {
      feedStore.select({ kind: "timeline" });
    } else if (kind === "author" && authorHandle.trim()) {
      feedStore.select({ kind: "author", actor: authorHandle.trim() });
    } else if (kind === "list" && listUri.trim()) {
      feedStore.select({ kind: "list", list: listUri.trim() });
    }
  };

  const submitAuthor = () => {
    if (disabled) return;
    const value = limitInput(authorHandle);
    if (!value) return;
    feedStore.select({ kind: "author", actor: value });
  };

  const submitList = () => {
    if (disabled) return;
    const value = limitInput(listUri);
    if (!value) return;
    feedStore.select({ kind: "list", list: value });
  };

  const tryLoadMore = () => {
    if (disabled || !hasMore || loading !== "idle") return;
    feedStore.loadMore();
  };
</script>

<div class="space-y-4">
  <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-4 shadow-lg shadow-slate-950/40">
    <header class="flex items-center justify-between gap-2">
      <nav
        class="inline-flex rounded-full border border-slate-800 bg-slate-950/60 p-1 text-xs font-semibold uppercase tracking-wide text-slate-400">
        {#each ["timeline", "author", "list"] as kind (kind)}
          <button
            type="button"
            class="rounded-full px-4 py-1 transition"
            class:bg-sky-600={tab === kind}
            class:text-slate-950={tab === kind}
            class:text-slate-300={tab !== kind}
            class:hover:text-sky-200={tab !== kind}
            {disabled}
            onclick={() => selectTab(kind as FeedKind)}>
            {kind === "timeline" ? "Timeline" : kind === "author" ? "Author" : "List"}
          </button>
        {/each}
      </nav>

      <button
        class="rounded-full border border-slate-800 bg-slate-900 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-300 transition hover:border-slate-700 hover:text-sky-200 disabled:cursor-not-allowed disabled:opacity-60"
        onclick={() => feedStore.reload()}
        disabled={loading !== "idle" || disabled}>
        {loading !== "idle" ? "…" : "Refresh"}
      </button>
    </header>

    {#if tab === "author"}
      <form
        class="mt-4 space-y-2"
        onsubmit={(e) => {
          e.preventDefault();
          submitAuthor();
        }}>
        <label class="text-xs font-semibold uppercase tracking-wide text-slate-400" for={authorInputId}>
          Author handle or DID
        </label>
        <input
          type="text"
          id={authorInputId}
          class="w-full rounded-lg border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/40 disabled:cursor-not-allowed disabled:opacity-60"
          bind:value={authorHandle}
          placeholder="handle.bsky.social"
          {disabled} />
        <button
          type="submit"
          class="w-full rounded-lg bg-sky-500 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-slate-950 transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={disabled || !authorHandle.trim()}>
          Load feed
        </button>
      </form>
    {:else if tab === "list"}
      <form
        class="mt-4 space-y-2"
        onsubmit={(e) => {
          e.preventDefault();
          submitList();
        }}>
        <label class="text-xs font-semibold uppercase tracking-wide text-slate-400" for={listInputId}>
          List AT-URI
        </label>
        <input
          type="text"
          id={listInputId}
          class="w-full rounded-lg border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/40 disabled:cursor-not-allowed disabled:opacity-60"
          bind:value={listUri}
          placeholder="at://did:example:list/slug"
          {disabled} />
        <button
          type="submit"
          class="w-full rounded-lg bg-sky-500 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-slate-950 transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={disabled || !listUri.trim()}>
          Load feed
        </button>
      </form>
    {:else}
      <p class="mt-4 text-xs text-slate-400">
        Browse your home timeline or switch to an author or list feed. Use the controls above to load a new source.
      </p>
    {/if}
  </section>

  {#if error}
    <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
      {error}
    </div>
  {/if}

  <section class="space-y-3">
    {#if !items.size && loading !== "idle"}
      <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
        Loading feed…
      </div>
    {:else if !items.size}
      <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
        Select a feed source to get started.
      </div>
    {:else}
      {#each items.values() as item (item.post.cid)}
        <FeedPostCard {item} />
      {/each}
    {/if}

    {#if hasMore}
      <div use:infiniteScroll={{ onIntersect: tryLoadMore }} class="h-6 w-full"></div>

      <button
        class="w-full rounded-lg border border-slate-800 bg-slate-900 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-slate-200 transition hover:border-slate-700 hover:text-sky-200 disabled:cursor-not-allowed disabled:opacity-60"
        onclick={() => feedStore.loadMore()}
        disabled={loading !== "idle" || disabled}>
        {loading === "next" ? "Loading…" : "Load more"}
      </button>
    {/if}
  </section>
</div>
