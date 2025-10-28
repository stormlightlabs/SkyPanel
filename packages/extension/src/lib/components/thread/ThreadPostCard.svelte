<script lang="ts">
  import FeedImageEmbed from "$lib/components/feed/FeedImageEmbed.svelte";
  import FeedLinkCard from "$lib/components/feed/FeedLinkCard.svelte";
  import FeedVideoEmbed from "$lib/components/feed/FeedVideoEmbed.svelte";
  import { isExternalEmbed, isImagesEmbed, isVideoEmbed, type PostEmbed } from "$lib/types/embed";
  import { formatDistanceToNow } from "$lib/utils/time";
  import type { AppBskyFeedDefs } from "@atproto/api";

  let {
    post,
    depth = 0,
    isTarget = false,
  }: { post: AppBskyFeedDefs.PostView; depth?: number; isTarget?: boolean } = $props();

  const author = $derived(post?.author);
  const record = $derived(post?.record as { text?: string } | undefined);
  const text = $derived(typeof record?.text === "string" ? record.text : "");
  const indexedAt = $derived(post?.indexedAt ? new Date(post.indexedAt) : null);
  const embed = $derived(post?.embed as PostEmbed | undefined);
  const indentClass = $derived(depth > 0 ? `ml-${Math.min(depth * 4, 12)}` : "");
  const borderClass = $derived(isTarget ? "border-sky-700/60 bg-sky-950/30" : "border-slate-800/40 bg-slate-900/80");
  /**
   * Thread post card component for displaying posts within a thread context.
   *
   * Renders a single post with indentation based on depth to show nesting.
   * Supports all the same embeds and metadata as FeedPostCard but adapted for threaded display.
   *
   * @param post - The post data to display
   * @param depth - Nesting depth for indentation (0 = root level)
   * @param isTarget - Whether this is the target post being viewed
   */
</script>

<article class="rounded-xl border p-4 shadow-lg shadow-slate-950/30 {borderClass} {indentClass}">
  <header class="flex items-start justify-between gap-2">
    <div>
      <p class="text-sm font-semibold text-slate-100">
        {#if author?.handle}
          <a
            href="https://bsky.app/profile/{author.handle}"
            target="_blank"
            rel="noopener noreferrer"
            class="hover:text-sky-400 hover:underline">
            {author.displayName || `@${author.handle}`}
          </a>
        {:else}
          {author?.displayName || "Unknown"}
        {/if}
      </p>
      {#if author?.displayName && author?.handle}
        <a
          href="https://bsky.app/profile/{author.handle}"
          target="_blank"
          rel="noopener noreferrer"
          class="text-xs text-slate-400 hover:text-sky-400 hover:underline">
          @{author.handle}
        </a>
      {/if}
    </div>
    {#if indexedAt}
      <time class="text-xs text-slate-500" datetime={indexedAt.toISOString()}>
        {formatDistanceToNow(indexedAt)}
      </time>
    {/if}
  </header>

  {#if text}
    <p class="mt-3 whitespace-pre-wrap text-sm text-slate-200">{text}</p>
  {/if}

  {#if embed}
    <div class="mt-3">
      {#if isImagesEmbed(embed)}
        <FeedImageEmbed {embed} />
      {:else if isVideoEmbed(embed)}
        <FeedVideoEmbed {embed} />
      {:else if isExternalEmbed(embed)}
        <FeedLinkCard {embed} />
      {/if}
    </div>
  {/if}

  <footer class="mt-4 flex items-center gap-4 text-xs text-slate-500">
    <span>💬 {post?.replyCount ?? 0}</span>
    <span>🔁 {post?.repostCount ?? 0}</span>
    <span>❤️ {post?.likeCount ?? 0}</span>
  </footer>
</article>
