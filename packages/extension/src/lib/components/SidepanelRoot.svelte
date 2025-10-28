<script lang="ts">
  import { computedFeedStore } from "$lib/state/computed-feed.svelte";
  import { feedStore } from "$lib/state/feed.svelte";
  import { sessionStore } from "$lib/state/session.svelte";
  import { onMount, untrack } from "svelte";
  import FeedSelector from "./feed/FeedSelector.svelte";
  import AuthCard from "./session/AuthCard.svelte";

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
  <div class="flex flex-col gap-5 p-4">
    <header class="rounded-2xl border border-slate-800/60 bg-slate-900/70 p-4 shadow-xl shadow-slate-950/40">
      <div class="flex items-center justify-between">
        <div>
          <p class="text-xs uppercase tracking-[0.3em] text-sky-300">SkyPanel</p>
          <h1 class="mt-1 text-xl font-semibold text-slate-100">Bluesky Feeds at a Glance</h1>
          <p class="mt-1 text-xs text-slate-400">
            Sign in with an app password, load your home timeline, and pivot into author or list feeds without leaving
            the current page.
          </p>
        </div>
      </div>
    </header>

    <AuthCard />

    {#if hydrated && authenticated}
      <FeedSelector />
    {:else}
      <div class="rounded-xl border border-slate-800/50 bg-slate-900/60 p-6 text-sm text-slate-400">
        {#if !hydrated}
          Checking for saved session…
        {:else}
          Sign in above to load your feeds.
        {/if}
      </div>
    {/if}
  </div>
</div>
