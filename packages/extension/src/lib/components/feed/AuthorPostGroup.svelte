<script lang="ts">
  import { readStateStore } from "$lib/state/read-state.svelte";
  import type { PostGroup } from "$lib/utils/post-grouping";
  import { formatDistanceToNow } from "$lib/utils/time";
  import { onMount } from "svelte";
  import FeedPostCard from "./FeedPostCard.svelte";

  let { group }: { group: PostGroup } = $props();

  let expanded = $state(false);
  let containerElement: HTMLElement;

  const author = $derived(group.author);
  const postCount = $derived(group.posts.length);
  const firstPost = $derived(group.posts[0]);
  const _lastPost = $derived(group.posts.at(-1));
  const firstPostTime = $derived(new Date(group.firstPostAt));
  const lastPostTime = $derived(new Date(group.lastPostAt));

  const previewText = $derived(() => {
    const record = firstPost.post.record as { text?: string } | undefined;
    const text = typeof record?.text === "string" ? record.text : "";
    return text.length > 100 ? text.slice(0, 100) + "..." : text;
  });

  function toggleExpand() {
    expanded = !expanded;

    if (expanded && group.isUnread) {
      readStateStore.markAuthorRead(author.did, firstPostTime.getTime());
    }
  }

  onMount(() => {
    if (!containerElement) return;

    const observer = new IntersectionObserver(
      (entries) => {
        const entry = entries[0];

        if (!entry.isIntersecting && group.isUnread) {
          setTimeout(() => {
            readStateStore.markAuthorRead(author.did, firstPostTime.getTime());
          }, 500);
        }
      },
      { threshold: 0, rootMargin: "0px" },
    );

    observer.observe(containerElement);

    return () => {
      observer.disconnect();
    };
  });
</script>

<div bind:this={containerElement}>
  {#if !expanded}
    <button
      onclick={toggleExpand}
      class="w-full rounded-xl border border-slate-800/40 bg-slate-900/80 p-4 text-left shadow-lg shadow-slate-950/30 transition-all hover:border-slate-700/60 hover:bg-slate-900">
      <div class="flex items-start justify-between gap-2">
        <div class="flex-1">
          <div class="flex items-center gap-2">
            <p class="text-sm font-semibold text-slate-100">
              {#if author.handle}
                <span class="hover:text-sky-400">
                  {author.displayName || `@${author.handle}`}
                </span>
              {:else}
                {author.displayName || "Unknown"}
              {/if}
            </p>
            {#if group.isUnread}
              <span class="rounded-full bg-sky-500/20 px-2 py-0.5 text-xs text-sky-400"> NEW </span>
            {/if}
          </div>

          {#if author.displayName && author.handle}
            <p class="text-xs text-slate-400">@{author.handle}</p>
          {/if}

          <p class="mt-2 text-xs text-slate-500">
            {postCount} posts · {formatDistanceToNow(firstPostTime)} - {formatDistanceToNow(lastPostTime)}
          </p>

          {#if previewText()}
            <p class="mt-2 text-sm text-slate-400">{previewText()}</p>
          {/if}
        </div>

        <div class="text-slate-500">
          <svg class="h-5 w-5 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </div>
    </button>
  {:else}
    <div class="space-y-4">
      <button
        onclick={toggleExpand}
        class="flex w-full items-center justify-between rounded-xl border border-slate-800/40 bg-slate-900/80 p-3 text-left shadow-lg shadow-slate-950/30 transition-all hover:border-slate-700/60 hover:bg-slate-900">
        <div>
          <p class="text-sm font-semibold text-slate-100">
            {author.displayName || `@${author.handle || "Unknown"}`}
          </p>
          <p class="text-xs text-slate-400">{postCount} posts</p>
        </div>

        <div class="text-slate-500">
          <svg class="h-5 w-5 rotate-180 transition-transform" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
          </svg>
        </div>
      </button>

      {#each group.posts as post (post.post.cid)}
        <FeedPostCard item={post} />
      {/each}
    </div>
  {/if}
</div>
