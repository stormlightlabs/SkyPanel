import { browser } from "wxt/browser";
import { FeedService } from "$lib/background/feed-service";
import { SessionManager } from "$lib/background/session-manager";
import {
  isBackgroundRequest,
  type BackgroundRequest,
  type BackgroundResponse,
  type SessionChangedEvent,
} from "$lib/messaging/messages";

const sessions = new SessionManager();
const feeds = new FeedService(sessions);

async function handleRequest(request: BackgroundRequest): Promise<BackgroundResponse> {
  switch (request.type) {
    case "session:get": {
      return { type: "session", session: sessions.snapshot, authenticated: sessions.authenticated };
    }
    case "session:login": {
      try {
        const session = await sessions.login(request.identifier, request.password);
        return { type: "session:login", ok: true, session };
      } catch (error) {
        console.error("[background] login failed", error);
        return { type: "session:login", ok: false, error: error instanceof Error ? error.message : "Unable to login" };
      }
    }
    case "session:logout": {
      await sessions.logout();
      return { type: "session:logout", ok: true };
    }
    case "feed:get": {
      try {
        const result = await feeds.fetch(request.request);
        return { type: "feed", ok: true, result };
      } catch (error) {
        console.error("[background] feed fetch failed", error);
        return { type: "feed", ok: false, error: error instanceof Error ? error.message : "Unable to load feed" };
      }
    }
    default:
      return { type: "error", error: `Unsupported message ${(request as { type: string }).type}` };
  }
}

export default defineBackground(() => {
  sessions.resumeFromStorage().catch((error) => {
    console.warn("[background] failed to resume session", error);
  });

  sessions.subscribe((snapshot) => {
    const message: SessionChangedEvent = {
      type: "session:changed",
      session: snapshot,
      authenticated: sessions.authenticated,
    };
    browser.runtime.sendMessage(message).catch(() => {
      // No-op: no listeners are available.
    });
  });

  browser.runtime.onMessage.addListener((message) => {
    if (!isBackgroundRequest(message)) {
      return undefined;
    }
    return handleRequest(message);
  });
});
