<script lang="ts">
  import { toastStore } from "$lib/state/toast.svelte";
  import type { ExternalEmbed } from "$lib/types/embed";
  import { copyToClipboard } from "$lib/utils/clipboard";

  /**
   * Displays external link embeds in a bookmark-style card.
   *
   * Shows the link title, description, thumbnail image, and domain. Provides
   * a clipboard copy button to copy the URL with toast feedback. The entire
   * card is clickable to open the link in a new tab.
   */

  let { embed }: { embed: ExternalEmbed } = $props();

  async function handleCopyUrl(event: MouseEvent) {
    event.preventDefault();
    event.stopPropagation();

    const success = await copyToClipboard(embed.external.uri);
    if (success) {
      toastStore.success("URL copied to clipboard");
    } else {
      toastStore.error("Failed to copy URL");
    }
  }

  function getDomain(url: string): string {
    try {
      return new URL(url).hostname;
    } catch {
      return url;
    }
  }
</script>

<a
  href={embed.external.uri}
  target="_blank"
  rel="noopener noreferrer"
  class="group block overflow-hidden rounded-lg border border-slate-800/50 bg-slate-950/30 transition hover:border-slate-700">
  <div class="flex">
    {#if embed.external.thumb}
      <div class="w-32 shrink-0">
        <img src={embed.external.thumb} alt="" class="h-full w-full object-cover" />
      </div>
    {/if}
    <div class="flex flex-1 flex-col justify-between p-3">
      <div>
        <p class="text-xs font-medium text-slate-400">{getDomain(embed.external.uri)}</p>
        <p class="mt-1 text-sm font-semibold text-slate-200 line-clamp-2">{embed.external.title}</p>
        {#if embed.external.description}
          <p class="mt-1 text-xs text-slate-400 line-clamp-2">{embed.external.description}</p>
        {/if}
      </div>
      <button
        type="button"
        class="mt-2 flex items-center gap-1.5 self-start rounded-md bg-slate-800/50 px-2 py-1 text-xs text-slate-300 transition hover:bg-slate-700"
        onclick={handleCopyUrl}>
        <svg class="h-3.5 w-3.5" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path
            stroke-linecap="round"
            stroke-linejoin="round"
            stroke-width="2"
            d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        Copy URL
      </button>
    </div>
  </div>
</a>
