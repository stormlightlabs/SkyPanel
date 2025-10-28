<script lang="ts">
  import { isExternalEmbed, isImagesEmbed, isVideoEmbed, type PostEmbed } from "$lib/types/embed";
  import { formatDistanceToNow } from "$lib/utils/time";
  import type { AppBskyActorDefs, AppBskyFeedDefs } from "@atproto/api";
  import FeedImageEmbed from "./FeedImageEmbed.svelte";
  import FeedLinkCard from "./FeedLinkCard.svelte";
  import FeedVideoEmbed from "./FeedVideoEmbed.svelte";

  let { item }: { item: AppBskyFeedDefs.FeedViewPost } = $props();

  const post = $derived(item.post);
  const author = $derived(post?.author);
  const record = $derived(post?.record as { text?: string } | undefined);
  const text = $derived(typeof record?.text === "string" ? record.text : "");
  const indexedAt = $derived(post?.indexedAt ? new Date(post.indexedAt) : null);
  const reason = $derived(item.reason as { $type?: string; by?: AppBskyActorDefs.ProfileViewBasic } | undefined);
  const embed = $derived(post?.embed as PostEmbed | undefined);
  const isRepost = $derived(reason?.$type === "app.bsky.feed.defs#reasonRepost");
  const isTrending = $derived(reason?.$type === "app.bsky.feed.defs#reasonTrend");
  const reposter = $derived(isRepost ? reason?.by : undefined);
</script>

<article class="rounded-xl border border-slate-800/40 bg-slate-900/80 p-4 shadow-lg shadow-slate-950/30">
  {#if isRepost && reposter}
    <p class="text-xs uppercase tracking-wide text-sky-300/80">
      Reposted by
      {#if reposter.handle}
        <a
          href="https://bsky.app/profile/{reposter.handle}"
          target="_blank"
          rel="noopener noreferrer"
          class="hover:text-sky-200 hover:underline">
          {reposter.displayName || `@${reposter.handle}`}
        </a>
      {:else}
        {reposter.displayName || "Unknown"}
      {/if}
    </p>
  {:else if isTrending}
    <p class="text-xs uppercase tracking-wide text-sky-300/80">Trending</p>
  {/if}

  <header class="mt-1 flex items-start justify-between gap-2">
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

  <!-- TODO: Replace emojis with icons (egoist?) -->
  <footer class="mt-4 flex items-center gap-4 text-xs text-slate-500">
    <span>💬 {post?.replyCount ?? 0}</span>
    <span>🔁 {post?.repostCount ?? 0}</span>
    <span>❤️ {post?.likeCount ?? 0}</span>
  </footer>
</article>
