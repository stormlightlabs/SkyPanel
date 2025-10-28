/**
 * Toast notification system using Svelte 5 runes.
 *
 * Provides a reactive store for managing toast notifications with automatic
 * dismissal and manual removal. Supports different toast types (success, error,
 * info) and configurable display duration.
 */

export type ToastType = 'success' | 'error' | 'info';

export type Toast = { id: string; message: string; type: ToastType; duration?: number };

class ToastStore {
	private static instance: ToastStore;

	private toasts = $state<Toast[]>([]);
	private nextId = 0;

	private constructor() {}

	static getInstance(): ToastStore {
		if (!ToastStore.instance) {
			ToastStore.instance = new ToastStore();
		}
		return ToastStore.instance;
	}

	get all(): Toast[] {
		return this.toasts;
	}

	/**
	 * Adds a toast notification to the queue.
	 *
	 * Automatically dismisses after the specified duration (default 3000ms).
	 * Returns the toast ID for manual dismissal if needed.
	 */
	add(message: string, type: ToastType = 'info', duration = 3000): string {
		const id = `toast-${this.nextId++}`;
		const toast: Toast = { id, message, type, duration };

		this.toasts = [...this.toasts, toast];

		if (duration > 0) {
			setTimeout(() => this.remove(id), duration);
		}

		return id;
	}

	/**
	 * Shows a success toast with a green checkmark style.
	 */
	success(message: string, duration = 3000): string {
		return this.add(message, 'success', duration);
	}

	/**
	 * Shows an error toast with a red warning style.
	 */
	error(message: string, duration = 4000): string {
		return this.add(message, 'error', duration);
	}

	/**
	 * Shows an info toast with a blue info style.
	 */
	info(message: string, duration = 3000): string {
		return this.add(message, 'info', duration);
	}

	/**
	 * Removes a specific toast by ID.
	 */
	remove(id: string): void {
		this.toasts = this.toasts.filter((t) => t.id !== id);
	}

	/**
	 * Clears all active toasts.
	 */
	clear(): void {
		this.toasts = [];
	}
}

export const toastStore = ToastStore.getInstance();
