import json
import urllib.request
from config import get_config


def create_session(cfg: dict) -> dict:
    url = cfg["pds_host"].rstrip("/") + "/xrpc/com.atproto.server.createSession"
    payload = {"identifier": cfg["handle"], "password": cfg["password"]}
    data = json.dumps(payload).encode("utf-8")
    req = urllib.request.Request(url, data=data, method="POST")
    req.add_header("Content-Type", "application/json")
    with urllib.request.urlopen(req) as resp:
        resp_data = resp.read().decode("utf-8")
        return json.loads(resp_data)


def save_session(sess: dict, path: str = "session.json") -> None:
    with open(path, "w", encoding="utf-8") as f:
        json.dump(sess, f, indent=2)


def load_session(path: str = "session.json") -> dict:
    with open(path, "r", encoding="utf-8") as f:
        return json.load(f)


def main():
    cfg = get_config()
    print("Creating session…")
    sess = create_session(cfg)
    print("Saving session to session.json")
    save_session(sess)


if __name__ == "__main__":
    main()
