import os


def load_env(path: str = ".env") -> dict:
    """Load environment variables from a simple .env file (KEY=VALUE lines, no quotes)."""
    env = {}
    try:
        with open(path, "r", encoding="utf-8") as f:
            for line in f:
                line = line.strip()
                if not line or line.startswith("#"):
                    continue
                if "=" in line:
                    key, val = line.split("=", 1)
                    env[key.strip()] = val.strip()
    except FileNotFoundError:
        pass
    for k, v in env.items():
        os.environ.setdefault(k, v)
    return env


def get_config() -> dict:
    cfg = load_env()
    for k in ("PDSHOST", "BLUESKY_HANDLE", "BLUESKY_PASSWORD"):
        if k not in cfg and k not in os.environ:
            raise RuntimeError(f"Missing required env var: {k}")
    return {
        "pds_host": os.environ.get("PDSHOST", cfg.get("PDSHOST")),
        "handle": os.environ.get("BLUESKY_HANDLE", cfg.get("BLUESKY_HANDLE")),
        "password": os.environ.get("BLUESKY_PASSWORD", cfg.get("BLUESKY_PASSWORD")),
    }
