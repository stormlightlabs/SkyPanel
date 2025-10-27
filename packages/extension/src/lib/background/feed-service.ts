import type { FeedResult, FeedRequest } from "$lib/types/feed";
import type { SessionManager } from "./session-manager";

const DEFAULT_LIMIT = 30;

export class FeedService {
  constructor(private readonly sessions: SessionManager) {}

  async fetch(request: FeedRequest): Promise<FeedResult> {
    const agent = this.sessions.agent;
    if (!agent.hasSession) {
      throw new Error("Not authenticated");
    }

    switch (request.kind) {
      case "timeline": {
        const response = await agent.app.bsky.feed.getTimeline({
          cursor: request.cursor,
          limit: request.limit ?? DEFAULT_LIMIT,
        });
        return { kind: "timeline", cursor: response.data.cursor, feed: response.data.feed };
      }
      case "author": {
        const response = await agent.app.bsky.feed.getAuthorFeed({
          actor: request.actor,
          cursor: request.cursor,
          limit: request.limit ?? DEFAULT_LIMIT,
        });
        return { kind: "author", actor: request.actor, cursor: response.data.cursor, feed: response.data.feed };
      }
      case "list": {
        const response = await agent.app.bsky.feed.getListFeed({
          list: request.list,
          cursor: request.cursor,
          limit: request.limit ?? DEFAULT_LIMIT,
        });
        return { kind: "list", list: request.list, cursor: response.data.cursor, feed: response.data.feed };
      }
      default: {
        const exhaustive: never = request;
        throw new Error(`Unsupported feed request ${(exhaustive as { kind: string }).kind}`);
      }
    }
  }
}
