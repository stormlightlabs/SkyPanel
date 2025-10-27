import json
import urllib.request
import urllib.parse
from config import get_config
from session import load_session


def fetch(endpoint: str, params: dict, session: dict, cfg: dict) -> dict:
    base = cfg["pds_host"].rstrip("/") + "/xrpc/" + endpoint
    if params:
        qs = urllib.parse.urlencode(params)
        url = f"{base}?{qs}"
    else:
        url = base
    req = urllib.request.Request(url, method="GET")
    token = session.get("accessJwt") or session.get("jwt")
    if not token:
        raise RuntimeError("No access JWT found in session")
    req.add_header("Authorization", f"Bearer {token}")
    with urllib.request.urlopen(req) as resp:
        resp_data = resp.read().decode("utf-8")
        return json.loads(resp_data)


def get_timeline(
    session: dict, cfg: dict, limit: int = 50, cursor: str | None = None
) -> dict:
    return fetch(
        "app.bsky.feed.getTimeline",
        {"limit": str(limit)}
        if cursor is None
        else {"limit": str(limit), "cursor": cursor},
        session,
        cfg,
    )


def get_follows(
    session: dict, cfg: dict, actor: str, limit: int = 50, cursor: str | None = None
) -> dict:
    return fetch(
        "app.bsky.graph.getFollows",
        {"actor": actor, "limit": str(limit)}
        if cursor is None
        else {"actor": actor, "limit": str(limit), "cursor": cursor},
        session,
        cfg,
    )


def get_author_feed(
    session: dict, cfg: dict, actor: str, limit: int = 50, cursor: str | None = None
) -> dict:
    return fetch(
        "app.bsky.feed.getAuthorFeed",
        {"actor": actor, "limit": str(limit)}
        if cursor is None
        else {"actor": actor, "limit": str(limit), "cursor": cursor},
        session,
        cfg,
    )


def search_posts(
    session: dict, cfg: dict, query: str, limit: int = 50, cursor: str | None = None
) -> dict:
    return fetch(
        "app.bsky.feed.searchPosts",
        {"query": query, "limit": str(limit)}
        if cursor is None
        else {"query": query, "limit": str(limit), "cursor": cursor},
        session,
        cfg,
    )


def main():
    cfg = get_config()
    session = load_session()

    timeline = get_timeline(session, cfg, limit=20)
    print("== Timeline ==")
    print(json.dumps(timeline, indent=2))

    follows = get_follows(session, cfg, actor=cfg["handle"], limit=20)
    print("== Follows ==")
    print(json.dumps(follows, indent=2))

    actor = input("Enter an actor handle or DID for author feed: ").strip()
    if actor:
        author_feed = get_author_feed(session, cfg, actor=actor, limit=20)
        print("== Author Feed for", actor, "==")
        print(json.dumps(author_feed, indent=2))

    query = input("Enter search query (or blank to skip): ").strip()
    if query:
        search = search_posts(session, cfg, query=query, limit=20)
        print("== Search Posts (", query, ") ==")
        print(json.dumps(search, indent=2))


if __name__ == "__main__":
    main()
