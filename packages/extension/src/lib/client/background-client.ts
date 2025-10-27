import { browser } from "wxt/browser";
import type { FeedRequest } from "$lib/types/feed";
import type {
  BackgroundRequest,
  BackgroundResponse,
  FeedResponseError,
  FeedResponseOk,
  LoginResponseError,
  LoginResponseOk,
  LogoutResponse,
  SessionResponse,
} from "$lib/messaging/messages";

class BackgroundClient {
  async request<T extends BackgroundResponse>(request: BackgroundRequest): Promise<T> {
    const response = (await browser.runtime.sendMessage(request)) as BackgroundResponse;
    return response as T;
  }

  async getSession(): Promise<SessionResponse> {
    return this.request<SessionResponse>({ type: "session:get" });
  }

  async login(identifier: string, password: string): Promise<LoginResponseOk | LoginResponseError> {
    return this.request<LoginResponseOk | LoginResponseError>({ type: "session:login", identifier, password });
  }

  async logout(): Promise<LogoutResponse> {
    return this.request<LogoutResponse>({ type: "session:logout" });
  }

  async fetchFeed(request: FeedRequest): Promise<FeedResponseOk | FeedResponseError> {
    return this.request<FeedResponseOk | FeedResponseError>({ type: "feed:get", request });
  }
}

export const backgroundClient = new BackgroundClient();
