<script lang="ts">
  import { toastStore } from "$lib/state/toast.svelte";
  import type { VideoEmbed } from "$lib/types/embed";
  import { downloadVideo } from "$lib/utils/download";

  /**
   * Displays video embeds from Bluesky posts with playback controls.
   *
   * Videos are not autoplayed and include native browser controls for
   * play/pause, volume, and fullscreen. Provides a download button with
   * toast feedback for success/failure.
   */

  let { embed }: { embed: VideoEmbed } = $props();

  async function handleDownload() {
    const filename = `video-${Date.now()}.mp4`;
    const success = await downloadVideo(embed.playlist, filename);

    if (success) {
      toastStore.success("Video downloaded");
    } else {
      toastStore.error("Failed to download video");
    }
  }
</script>

<div class="group relative overflow-hidden rounded-lg border border-slate-800/50 bg-slate-950/30">
  <video controls class="w-full" poster={embed.thumbnail}>
    <source src={embed.playlist} type="application/x-mpegURL" />
    <track kind="captions" />
    Your browser does not support video playback.
  </video>

  <div class="mt-2 flex items-center justify-between px-3 pb-2">
    {#if embed.alt}
      <span class="text-xs text-slate-400">{embed.alt}</span>
    {:else}
      <span></span>
    {/if}
    <button
      type="button"
      class="rounded-lg bg-sky-600 px-3 py-1.5 text-xs font-medium text-white transition hover:bg-sky-500"
      onclick={handleDownload}>
      Download
    </button>
  </div>
</div>
