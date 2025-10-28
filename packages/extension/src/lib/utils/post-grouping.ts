/**
 * Utilities for grouping consecutive posts by author.
 *
 * Groups multiple consecutive posts from the same author into collapsible blocks to reduce visual clutter from prolific posters.
 */

import type { AppBskyFeedDefs } from '@atproto/api';

/**
 * Represents a group of consecutive posts from the same author.
 *
 * Groups are only created for 2+ consecutive posts.
 */
export type PostGroup = {
	author: AppBskyFeedDefs.FeedViewPost['post']['author'];
	posts: AppBskyFeedDefs.FeedViewPost[];
	firstPostAt: string;
	lastPostAt: string;
	isUnread: boolean;
	collapsed: boolean;
};

/**
 * Represents either a grouped set of posts or a single ungrouped post.
 */
export type FeedItem = { type: 'group'; group: PostGroup } | { type: 'single'; post: AppBskyFeedDefs.FeedViewPost };

type UnreadFn = (authorDid: string, timestamp: string) => boolean;
/**
 * Groups consecutive posts by the same author into collapsible groups.
 *
 * Single posts are returned as ungrouped items for normal display.
 *
 * @param posts - Array of feed posts in chronological order (newest first)
 * @param isUnreadFn - Function to check if a post is unread based on author DID and timestamp
 * @returns Array of feed items (groups or singles) maintaining original order
 */
export function groupConsecutivePosts(posts: AppBskyFeedDefs.FeedViewPost[], isUnreadFn: UnreadFn): FeedItem[] {
	if (posts.length === 0) {
		return [];
	}

	const items: FeedItem[] = [];
	let pendingGroup: PostGroup | null = null;

	for (const post of posts) {
		const authorDid = post.post.author.did;
		const postTime = post.post.indexedAt;
		const isUnread = isUnreadFn(authorDid, postTime);

		if (pendingGroup && pendingGroup.author.did === authorDid) {
			pendingGroup.posts.push(post);
			pendingGroup.lastPostAt = postTime;
			pendingGroup.isUnread = pendingGroup.isUnread || isUnread;
		} else {
			if (pendingGroup) {
				if (pendingGroup.posts.length === 1) {
					items.push({ type: 'single', post: pendingGroup.posts[0] });
				} else {
					items.push({ type: 'group', group: pendingGroup });
				}
			}

			pendingGroup = {
				author: post.post.author,
				posts: [post],
				firstPostAt: postTime,
				lastPostAt: postTime,
				isUnread: isUnread,
				collapsed: isUnread
			};
		}
	}

	if (pendingGroup) {
		if (pendingGroup.posts.length === 1) {
			items.push({ type: 'single', post: pendingGroup.posts[0] });
		} else {
			items.push({ type: 'group', group: pendingGroup });
		}
	}

	return items;
}

/**
 * Get the latest (most recent) post timestamp from a group.
 *
 * Helps determine mark-read timestamp when scrolling past.
 */
export function getGroupLatestTimestamp(group: PostGroup): number {
	return new Date(group.firstPostAt).getTime();
}

/**
 * Get the oldest post timestamp from a group.
 *
 * Helps determine if any posts in the group are unread.
 */
export function getGroupOldestTimestamp(group: PostGroup): number {
	return new Date(group.lastPostAt).getTime();
}
