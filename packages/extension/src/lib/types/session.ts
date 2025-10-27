import type { AtpSessionData } from "@atproto/api";

export const DEFAULT_SERVICE_URL = "https://bsky.social";

export type SessionSnapshot = {
  did: string;
  handle: string;
  email?: string;
  emailConfirmed?: boolean;
  emailAuthFactor?: boolean;
  active: boolean;
  status?: string;
  serviceUrl: string;
};

export type PersistedSession = { session: AtpSessionData; serviceUrl: string; storedAt: number };

export type SessionState = { session: SessionSnapshot | null; authenticated: boolean };
