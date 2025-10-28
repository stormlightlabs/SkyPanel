import { browser } from 'wxt/browser';
import type { FeedRequest } from '$lib/types/feed';
import type { ComputedFeedRequest } from '$lib/types/computed-feed';
import type {
	BackgroundRequest,
	BackgroundResponse,
	FeedResponseError,
	FeedResponseOk,
	LoginResponseError,
	LoginResponseOk,
	LogoutResponse,
	SessionResponse,
	ComputedFeedResponseOk,
	ComputedFeedResponseError,
	ProfileResponseOk,
	ProfileResponseError
} from '$lib/messaging/messages';

export class BackgroundClient {
	private readonly maxRetries = 5;
	private readonly initialDelayMs = 50;

	private async sleep(ms: number): Promise<void> {
		return new Promise((resolve) => setTimeout(resolve, ms));
	}

	/**
	 * Send a message to the background script with exponential backoff retry logic.
	 *
	 * This handles the race condition where the sidepanel loads before the background script's message listener is fully registered.
	 * The background service worker may not be ready immediately when the extension starts or when the sidepanel opens.
	 *
	 * @param request The background request to send
	 * @returns The response from the background script
	 * @throws {Error} if all retry attempts fail
	 */
	async request<T extends BackgroundResponse>(request: BackgroundRequest): Promise<T> {
		let lastError: Error | undefined;

		for (let attempt = 0; attempt < this.maxRetries; attempt++) {
			try {
				const response = (await browser.runtime.sendMessage(request)) as BackgroundResponse | undefined;

				if (response) {
					if (attempt > 0) {
						console.log(`[BackgroundClient] request "${request.type}" succeeded on attempt ${attempt + 1}`);
					}
					return response as T;
				}

				lastError = new Error(`No response received for background request "${request.type}"`);
			} catch (error) {
				lastError = error instanceof Error ? error : new Error(String(error));
				console.warn(`[BackgroundClient] request "${request.type}" attempt ${attempt + 1} failed:`, lastError.message);
			}

			if (attempt < this.maxRetries - 1) {
				const delayMs = this.initialDelayMs * Math.pow(2, attempt);
				console.log(`[BackgroundClient] retrying "${request.type}" in ${delayMs}ms...`);
				await this.sleep(delayMs);
			}
		}

		throw new Error(
			`Background request "${request.type}" failed after ${this.maxRetries} attempts: ${lastError?.message || 'Unknown error'}`
		);
	}

	async getSession(): Promise<SessionResponse> {
		return this.request<SessionResponse>({ type: 'session:get' });
	}

	async login(identifier: string, password: string): Promise<LoginResponseOk | LoginResponseError> {
		return this.request<LoginResponseOk | LoginResponseError>({ type: 'session:login', identifier, password });
	}

	async logout(): Promise<LogoutResponse> {
		return this.request<LogoutResponse>({ type: 'session:logout' });
	}

	async fetchFeed(request: FeedRequest): Promise<FeedResponseOk | FeedResponseError> {
		return this.request<FeedResponseOk | FeedResponseError>({ type: 'feed:get', request });
	}

	async fetchComputedFeed(request: ComputedFeedRequest): Promise<ComputedFeedResponseOk | ComputedFeedResponseError> {
		return this.request<ComputedFeedResponseOk | ComputedFeedResponseError>({ type: 'computed-feed:get', request });
	}

	async getProfile(actor?: string): Promise<ProfileResponseOk | ProfileResponseError> {
		return this.request<ProfileResponseOk | ProfileResponseError>({ type: 'profile:get', actor });
	}
}

export const backgroundClient = new BackgroundClient();
