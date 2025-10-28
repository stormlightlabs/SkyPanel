<script lang="ts">
  import {
    hydrateSession,
    isAuthenticated,
    login,
    logout,
    sessionError,
    sessionStatus,
    sessionStore,
  } from "$lib/state/session.svelte";
  import { onMount } from "svelte";

  let identifier = $state("");
  let password = $state("");
  let showPassword = $state(false);
  const handleId = $state("auth-handle");
  const passwordId = $state("auth-password");
  const authenticated = $derived.by(isAuthenticated);

  onMount(() => {
    hydrateSession();
  });

  const submit = async () => {
    if (!identifier || !password) return;
    await login(identifier.trim(), password);
    password = "";
  };
</script>

{#if authenticated}
  <div
    class="rounded-xl border border-slate-700 bg-slate-900/70 p-4 text-sm text-slate-200 shadow-lg shadow-slate-950/30">
    <div class="flex items-start gap-3">
      <div class="flex-1">
        <p class="text-xs uppercase tracking-[0.22em] text-slate-400">Signed in</p>
        <p class="mt-1 text-lg font-semibold text-sky-200">@{sessionStore?.handle}</p>
        <p class="mt-1 text-slate-400">Session backed by your Bluesky app password.</p>
      </div>
      <button
        class="rounded-full border border-slate-700 bg-slate-800 px-3 py-1 text-xs font-semibold uppercase tracking-wide text-slate-200 transition hover:border-slate-600 hover:text-sky-200"
        onclick={(e) => {
          e.preventDefault();
          logout();
        }}
        disabled={sessionStatus === "loading"}>
        {sessionStatus === "loading" ? "…" : "Sign out"}
      </button>
    </div>
  </div>
{:else}
  <form
    class="rounded-xl border border-slate-800 bg-slate-900/60 p-4 shadow-xl shadow-slate-950/40"
    onsubmit={(e) => {
      e.preventDefault();
      submit();
    }}>
    <div class="space-y-3">
      <div>
        <label class="text-xs font-semibold uppercase tracking-wide text-slate-400" for={handleId}>Handle</label>
        <input
          type="text"
          id={handleId}
          class="mt-1 w-full rounded-lg border border-slate-700 bg-slate-950 px-3 py-2 text-sm text-slate-100 outline-none transition focus:border-sky-400 focus:ring-2 focus:ring-sky-500/40"
          placeholder="alice.bsky.social"
          bind:value={identifier}
          required />
      </div>

      <div>
        <label class="text-xs font-semibold uppercase tracking-wide text-slate-400" for={passwordId}>
          App password</label>
        <div
          class="mt-1 flex rounded-lg border border-slate-700 bg-slate-950 transition focus-within:border-sky-400 focus-within:ring-2 focus-within:ring-sky-500/40">
          <input
            type={showPassword ? "text" : "password"}
            id={passwordId}
            class="w-full rounded-l-lg bg-transparent px-3 py-2 text-sm text-slate-100 outline-none"
            placeholder="abcd-1234-..."
            bind:value={password}
            required />
          <button
            type="button"
            class="rounded-r-lg border-l border-slate-800 px-3 text-xs font-semibold uppercase tracking-wide text-slate-400 transition hover:text-sky-200"
            onclick={(e) => (showPassword = !showPassword)}>
            {showPassword ? "Hide" : "Show"}
          </button>
        </div>
      </div>

      {#if sessionError}
        <p class="rounded-md border border-red-500/40 bg-red-950/40 px-3 py-2 text-xs text-red-200">
          {sessionError}
        </p>
      {/if}

      <button
        type="submit"
        class="w-full rounded-lg bg-sky-500 px-4 py-2 text-sm font-semibold uppercase tracking-wide text-slate-950 transition hover:bg-sky-400 disabled:cursor-not-allowed disabled:opacity-60"
        disabled={sessionStatus === "loading"}>
        {sessionStatus === "loading" ? "Signing in…" : "Sign in"}
      </button>

      <p class="text-xs text-slate-400">
        Use a Bluesky <span class="font-semibold text-sky-200">app password</span>.
        <span>Create one in Settings → Advanced → App passwords.</span>
      </p>
    </div>
  </form>
{/if}
