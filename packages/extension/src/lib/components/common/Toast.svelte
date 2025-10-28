<script lang="ts">
  import { toastStore, type ToastType } from "$lib/state/toast.svelte";

  function getBgColor(kind: ToastType) {
    switch (kind) {
      case "success": {
        return "bg-green-900/90";
      }
      case "error": {
        return "bg-red-900/90";
      }
      case "info": {
        return "bg-blue-900/90";
      }
    }
  }

  function getTextColor(kind: ToastType) {
    switch (kind) {
      case "success": {
        return "text-green-100";
      }
      case "error": {
        return "text-red-100";
      }
      case "info": {
        return "text-blue-100";
      }
    }
  }

  function getBorderColor(kind: ToastType) {
    switch (kind) {
      case "success": {
        return "border-green-700";
      }
      case "error": {
        return "border-red-700";
      }
      case "info": {
        return "border-blue-700";
      }
    }
  }

  /**
   * Toast notification container component.
   *
   * Renders all active toasts from the toast store in a fixed position
   * at the bottom-right of the viewport. Each toast displays with appropriate
   * styling based on its type (success/error/info) and can be manually dismissed.
   */
</script>

<div class="fixed bottom-4 right-4 z-50 flex flex-col gap-2">
  {#each toastStore.all as toast (toast.id)}
    <div
      class="flex items-center gap-3 rounded-lg border px-4 py-3 shadow-lg transition-all duration-200 {getBgColor(
        toast.type,
      )} {getBorderColor(toast.type)} {getTextColor(toast.type)}">
      <span class="flex-1 text-sm font-medium">{toast.message}</span>
      <button
        type="button"
        class="text-current opacity-60 transition hover:opacity-100"
        onclick={() => toastStore.remove(toast.id)}
        aria-label={toast.type}>
        <svg class="h-4 w-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M6 18L18 6M6 6l12 12" />
        </svg>
      </button>
    </div>
  {/each}
</div>
