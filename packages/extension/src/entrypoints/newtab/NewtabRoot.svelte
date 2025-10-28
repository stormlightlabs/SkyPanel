<script lang="ts">
  import FeedSelector from "$lib/components/feed/FeedSelector.svelte";
  import AuthCard from "$lib/components/session/AuthCard.svelte";
  import { computedFeedStore } from "$lib/state/computed-feed.svelte";
  import { feedStore } from "$lib/state/feed.svelte";
  import { sessionStore } from "$lib/state/session.svelte";
  import { onMount, untrack } from "svelte";

  let bootstrapped = $state(false);
  const authenticated = $derived(sessionStore.isAuthenticated);
  const hydrated = $derived(sessionStore.isHydrated);

  onMount(() => {
    sessionStore.hydrate();
  });

  $effect(() => {
    if (authenticated && !bootstrapped) {
      untrack(() => {
        bootstrapped = true;
        feedStore.select({ kind: "timeline" });
      });
    }
  });

  $effect(() => {
    if (!authenticated && bootstrapped) {
      untrack(() => {
        bootstrapped = false;
      });
      feedStore.reset();
      computedFeedStore.reset();
    }
  });
</script>

<div class="min-h-screen bg-linear-to-b from-slate-950 via-slate-900 to-slate-950 text-slate-100">
  <div class="mx-auto max-w-7xl px-6 py-8">
    <header class="mb-8 rounded-2xl border border-slate-800/60 bg-slate-900/70 p-6 shadow-xl shadow-slate-950/40">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-sm uppercase tracking-[0.3em] text-sky-300">SkyPanel</p>
          <h1 class="mt-2 text-3xl font-semibold text-slate-100">Bluesky Feeds</h1>
          <p class="mt-2 text-sm text-slate-400">
            Full-width view with advanced feed controls. Sign in with an app password to access your home timeline,
            author feeds, list feeds, and computed default feeds (Mutuals, Quiet Posters).
          </p>
        </div>
      </div>
    </header>

    <div class="grid gap-6 lg:grid-cols-[1fr_minmax(400px,480px)]">
      <main class="space-y-6">
        {#if hydrated && authenticated}
          <FeedSelector />
        {:else}
          <div class="rounded-xl border border-slate-800/50 bg-slate-900/60 p-8 text-center text-sm text-slate-400">
            {#if !hydrated}
              Checking for saved session…
            {:else}
              Sign in to load your feeds.
            {/if}
          </div>
        {/if}
      </main>

      <aside class="space-y-6">
        <AuthCard />

        {#if hydrated && authenticated}
          <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/40">
            <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-300">Compose</h2>
            <p class="mt-3 text-xs text-slate-400">Compose and publish posts directly from SkyPanel.</p>
            <div
              class="mt-4 rounded-lg border border-slate-800/50 bg-slate-950/60 p-4 text-center text-xs text-slate-500">
              Coming soon
            </div>
          </section>

          <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/40">
            <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-300">Search</h2>
            <p class="mt-3 text-xs text-slate-400">
              Search posts with filters (author, hashtags, domains, date ranges).
            </p>
            <div
              class="mt-4 rounded-lg border border-slate-800/50 bg-slate-950/60 p-4 text-center text-xs text-slate-500">
              Coming soon
            </div>
          </section>

          <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/40">
            <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-300">Profile</h2>
            <p class="mt-3 text-xs text-slate-400">View your profile, stats, and followers.</p>
            <div
              class="mt-4 rounded-lg border border-slate-800/50 bg-slate-950/60 p-4 text-center text-xs text-slate-500">
              Coming soon
            </div>
          </section>

          <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-6 shadow-lg shadow-slate-950/40">
            <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-300">Settings</h2>
            <p class="mt-3 text-xs text-slate-400">
              Configure feed preferences, quiet poster thresholds, cache TTL, and more.
            </p>
            <div
              class="mt-4 rounded-lg border border-slate-800/50 bg-slate-950/60 p-4 text-center text-xs text-slate-500">
              Coming soon
            </div>
          </section>
        {/if}
      </aside>
    </div>
  </div>
</div>
