import { browser } from "wxt/browser";
import { backgroundClient } from "$lib/client/background-client";
import { isSessionChangedEvent, type SessionChangedEvent } from "$lib/messaging/messages";
import type { SessionSnapshot } from "$lib/types/session";

let session = $state<SessionSnapshot>();
let status = $state<"idle" | "loading" | "error">("idle");
let errorMessage = $state<string>();
let hydrated = $state(false);
let hydratePromise: Promise<void> | null = null;

export const isAuthenticated = () => !!session;
export const sessionStore = session;
export const sessionStatus = status;
export const sessionError = errorMessage;
export const sessionHydrated = hydrated;

browser.runtime.onMessage.addListener((message: unknown) => {
  if (!isSessionChangedEvent(message)) {
    return;
  }
  const event = message as SessionChangedEvent;
  session = event.session ?? undefined;
  errorMessage = undefined;
  status = "idle";
});

export async function hydrateSession(): Promise<void> {
  if (hydratePromise) {
    return hydratePromise;
  }

  if (hydrated) {
    return;
  }

  status = "loading";
  hydratePromise = (async () => {
    try {
      const response = await backgroundClient.getSession();
      session = response.session ?? undefined;
      errorMessage = undefined;
      status = "idle";
    } catch (error) {
      console.error("[session-store] hydrate failed", error);
      errorMessage = error instanceof Error ? error.message : "Unable to load session";
      status = "error";
    } finally {
      hydrated = true;
      hydratePromise = null;
    }
  })();

  await hydratePromise;
}

export async function login(identifier: string, password: string): Promise<boolean> {
  status = "loading";
  errorMessage = undefined;
  try {
    const response = await backgroundClient.login(identifier, password);
    if (!response.ok) {
      errorMessage = response.error;
      status = "idle";
      return false;
    }
    session = response.session;
    status = "idle";
    return true;
  } catch (error) {
    console.error("[session-store] login failed", error);
    errorMessage = error instanceof Error ? error.message : "Unable to login";
    status = "idle";
    return false;
  }
}

export async function logout(): Promise<void> {
  status = "loading";
  errorMessage = undefined;
  try {
    await backgroundClient.logout();
    session = undefined;
  } catch (error) {
    console.error("[session-store] logout failed", error);
    errorMessage = error instanceof Error ? error.message : "Unable to logout";
  } finally {
    status = "idle";
  }
}

export async function refreshSession(): Promise<void> {
  status = "loading";
  try {
    const response = await backgroundClient.getSession();
    session = response.session ?? undefined;
    errorMessage = undefined;
    status = "idle";
  } catch (error) {
    console.error("[session-store] refresh failed", error);
    errorMessage = error instanceof Error ? error.message : "Unable to refresh session";
    status = "error";
  }
}
