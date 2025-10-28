import { browser } from "wxt/browser";
import { FeedService } from "$lib/background/feed-service";
import { SessionManager } from "$lib/background/session-manager";
import { GraphService } from "$lib/background/graph-service";
import { FeedComputer } from "$lib/background/feed-computer";
import { computedFeedStorage } from "$lib/storage/computed-feed-storage";
import {
  isBackgroundRequest,
  type BackgroundRequest,
  type BackgroundResponse,
  type SessionChangedEvent,
} from "$lib/messaging/messages";

const sessions = new SessionManager();
const feeds = new FeedService(sessions);
const graphs = new GraphService(sessions);
const computer = new FeedComputer(sessions, graphs, feeds);

type ChromiumSidePanelApi = {
  open(options?: { windowId?: number }): Promise<void>;
  setPanelBehavior?(options: { openPanelOnActionClick: boolean }): Promise<void>;
};

async function handleRequest(request: BackgroundRequest): Promise<BackgroundResponse> {
  switch (request.type) {
    case "session:get": {
      const response = { type: "session" as const, session: sessions.snapshot, authenticated: sessions.authenticated };
      console.log("[background] session:get response:", { hasSession: !!sessions.snapshot, authenticated: sessions.authenticated });
      return response;
    }
    case "session:login": {
      try {
        console.log("[background] login attempt for:", request.identifier);
        const session = await sessions.login(request.identifier, request.password);
        console.log("[background] login succeeded:", { did: session.did, handle: session.handle });
        return { type: "session:login", ok: true, session };
      } catch (error) {
        console.error("[background] login failed", error);
        return { type: "session:login", ok: false, error: error instanceof Error ? error.message : "Unable to login" };
      }
    }
    case "session:logout": {
      await sessions.logout();
      await computedFeedStorage.clearAll();
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
    case "computed-feed:get": {
      try {
        const { request: feedRequest } = request;
        const forceRefresh = feedRequest.forceRefresh ?? false;

        if (!forceRefresh) {
          const cached = await computedFeedStorage.load(feedRequest.kind);
          if (cached) {
            return { type: "computed-feed", ok: true, result: cached.data };
          }
        }

        let result;
        switch (feedRequest.kind) {
          case "mutuals":
            result = await computer.computeMutualsFeed(feedRequest.cursor, feedRequest.limit);
            break;
          case "quiet":
            result = await computer.computeQuietPostersFeed(feedRequest.cursor, feedRequest.limit);
            break;
          default: {
            const exhaustive: never = feedRequest;
            throw new Error(`Unsupported computed feed kind: ${(exhaustive as { kind: string }).kind}`);
          }
        }

        const cached = computedFeedStorage.createCached(feedRequest.kind, result);
        await computedFeedStorage.save(cached);

        return { type: "computed-feed", ok: true, result };
      } catch (error) {
        console.error("[background] computed feed fetch failed", error);
        return {
          type: "computed-feed",
          ok: false,
          error: error instanceof Error ? error.message : "Unable to compute feed",
        };
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

  const sidePanel = (browser as typeof browser & { sidePanel?: ChromiumSidePanelApi }).sidePanel;
  if (sidePanel) {
    sidePanel.setPanelBehavior?.({ openPanelOnActionClick: true }).catch((error) => {
      console.warn("[background] failed to enable action side panel behavior", error);
    });

    browser.action?.onClicked.addListener((tab) => {
      const windowId = tab.windowId;
      sidePanel.open(windowId != null ? { windowId } : {}).catch((error) => {
        console.error("[background] failed to open side panel from action click", error);
      });
    });
  }

  sessions.subscribe((snapshot) => {
    const message: SessionChangedEvent = {
      type: "session:changed",
      session: snapshot,
      authenticated: sessions.authenticated,
    };
    browser.runtime.sendMessage(message).catch(
      // No-op: no listeners are available.
      () => void 0,
    );
  });

  browser.runtime.onMessage.addListener((message, _sender, sendResponse) => {
    console.log("[background] received message:", message);
    const isValid = isBackgroundRequest(message);
    console.log("[background] isBackgroundRequest:", isValid);

    if (!isValid) {
      console.log("[background] ignoring non-background message");
      return false;
    }

    console.log("[background] handling request:", message.type);
    handleRequest(message)
      .then((response) => {
        console.log("[background] sending response:", response);
        sendResponse(response);
      })
      .catch((error) => {
        console.error("[background] request handler error:", error);
        sendResponse({ type: "error", error: error instanceof Error ? error.message : "Unknown error" });
      });

    return true; // Required for async response in MV3
  });
});
