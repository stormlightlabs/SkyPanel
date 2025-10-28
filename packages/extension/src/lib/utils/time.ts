const SECOND = 1000;
const MINUTE = 60 * SECOND;
const HOUR = 60 * MINUTE;
const DAY = 24 * HOUR;

export function formatDistanceToNow(date: Date | number): string {
  const now = Date.now();
  const timestamp = typeof date === "number" ? date : date.getTime();
  const diff = Math.max(0, now - timestamp);

  if (diff < MINUTE) {
    const seconds = Math.floor(diff / SECOND);
    return seconds <= 1 ? "just now" : `${seconds}s ago`;
  }

  if (diff < HOUR) {
    const minutes = Math.floor(diff / MINUTE);
    return `${minutes}m ago`;
  }

  if (diff < DAY) {
    const hours = Math.floor(diff / HOUR);
    return `${hours}h ago`;
  }

  const days = Math.floor(diff / DAY);
  if (days < 7) {
    return `${days}d ago`;
  }

  const dateObj = typeof date === "number" ? new Date(date) : date;
  return dateObj.toLocaleDateString(undefined, {
    month: "short",
    day: "numeric",
    year: dateObj.getFullYear() !== new Date().getFullYear() ? "numeric" : undefined,
  });
}
