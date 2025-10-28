/**
 * Download utility functions for fetching and saving media files.
 *
 * Provides functionality to download images and videos from URLs by creating temporary anchor elements and triggering browser downloads.
 * Handles filename extraction from URLs and error cases gracefully.
 */

/**
 * Downloads a file from a URL by creating a temporary anchor element.
 *
 * Attempts to extract a meaningful filename from the URL.
 * If the download fails (e.g., CORS restrictions), returns false so the caller can show appropriate user feedback.
 */
export async function downloadFile(url: string, filename?: string): Promise<boolean> {
	try {
		const response = await fetch(url);
		if (!response.ok) {
			console.error(`[download] Failed to fetch ${url}: ${response.status}`);
			return false;
		}

		const blob = await response.blob();
		const blobUrl = URL.createObjectURL(blob);

		const finalFilename = filename || extractFilename(url);

		const anchor = document.createElement('a');
		anchor.href = blobUrl;
		anchor.download = finalFilename;
		document.body.append(anchor);
		anchor.click();
		anchor.remove();

		setTimeout(() => URL.revokeObjectURL(blobUrl), 100);

		return true;
	} catch (error) {
		console.error('[download] Failed to download file:', error);
		return false;
	}
}

/**
 * Extracts a filename from a URL path.
 *
 * Falls back to a generic "download" filename if extraction fails.
 */
function extractFilename(url: string): string {
	try {
		const urlObj = new URL(url);
		const pathname = urlObj.pathname;
		const parts = pathname.split('/');
		const filename = parts.at(-1);
		return filename || 'download';
	} catch {
		return 'download';
	}
}

/**
 * Downloads an image file with appropriate extension detection.
 */
export async function downloadImage(url: string, filename?: string): Promise<boolean> {
	const finalFilename = filename || `image-${Date.now()}.jpg`;
	return downloadFile(url, finalFilename);
}

/**
 * Downloads a video file with appropriate extension detection.
 */
export async function downloadVideo(url: string, filename?: string): Promise<boolean> {
	const finalFilename = filename || `video-${Date.now()}.mp4`;
	return downloadFile(url, finalFilename);
}
