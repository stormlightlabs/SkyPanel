/**
 * Type definitions for user profiles from the ATProto API.
 *
 * Re-exports profile view types and provides utility types for profile
 * state management. Profiles include metadata like avatar, banner, bio,
 * follower/following counts, and verification status.
 */

import type { AppBskyActorDefs } from '@atproto/api';

export type ProfileView = AppBskyActorDefs.ProfileView;
export type ProfileViewDetailed = AppBskyActorDefs.ProfileViewDetailed;
export type ProfileStatus = 'idle' | 'loading' | 'error';
/**
 * Profile state for loading and error handling.
 */
export type ProfileState = { profile?: ProfileViewDetailed; status: ProfileStatus; errorMessage?: string };
