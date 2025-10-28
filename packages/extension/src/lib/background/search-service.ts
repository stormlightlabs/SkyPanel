import type { SessionManager } from './session-manager';
import type { SearchRequest, SearchResult } from '$lib/types/search';

/**
 * Default limit for search results per page.
 */
const DEFAULT_LIMIT = 25;

/**
 * Service for searching posts via app.bsky.feed.searchPosts.
 *
 * Supports filtering by author, language, domain, tags, and date ranges.
 * All operations require authentication via AtpAgent.
 */
export class SearchService {
	constructor(private readonly sessions: SessionManager) {}

	/**
	 * Search for posts matching the given query and filters.
	 *
	 * Supports rich query syntax including:
	 * - Basic text search
	 * - Author filtering (via author parameter)
	 * - Hashtag filtering (via tag parameter)
	 * - Domain filtering (via domain parameter)
	 * - Date range filtering (via since/until parameters)
	 * - Language filtering (via lang parameter)
	 *
	 * @param request - Search parameters
	 * @returns Search results with posts and pagination cursor
	 * @throws Error if not authenticated or API request fails
	 */
	async search(request: SearchRequest): Promise<SearchResult> {
		const agent = this.sessions.agent;
		if (!agent.hasSession) {
			throw new Error('Not authenticated - please log in to search posts');
		}

		try {
			const response = await agent.app.bsky.feed.searchPosts({
				q: request.query,
				author: request.author,
				lang: request.lang,
				domain: request.domain,
				url: request.url,
				tag: request.tag,
				since: request.since,
				until: request.until,
				cursor: request.cursor,
				limit: request.limit ?? DEFAULT_LIMIT
			});

			return { posts: response.data.posts, cursor: response.data.cursor, hitsTotal: response.data.hitsTotal };
		} catch (error) {
			console.error('[SearchService] Search failed', { request, error });
			throw error;
		}
	}
}
