<script lang="ts">
  import { threadStore } from "$lib/state/thread.svelte";
  import type { AppBskyFeedDefs } from "@atproto/api";
  import ThreadPostCard from "./ThreadPostCard.svelte";

  let { uri }: { uri: string } = $props();

  const thread = $derived(threadStore.currentThread);
  const status = $derived(threadStore.currentStatus);
  const error = $derived(threadStore.error);

  $effect(() => {
    if (uri) {
      threadStore.load(uri);
    }

    return () => {
      threadStore.reset();
    };
  });

  /**
   * Check if thread node is a valid ThreadViewPost.
   *
   * Handles NotFoundPost and BlockedPost cases.
   */
  function isThreadViewPost(node: unknown): node is AppBskyFeedDefs.ThreadViewPost {
    return !!node && typeof node === "object" && "$type" in node && node.$type === "app.bsky.feed.defs#threadViewPost";
  }

  /**
   * Recursively render parent chain from root to target.
   *
   * Parents are stored as nested structure, so we need to traverse to root first.
   */
  function renderParents(
    node: AppBskyFeedDefs.ThreadViewPost,
    depth: number = 0,
  ): Array<{ post: AppBskyFeedDefs.PostView; depth: number }> {
    const parents: Array<{ post: AppBskyFeedDefs.PostView; depth: number }> = [];

    if (node.parent && isThreadViewPost(node.parent)) {
      parents.push(...renderParents(node.parent, depth + 1));
    }

    return parents;
  }

  /**
   * Recursively render replies with proper nesting.
   */
  function renderReplies(
    node: AppBskyFeedDefs.ThreadViewPost,
    depth: number = 0,
  ): Array<{ post: AppBskyFeedDefs.PostView; depth: number }> {
    const replies: Array<{ post: AppBskyFeedDefs.PostView; depth: number }> = [];

    if (node.replies && Array.isArray(node.replies)) {
      for (const reply of node.replies) {
        if (isThreadViewPost(reply)) {
          replies.push({ post: reply.post, depth }, ...renderReplies(reply, depth + 1));
        }
      }
    }

    return replies;
  }

  const parents = $derived(thread && isThreadViewPost(thread) ? renderParents(thread).toReversed() : []);
  const targetPost = $derived(thread && isThreadViewPost(thread) ? thread.post : undefined);
  const replies = $derived(thread && isThreadViewPost(thread) ? renderReplies(thread, 1) : []);

  /**
   * Thread view component displaying full conversation context.
   *
   * Shows a complete thread with:
   * - Parent chain (ancestors) from root to target
   * - Target post (highlighted)
   * - Nested replies (descendants) with indentation
   *
   * Handles loading states and errors. Thread data is managed by {@link threadStore}.
   */
</script>

<div class="space-y-4">
  {#if status === "loading"}
    <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
      Loading thread…
    </div>
  {:else if error}
    <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
      {error}
    </div>
  {:else if thread && isThreadViewPost(thread)}
    <div class="space-y-3">
      {#if parents.length > 0}
        <div class="space-y-2">
          <h3 class="text-xs font-semibold uppercase tracking-wide text-slate-400">Thread</h3>
          {#each parents as { post, depth } (post.cid)}
            <ThreadPostCard {post} {depth} />
          {/each}
        </div>
      {/if}

      {#if targetPost}
        <div class="space-y-2">
          {#if parents.length === 0}
            <h3 class="text-xs font-semibold uppercase tracking-wide text-slate-400">Post</h3>
          {/if}
          <ThreadPostCard post={targetPost} depth={parents.length} isTarget={true} />
        </div>
      {/if}

      {#if replies.length > 0}
        <div class="space-y-2">
          <h3 class="text-xs font-semibold uppercase tracking-wide text-slate-400">
            Replies ({replies.length})
          </h3>
          {#each replies as { post, depth } (post.cid)}
            <ThreadPostCard {post} {depth} />
          {/each}
        </div>
      {/if}
    </div>
  {:else}
    <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
      Thread not found
    </div>
  {/if}
</div>
