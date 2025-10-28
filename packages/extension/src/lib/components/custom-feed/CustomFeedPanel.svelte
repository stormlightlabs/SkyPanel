<script lang="ts">
  import { customFeedStore } from "$lib/state/custom-feed.svelte";
  import { toastStore } from "$lib/state/toast.svelte";
  import { onMount } from "svelte";

  // TODO: Add feed editing capability
  // TODO: Add feed preview/details view
  // TODO: Add drag-and-drop reordering
  // TODO: Implement actual feed execution (fetching and filtering posts)
  const definitions = $derived(customFeedStore.allDefinitions);
  const status = $derived(customFeedStore.currentStatus);
  const error = $derived(customFeedStore.error);

  onMount(() => {
    customFeedStore.hydrate();
  });

  async function handleDelete(feedId: string, feedName: string) {
    if (!confirm(`Delete feed "${feedName}"?`)) {
      return;
    }

    try {
      await customFeedStore.delete(feedId);
      toastStore.success(`Deleted feed: ${feedName}`);
    } catch (error_) {
      console.error("[CustomFeedPanel] Delete failed", error_);
      toastStore.error("Failed to delete feed");
    }
  }

  async function handleClone(feedId: string, feedName: string) {
    const newName = prompt(`Clone "${feedName}" as:`, `${feedName} (copy)`);
    if (!newName) return;

    try {
      await customFeedStore.clone(feedId, newName);
      toastStore.success(`Cloned feed: ${newName}`);
    } catch (error_) {
      console.error("[CustomFeedPanel] Clone failed", error_);
      toastStore.error("Failed to clone feed");
    }
  }

  /**
   * Panel for managing custom feed definitions.
   *
   * Displays list of saved custom feeds with CRUD operations:
   * - View all saved feeds
   * - Delete feeds
   * - Clone feeds (creates duplicate with new name)
   */
</script>

<div class="space-y-4">
  <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-4 shadow-lg shadow-slate-950/40">
    <div class="flex items-center justify-between">
      <h2 class="text-sm font-semibold uppercase tracking-wide text-slate-400">Custom Feeds</h2>
      <span class="text-xs text-slate-500">{definitions.size} feed(s)</span>
    </div>

    <p class="mt-2 text-xs text-slate-400">
      Manage your privately stored feed definitions. Use the Search panel to create new feeds.
    </p>
  </section>

  {#if error}
    <div class="rounded-lg border border-red-500/40 bg-red-950/40 px-4 py-3 text-xs text-red-200">
      {error}
    </div>
  {/if}

  {#if status === "loading"}
    <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
      Loading feeds...
    </div>
  {:else if definitions.size === 0}
    <div class="rounded-xl border border-slate-800/40 bg-slate-900/70 p-6 text-center text-sm text-slate-400">
      No custom feeds yet. Use the Search panel to create one!
    </div>
  {:else}
    <section class="space-y-2">
      {#each definitions.values() as feed (feed.id)}
        <div class="rounded-lg border border-slate-800/50 bg-slate-900/80 p-4">
          <div class="flex items-start justify-between">
            <div class="flex-1">
              <h3 class="text-sm font-semibold text-slate-100">{feed.name}</h3>
              {#if feed.description}
                <p class="mt-1 text-xs text-slate-400">{feed.description}</p>
              {/if}
              <p class="mt-2 text-xs text-slate-500">
                Created {new Date(feed.createdAt).toLocaleDateString()}
              </p>
            </div>

            <div class="flex gap-2">
              <button
                type="button"
                class="rounded-md bg-slate-800/50 px-2 py-1 text-xs text-slate-300 transition hover:bg-slate-700"
                onclick={() => handleClone(feed.id, feed.name)}>
                Clone
              </button>
              <button
                type="button"
                class="rounded-md bg-red-900/50 px-2 py-1 text-xs text-red-300 transition hover:bg-red-800/50"
                onclick={() => handleDelete(feed.id, feed.name)}>
                Delete
              </button>
            </div>
          </div>
        </div>
      {/each}
    </section>
  {/if}

  <div class="rounded-lg border border-amber-500/40 bg-amber-950/20 px-4 py-3 text-xs text-amber-200">
    <p class="font-semibold">Note:</p>
    <p class="mt-1">
      Feed execution is not yet implemented. These are saved definitions only. Future updates will add the ability to
      view and filter posts based on these feed configurations.
    </p>
  </div>
</div>
