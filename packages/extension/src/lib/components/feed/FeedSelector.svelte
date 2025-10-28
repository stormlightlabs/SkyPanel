<script lang="ts">
  import ProfileView from "../profile/ProfileView.svelte";
  import DefaultFeedSelector from "./DefaultFeedSelector.svelte";
  import FeedPanel from "./FeedPanel.svelte";

  let { disabled = false } = $props();
  let feedMode = $state<"standard" | "default" | "profile">("standard");

  const selectMode = (mode: "standard" | "default" | "profile") => {
    if (disabled) return;
    feedMode = mode;
  };

  const options = [
    { mode: "standard", label: "Standard Feeds" },
    { mode: "default", label: "Default Feeds" },
    { mode: "profile", label: "My Profile" },
  ] as const;
</script>

<div class="space-y-4">
  <section class="rounded-xl border border-slate-800/50 bg-slate-900/70 p-3 shadow-lg shadow-slate-950/40">
    <nav
      class="inline-flex w-full rounded-full border border-slate-800 bg-slate-950/60 p-1 text-xs font-semibold uppercase tracking-wide text-slate-400">
      {#each options as option (option.mode)}
        <button
          type="button"
          class="flex-1 rounded-full px-4 py-2 transition"
          class:bg-sky-600={feedMode === option.mode}
          class:text-slate-950={feedMode === option.mode}
          class:text-slate-300={feedMode !== option.mode}
          class:hover:text-sky-200={feedMode !== option.mode}
          {disabled}
          onclick={() => selectMode(option.mode)}>
          {option.label}
        </button>
      {/each}
    </nav>
  </section>

  {#if feedMode === "standard"}
    <FeedPanel {disabled} />
  {:else if feedMode === "default"}
    <DefaultFeedSelector {disabled} />
  {:else if feedMode === "profile"}
    <ProfileView />
  {/if}
</div>
