<script lang="ts">
  import { backgroundClient } from "$lib/client/background-client";
  import { customFeedStore } from "$lib/state/custom-feed.svelte";
  import { toastStore } from "$lib/state/toast.svelte";
  import type { SearchRequest } from "$lib/types/search";
  import type { AppBskyFeedDefs } from "@atproto/api";

  // TODO: Add advanced filters (tags, domain, URL)
  // TODO: Add search result preview before saving
  // TODO: Add pagination for search results
  let query = $state("");
  let author = $state("");
  let lang = $state("");
  let since = $state("");
  let until = $state("");

  let posts = $state<AppBskyFeedDefs.PostView[]>([]);
  let _cursor = $state<string>();
  let hitsTotal = $state<number>();
  let loading = $state(false);
  let error = $state<string>();

  async function handleSearch() {
    if (!query.trim()) {
      toastStore.error("Please enter a search query");
      return;
    }

    loading = true;
    error = undefined;

    try {
      const request: SearchRequest = {
        query: query.trim(),
        author: author.trim() || undefined,
        lang: lang.trim() || undefined,
        since: since || undefined,
        until: until || undefined,
      };

      const response = await backgroundClient.searchPosts(request);

      if (response.ok) {
        posts = response.result.posts;
        _cursor = response.result.cursor;
        hitsTotal = response.result.hitsTotal;
      } else {
        error = response.error;
      }
    } catch (error_) {
      console.error("[SearchPanel] Search failed", error_);
      error = error_ instanceof Error ? error_.message : "Search failed";
    } finally {
      loading = false;
    }
  }

  async function handleSaveAsFeed() {
    if (!query.trim()) {
      toastStore.error("Please enter a search query first");
      return;
    }

    try {
      const feedName = `Search: ${query.slice(0, 30)}${query.length > 30 ? "..." : ""}`;

      await customFeedStore.create({
        name: feedName,
        description: `Saved search query: ${query}`,
        sources: [{ type: "timeline" }],
      });

      toastStore.success(`Saved as custom feed: ${feedName}`);
    } catch (error_) {
      console.error("[SearchPanel] Save failed", error_);
      toastStore.error("Failed to save search as feed");
    }
  }

  /**
   * Search panel for querying Bluesky posts with filters.
   *
   * Supports basic search with the ability to save search queries as custom feeds.
   * Includes filters for author, language, and date ranges.
   */
</script>

<div class="space-y-4">
  <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-4 shadow-lg shadow-slate-950/40">
    <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-400">Search Posts</h2>

    <form
      class="mt-4 space-y-3"
      onsubmit={(e) => {
        e.preventDefault();
        handleSearch();
      }}>
      <div>
        <label class="text-xs font-medium text-slate-300" for="search-query">Query</label>
        <input
          type="text"
          id="search-query"
          class="mt-1 w-full rounded-lg border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/40"
          bind:value={query}
          placeholder="Search for posts..."
          disabled={loading} />
      </div>

      <div class="grid grid-cols-2 gap-3">
        <div>
          <label class="text-xs font-medium text-slate-300" for="search-author">Author (optional)</label>
          <input
            type="text"
            id="search-author"
            class="mt-1 w-full rounded-lg border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/40"
            bind:value={author}
            placeholder="@handle or DID"
            disabled={loading} />
        </div>

        <div>
          <label class="text-xs font-medium text-slate-300" for="search-lang">Language (optional)</label>
          <input
            type="text"
            id="search-lang"
            class="mt-1 w-full rounded-lg border border-slate-800 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/40"
            bind:value={lang}
            placeholder="en, es, fr..."
            disabled={loading} />
        </div>
      </div>

      <div class="flex gap-2">
        <button
          type="submit"
          class="flex-1 rounded-lg bg-sky-600 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-white transition hover:bg-sky-500 disabled:cursor-not-allowed disabled:opacity-60"
          disabled={loading || !query.trim()}>
          {loading ? "Searching..." : "Search"}
        </button>

        <button
          type="button"
          class="rounded-lg border border-slate-700 bg-slate-800/50 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-slate-300 transition hover:bg-slate-700 disabled:cursor-not-allowed disabled:opacity-60"
          onclick={handleSaveAsFeed}
          disabled={!query.trim()}>
          Save as Feed
        </button>
      </div>
    </form>
  </section>

  {#if error}
    <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
      {error}
    </div>
  {/if}

  {#if hitsTotal !== undefined}
    <div class="text-xs text-slate-400">Found {hitsTotal} results</div>
  {/if}

  {#if posts.length > 0}
    <section class="space-y-3">
      {#each posts as post (post.cid)}
        <div class="rounded-xl border border-slate-800/40 bg-slate-900/80 p-4">
          <p class="text-sm font-semibold text-slate-100">
            {post.author.displayName || `@${post.author.handle}`}
          </p>
          {#if post.record && typeof post.record === "object" && "text" in post.record}
            <p class="mt-2 whitespace-pre-wrap text-sm text-slate-200">{post.record.text}</p>
          {/if}
        </div>
      {/each}
    </section>
  {/if}
</div>
