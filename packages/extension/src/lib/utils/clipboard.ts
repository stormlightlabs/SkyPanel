/**
 * Clipboard utility functions for copying text to the clipboard.
 *
 * Provides a simple async interface to the Clipboard API with error handling.
 * Falls back gracefully if clipboard access is denied or unavailable.
 */

/**
 * Copies text to the clipboard using the modern Clipboard API.
 *
 * Returns true if successful, false if the operation fails or clipboard access is unavailable.
 * The caller is responsible for showing user feedback.
 */
export async function copyToClipboard(text: string): Promise<boolean> {
	if (!navigator.clipboard) {
		console.error('[clipboard] Clipboard API not available');
		return false;
	}

	try {
		await navigator.clipboard.writeText(text);
		return true;
	} catch (error) {
		console.error('[clipboard] Failed to copy text:', error);
		return false;
	}
}
