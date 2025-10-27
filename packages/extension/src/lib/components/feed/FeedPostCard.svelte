<script lang="ts">
  import { formatDistanceToNow } from "$lib/utils/time";
  import type { AppBskyActorDefs, AppBskyFeedDefs } from "@atproto/api";

  let { item }: { item: AppBskyFeedDefs.FeedViewPost } = $props();

  const post = $derived(item.post);
  const author = $derived(post?.author);
  const record = $derived(post?.record as { text?: string } | undefined);
  const text = $derived(typeof record?.text === "string" ? record.text : "");
  const indexedAt = $derived(post?.indexedAt ? new Date(post.indexedAt) : null);
  const reason = $derived(item.reason as { $type?: string; by?: AppBskyActorDefs.ProfileViewBasic } | undefined);

  const reasonLabel = $derived.by(() => {
    if (reason?.$type === "app.bsky.feed.defs#reasonRepost") {
      return `Reposted by ${reason?.by?.displayName ?? `@${reason?.by?.handle ?? "unknown"}`}`;
    } else if (reason?.$type === "app.bsky.feed.defs#reasonTrend") {
      return "Trending";
    }
    return null;
  });
</script>

<article class="rounded-xl border border-slate-800/40 bg-slate-900/80 p-4 shadow-lg shadow-slate-950/30">
  {#if reasonLabel}
    <p class="text-xs uppercase tracking-wide text-sky-300/80">{reasonLabel}</p>
  {/if}

  <header class="mt-1 flex items-start justify-between gap-2">
    <div>
      <p class="text-sm font-semibold text-slate-100">
        {#if author?.displayName}{author.displayName}{:else}@{author?.handle}{/if}
      </p>
      {#if author?.displayName}
        <p class="text-xs text-slate-400">@{author.handle}</p>
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

  <footer class="mt-4 flex items-center gap-4 text-xs text-slate-500">
    <span>💬 {post?.replyCount ?? 0}</span>
    <span>🔁 {post?.repostCount ?? 0}</span>
    <span>❤️ {post?.likeCount ?? 0}</span>
  </footer>
</article>
