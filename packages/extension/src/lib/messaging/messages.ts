import type { FeedRequest, FeedResult } from "$lib/types/feed";
import type { SessionSnapshot } from "$lib/types/session";

export type BackgroundRequest =
  | { type: "session:get" }
  | { type: "session:login"; identifier: string; password: string }
  | { type: "session:logout" }
  | { type: "feed:get"; request: FeedRequest };

export type BackgroundError = { type: "error"; error: string };
export type SessionResponse = { type: "session"; session?: SessionSnapshot; authenticated: boolean };
export type LoginResponseOk = { type: "session:login"; ok: true; session: SessionSnapshot };
export type LoginResponseError = { type: "session:login"; ok: false; error: string };
export type LogoutResponse = { type: "session:logout"; ok: true };
export type FeedResponseOk = { type: "feed"; ok: true; result: FeedResult };
export type FeedResponseError = { type: "feed"; ok: false; error: string };
export type SessionChangedEvent = { type: "session:changed"; session?: SessionSnapshot; authenticated: boolean };

export type BackgroundResponse =
  | SessionResponse
  | LoginResponseOk
  | LoginResponseError
  | LogoutResponse
  | FeedResponseOk
  | FeedResponseError
  | BackgroundError;

export type RuntimeMessage = BackgroundResponse | SessionChangedEvent;

export const isBackgroundRequest = (input: unknown): input is BackgroundRequest => {
  if (!input || typeof input !== "object" || !("type" in input)) {
    return false;
  }
  const { type } = input as { type: string };
  switch (type) {
    case "session:get":
    case "session:logout":
      return true;
    case "session:login": {
      const req = input as { identifier?: unknown; password?: unknown };
      return typeof req.identifier === "string" && typeof req.password === "string";
    }
    case "feed:get": {
      const req = (input as { request?: FeedRequest }).request;
      return !!req && typeof req === "object" && "kind" in req;
    }
    default:
      return false;
  }
};

export const isSessionChangedEvent = (input: unknown): input is SessionChangedEvent => {
  if (!input || typeof input !== "object") {
    return false;
  }
  const message = input as { type?: unknown };
  if (message.type !== "session:changed") {
    return false;
  }
  return true;
};
