import { backgroundClient } from '$lib/client/background-client';
import type { ProfileViewDetailed } from '$lib/types/profile';

/**
 * Manages profile state using Svelte 5 runes.
 *
 * Provides reactive state management for user profile data, including loading
 * states and error handling. Coordinates with background script via
 * BackgroundClient to fetch profile information.
 */
class ProfileStore {
	private static instance: ProfileStore;

	private profile = $state<ProfileViewDetailed>();
	private status = $state<'idle' | 'loading' | 'error'>('idle');
	private errorMessage = $state<string>();
	isLoading = $derived(this.status === 'loading');

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

	/**
	 * Fetches profile data for the specified actor.
	 *
	 * If no actor is provided, fetches the profile for the currently authenticated user.
	 * Sets loading state during fetch and handles errors with user-friendly messages.
	 */
	async load(actor?: string): Promise<void> {
		this.status = 'loading';
		this.errorMessage = undefined;

		try {
			const response = await backgroundClient.getProfile(actor);

			if (response.ok) {
				this.profile = response.profile;
				this.status = 'idle';
			} else {
				this.errorMessage = response.error;
				this.status = 'error';
			}
		} catch (error) {
			console.error('[profile-store] load failed', error);
			this.errorMessage = error instanceof Error ? error.message : 'Unable to load profile';
			this.status = 'error';
		}
	}

	/**
	 * Clears the current profile and resets state.
	 */
	reset(): void {
		this.profile = undefined;
		this.status = 'idle';
		this.errorMessage = undefined;
	}
}

export const profileStore = ProfileStore.getInstance();
