function statusColor(status) {
  switch (status) {
    case "ready":
    case "succeeded":
      return "#22c55e";
    case "running":
    case "queued":
    case "cloning":
    case "pending":
      return "#eab308";
    case "failed":
    case "error":
      return "#ef4444";
    case "cancelled":
      return "var(--text-dim)";
    default:
      return "var(--text-dim)";
  }
}
function formatRelative(iso) {
  if (!iso) return "—";
  const ms = Date.now() - new Date(iso).getTime();
  const min = Math.floor(ms / 6e4);
  if (min < 1) return "just now";
  if (min < 60) return `${min}m ago`;
  const hr = Math.floor(min / 60);
  if (hr < 24) return `${hr}h ago`;
  const d = Math.floor(hr / 24);
  return `${d}d ago`;
}
function shortSha(sha, len = 7) {
  return sha ? sha.slice(0, len) : "—";
}
function formatDuration(start, end) {
  if (!start) return "—";
  const startMs = new Date(start).getTime();
  const endMs = end ? new Date(end).getTime() : Date.now();
  const sec = Math.round((endMs - startMs) / 1e3);
  if (sec < 60) return `${sec}s`;
  return `${Math.floor(sec / 60)}m ${sec % 60}s`;
}
function freshnessClass(iso) {
  if (!iso) return "fresh-stale";
  const days = (Date.now() - new Date(iso).getTime()) / (24 * 60 * 60 * 1e3);
  if (days < 7) return "fresh-good";
  if (days < 30) return "fresh-warn";
  return "fresh-stale";
}

export { statusColor as a, formatRelative as b, formatDuration as c, freshnessClass as f, shortSha as s };
//# sourceMappingURL=utils2-BUPlP7zG.js.map
