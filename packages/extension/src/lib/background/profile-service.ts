import type { SessionManager } from './session-manager';
import type { ProfileViewDetailed } from '$lib/types/profile';

/**
 * Service for fetching user profile data from Bluesky via authenticated AtpAgent.
 *
 * Supports fetching detailed profile information including:
 * - Display name, handle, avatar, banner
 * - Bio/description
 * - Follower and following counts
 * - Post count and indexed date
 * - Labels and verification status
 */
export class ProfileService {
	constructor(private readonly sessions: SessionManager) {}

	/**
	 * Fetches detailed profile data for the specified actor.
	 *
	 * The actor can be a handle (e.g., "alice.bsky.social") or a DID.
	 * If no actor is provided, fetches the profile for the currently authenticated user.
	 *
	 * @param actor - Actor identifier (handle or DID), defaults to current user
	 * @returns Detailed profile view with all metadata
	 * @throws Error if not authenticated or API request fails
	 */
	async getProfile(actor?: string): Promise<ProfileViewDetailed> {
		const agent = this.sessions.agent;
		if (!agent.hasSession) {
			throw new Error('Not authenticated - please log in to fetch profile');
		}

		const targetActor = actor ?? agent.session?.did;
		if (!targetActor) {
			throw new Error('No actor specified and no session DID available');
		}

		try {
			const response = await agent.app.bsky.actor.getProfile({ actor: targetActor });
			return response.data;
		} catch (error) {
			console.error('[ProfileService] Get profile failed', { actor: targetActor, error });
			throw error;
		}
	}
}
