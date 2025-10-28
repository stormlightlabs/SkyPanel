import { backgroundClient } from '$lib/client/background-client';
import type { ProfileViewDetailed } from '$lib/types/profile';

/**
 * Manages profile state
 *
 * Provides reactive state management for user profile data, including loading states, error handling, and caching.
 * Coordinates with background script via BackgroundClient to fetch profile information.
 *
 * Supports both initial load and manual refresh.
 * Cache is managed by the background service layer with 10-minute TTL.
 */
class ProfileStore {
	private static instance: ProfileStore;

	private profile = $state<ProfileViewDetailed>();
	private status = $state<'idle' | 'loading' | 'refreshing' | 'error'>('idle');
	private errorMessage = $state<string>();
	private lastFetchedAt = $state<number>();
	private inflight = false;
	isLoading = $derived(this.status === 'loading' || this.status === 'refreshing');
	isRefreshing = $derived(this.status === 'refreshing');

	private constructor() {}

	static getInstance(): ProfileStore {
		if (!ProfileStore.instance) {
			ProfileStore.instance = new ProfileStore();
		}
		return ProfileStore.instance;
	}

	get currentProfile() {
		return this.profile;
	}

	get currentStatus() {
		return this.status;
	}

	get error() {
		return this.errorMessage;
	}

	get fetchedAt() {
		return this.lastFetchedAt;
	}

	/**
	 * Fetches profile data for the specified actor.
	 *
	 * If no actor is provided, fetches the profile for the currently authenticated user.
	 * Sets loading state during fetch and handles errors with user-friendly messages.
	 * Uses cache if available (10-minute TTL managed by background service).
	 */
	async load(actor?: string): Promise<void> {
		if (this.inflight) {
			console.warn('[profile-store] load already in progress, skipping');
			return;
		}

		this.inflight = true;
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			const response = await backgroundClient.getProfile(actor, false);

			if (response.ok) {
				this.profile = response.profile;
				this.lastFetchedAt = response.fetchedAt;
				this.status = 'idle';
			} else {
				this.errorMessage = response.error;
				this.status = 'error';
			}
		} catch (error) {
			console.error('[profile-store] load failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to load profile';
			this.status = 'error';
		} finally {
			this.inflight = false;
		}
	}

	/**
	 * Force refresh profile data, bypassing cache.
	 *
	 * Uses 'refreshing' state to distinguish from initial load.
	 */
	async refresh(actor?: string): Promise<void> {
		if (this.inflight) {
			console.warn('[profile-store] refresh already in progress, skipping');
			return;
		}

		this.inflight = true;
		this.status = 'refreshing';
		this.errorMessage = undefined;

		try {
			const response = await backgroundClient.getProfile(actor, true);

			if (response.ok) {
				this.profile = response.profile;
				this.lastFetchedAt = response.fetchedAt;
				this.status = 'idle';
			} else {
				this.errorMessage = response.error;
				this.status = 'error';
			}
		} catch (error) {
			console.error('[profile-store] refresh failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to refresh profile';
			this.status = 'error';
		} finally {
			this.inflight = false;
		}
	}

	/**
	 * Clears the current profile and resets state.
	 */
	reset(): void {
		this.profile = undefined;
		this.status = 'idle';
		this.errorMessage = undefined;
		this.lastFetchedAt = undefined;
		this.inflight = false;
	}
}

export const profileStore = ProfileStore.getInstance();
