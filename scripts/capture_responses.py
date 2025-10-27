import json
from config import get_config
from session import load_session
from get_feeds import get_timeline, get_follows, get_author_feed, search_posts


def main():
    cfg = get_config()
    session = load_session()
    timeline = get_timeline(session, cfg, limit=5)
    with open("response_timeline.json", "w") as f:
        json.dump(timeline, f, indent=2)
    print("Saved timeline response to response_timeline.json")

    follows = get_follows(session, cfg, actor=cfg["handle"], limit=5)
    with open("response_follows.json", "w") as f:
        json.dump(follows, f, indent=2)
    print("Saved follows response to response_follows.json")

    author_feed = get_author_feed(session, cfg, actor=cfg["handle"], limit=5)
    with open("response_author_feed.json", "w") as f:
        json.dump(author_feed, f, indent=2)
    print("Saved author feed response to response_author_feed.json")

    search = search_posts(session, cfg, query="bluesky", limit=5)
    with open("response_search_posts.json", "w") as f:
        json.dump(search, f, indent=2)
    print("Saved search posts response to response_search_posts.json")


if __name__ == "__main__":
    main()
