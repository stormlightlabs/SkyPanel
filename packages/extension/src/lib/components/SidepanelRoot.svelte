<script lang="ts">
  import { resetFeed, selectFeed } from "$lib/state/feed.svelte";
  import { hydrateSession, isAuthenticated, sessionHydrated } from "$lib/state/session.svelte";
  import { onMount, untrack } from "svelte";
  import FeedPanel from "./feed/FeedPanel.svelte";
  import AuthCard from "./session/AuthCard.svelte";

  let bootstrapped = $state(false);
  const authenticated = $derived.by(isAuthenticated);
  onMount(() => {
    hydrateSession();
  });

  $effect(() => {
    if (authenticated && !bootstrapped) {
      untrack(() => {
        bootstrapped = true;
        selectFeed({ kind: "timeline" });
      });
    }
  });

  $effect(() => {
    if (!authenticated && bootstrapped) {
      untrack(() => {
        bootstrapped = false;
      });
      resetFeed();
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

    {#if sessionHydrated && authenticated}
      <FeedPanel />
    {:else}
      <div class="rounded-xl border border-slate-800/50 bg-slate-900/60 p-6 text-sm text-slate-400">
        {#if !sessionHydrated}
          Checking for saved session…
        {:else}
          Sign in above to load your feeds.
        {/if}
      </div>
    {/if}
  </div>
</div>
