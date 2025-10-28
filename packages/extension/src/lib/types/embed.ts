/**
 * Type definitions for post embeds from the ATProto API.
 *
 * Re-exports and provides type guards for different embed types including images, videos, external links, and records (quote posts).
 * These types match the Bluesky/ATProto embed view structures.
 */

import type {
	AppBskyEmbedImages,
	AppBskyEmbedExternal,
	AppBskyEmbedRecord,
	AppBskyEmbedRecordWithMedia,
	AppBskyEmbedVideo
} from '@atproto/api';

export type ImagesEmbed = AppBskyEmbedImages.View;
export type ExternalEmbed = AppBskyEmbedExternal.View;
export type RecordEmbed = AppBskyEmbedRecord.View;
export type RecordWithMediaEmbed = AppBskyEmbedRecordWithMedia.View;
export type VideoEmbed = AppBskyEmbedVideo.View;
export type PostEmbed = ImagesEmbed | ExternalEmbed | RecordEmbed | RecordWithMediaEmbed | VideoEmbed;

/**
 * Type guard to check if an embed is an images embed.
 */
export function isImagesEmbed(embed: PostEmbed): embed is ImagesEmbed {
	return embed.$type === 'app.bsky.embed.images#view';
}

/**
 * Type guard to check if an embed is an external link embed.
 */
export function isExternalEmbed(embed: PostEmbed): embed is ExternalEmbed {
	return embed.$type === 'app.bsky.embed.external#view';
}

/**
 * Type guard to check if an embed is a record (quote post) embed.
 */
export function isRecordEmbed(embed: PostEmbed): embed is RecordEmbed {
	return embed.$type === 'app.bsky.embed.record#view';
}

/**
 * Type guard to check if an embed is a record with media embed.
 */
export function isRecordWithMediaEmbed(embed: PostEmbed): embed is RecordWithMediaEmbed {
	return embed.$type === 'app.bsky.embed.recordWithMedia#view';
}

/**
 * Type guard to check if an embed is a video embed.
 */
export function isVideoEmbed(embed: PostEmbed): embed is VideoEmbed {
	return embed.$type === 'app.bsky.embed.video#view';
}
