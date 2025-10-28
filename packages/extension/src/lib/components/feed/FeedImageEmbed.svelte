<script lang="ts">
  import { toastStore } from "$lib/state/toast.svelte";
  import type { ImagesEmbed } from "$lib/types/embed";
  import { downloadImage } from "$lib/utils/download";

  /**
   * Displays image embeds from Bluesky posts in a stacked list layout.
   *
   * Each image is shown with its alt text and a download button. Images are
   * displayed at their full width with aspect ratio preserved. Handles download
   * failures with user feedback via toast notifications.
   *
   * TODO: Add grid layout option for multiple images (2x2 for 4 images, 2x1 for 2)
   * TODO: Add carousel layout with prev/next navigation
   * TODO: Add layout switcher UI for users to toggle between stacked/grid/carousel
   */

  let { embed }: { embed: ImagesEmbed } = $props();

  async function handleDownload(url: string, alt: string, index: number) {
    const filename = `image-${index + 1}-${Date.now()}.jpg`;
    const success = await downloadImage(url, filename);

    if (success) {
      toastStore.success("Image downloaded");
    } else {
      toastStore.error("Failed to download image");
    }
  }
</script>

<div class="space-y-2">
  {#each embed.images as image, index (image.thumb)}
    <div class="group relative overflow-hidden rounded-lg border border-slate-800/50 bg-slate-950/30">
      <img src={image.fullsize} alt={image.alt} class="w-full" />

      <div
        class="absolute bottom-0 left-0 right-0 flex items-center justify-between bg-linear-to-t from-slate-950/90 to-transparent p-3 opacity-0 transition-opacity group-hover:opacity-100">
        {#if image.alt}
          <span class="text-xs text-slate-300">{image.alt}</span>
        {:else}
          <span></span>
        {/if}
        <button
          type="button"
          class="rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-sky-500"
          onclick={() => handleDownload(image.fullsize, image.alt, index)}>
          Download
        </button>
      </div>
    </div>
  {/each}
</div>
