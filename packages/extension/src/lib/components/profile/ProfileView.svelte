<script lang="ts">
  import FeedPostCard from "$lib/components/feed/FeedPostCard.svelte";
  import { infiniteScroll } from "$lib/components/feed/infinite-scroll";
  import { feedStore } from "$lib/state/feed.svelte";
  import { profileStore } from "$lib/state/profile.svelte";
  import { sessionStore } from "$lib/state/session.svelte";
  import { onMount } from "svelte";

  const profile = $derived(profileStore.currentProfile);
  const profileStatus = $derived(profileStore.currentStatus);
  const profileError = $derived(profileStore.error);
  const isRefreshing = $derived(profileStore.isRefreshing);
  const fetchedAt = $derived(profileStore.fetchedAt);

  const items = $derived(feedStore.currentItems);
  const loading = $derived(feedStore.currentLoading);
  const feedError = $derived(feedStore.error);
  const hasMore = $derived(feedStore.hasMore);

  const currentSession = $derived(sessionStore.currentSession);

  onMount(() => {
    profileStore.load();

    if (currentSession?.handle) {
      feedStore.select({ kind: "author", actor: currentSession.handle });
    }
  });

  const tryLoadMore = () => {
    if (!hasMore || loading !== "idle") return;
    feedStore.loadMore();
  };

  function formatCount(count: number | undefined): string {
    if (!count) return "0";
    if (count >= 1_000_000) return `${(count / 1_000_000).toFixed(1)}M`;
    if (count >= 1000) return `${(count / 1000).toFixed(1)}K`;
    return count.toString();
  }

  function formatTimeSince(timestamp: number | undefined): string {
    if (!timestamp) return "";
    const now = Date.now();
    const diffMs = now - timestamp;
    const diffMin = Math.floor(diffMs / 60_000);

    if (diffMin < 1) return "just now";
    if (diffMin < 60) return `${diffMin}m ago`;

    const diffHour = Math.floor(diffMin / 60);
    if (diffHour < 24) return `${diffHour}h ago`;

    const diffDay = Math.floor(diffHour / 24);
    return `${diffDay}d ago`;
  }

  const handleRefresh = async () => {
    await profileStore.refresh();
  };

  /**
   * Profile view component displaying user profile metadata and posts.
   *
   * Shows avatar, banner, bio, follower/following counts, and the user's posts feed.
   * Includes placeholder UI for edit and share actions.
   * Automatically loads the current user's profile on mount.
   */
</script>

<div class="space-y-4">
  {#if profileStatus === "loading"}
    <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
      Loading profile…
    </div>
  {:else if profileStatus === "refreshing"}
    <div class="rounded-xl border border-sky-800/40 bg-sky-900/20 p-6 text-center text-sm text-sky-300">
      Refreshing profile…
    </div>
  {:else if profileError}
    <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
      {profileError}
    </div>
  {:else if profile}
    <section
      class="overflow-hidden rounded-xl border border-slate-800/50 bg-slate-900/70 shadow-lg shadow-slate-950/40">
      {#if profile.banner}
        <div class="h-32 w-full bg-slate-800">
          <img src={profile.banner} alt="" class="h-full w-full object-cover" />
        </div>
      {:else}
        <div class="h-32 w-full bg-linear-to-br from-slate-800 to-slate-900"></div>
      {/if}

      <div class="relative px-4 pb-4">
        <div class="-mt-12 mb-3">
          {#if profile.avatar}
            <img
              src={profile.avatar}
              alt={profile.displayName || profile.handle}
              class="h-24 w-24 rounded-full border-4 border-slate-900 bg-slate-800" />
          {:else}
            <div class="h-24 w-24 rounded-full border-4 border-slate-900 bg-slate-800"></div>
          {/if}
        </div>

        <div class="space-y-3">
          <div>
            {#if profile.displayName}
              <h2 class="text-xl font-bold text-slate-100">{profile.displayName}</h2>
              <p class="text-sm text-slate-400">@{profile.handle}</p>
            {:else}
              <h2 class="text-xl font-bold text-slate-100">@{profile.handle}</h2>
            {/if}
          </div>

          {#if profile.description}
            <p class="whitespace-pre-wrap text-sm text-slate-200">{profile.description}</p>
          {/if}

          <div class="flex items-center gap-4 text-sm">
            <div>
              <span class="font-semibold text-slate-100">{formatCount(profile.followsCount)}</span>
              <span class="text-slate-400">Following</span>
            </div>
            <div>
              <span class="font-semibold text-slate-100">{formatCount(profile.followersCount)}</span>
              <span class="text-slate-400">Followers</span>
            </div>
            <div>
              <span class="font-semibold text-slate-100">{formatCount(profile.postsCount)}</span>
              <span class="text-slate-400">Posts</span>
            </div>
          </div>

          <div class="flex items-center gap-2">
            <button
              type="button"
              class="rounded-lg border border-slate-700 bg-slate-800/50 px-4 py-2 text-xs font-medium text-slate-300 opacity-50"
              disabled
              title="Coming soon">
              Edit Profile
            </button>
            <button
              type="button"
              class="rounded-lg border border-slate-700 bg-slate-800/50 px-4 py-2 text-xs font-medium text-slate-300 opacity-50"
              disabled
              title="Coming soon">
              Share Profile
            </button>
            <button
              type="button"
              class="rounded-lg border border-sky-700/50 bg-sky-900/30 px-4 py-2 text-xs font-medium text-sky-300 transition hover:border-sky-600 hover:bg-sky-900/50 disabled:cursor-not-allowed disabled:opacity-50"
              onclick={handleRefresh}
              disabled={isRefreshing}
              title="Refresh profile">
              {isRefreshing ? "Refreshing…" : "Refresh"}
            </button>
            {#if fetchedAt}
              <span class="text-xs text-slate-500" title={new Date(fetchedAt).toLocaleString()}>
                Updated {formatTimeSince(fetchedAt)}
              </span>
            {/if}
          </div>
        </div>
      </div>
    </section>

    <section class="space-y-3">
      <h3 class="text-sm font-semibold uppercase tracking-wide text-slate-400">Your Posts</h3>

      {#if feedError}
        <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
          {feedError}
        </div>
      {/if}

      {#if items.size === 0 && loading !== "idle"}
        <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
          Loading posts…
        </div>
      {:else if items.size === 0}
        <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
          No posts yet
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
          disabled={loading !== "idle"}>
          {loading === "next" ? "Loading…" : "Load more"}
        </button>
      {/if}
    </section>
  {/if}
</div>
