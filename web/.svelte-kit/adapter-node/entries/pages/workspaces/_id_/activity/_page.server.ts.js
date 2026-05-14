import { redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const before = url.searchParams.get("before") || "";
  const limit = 50;
  const feedUrl = before ? `/api/workspaces/${params.id}/feed?limit=${limit}&before=${before}` : `/api/workspaces/${params.id}/feed?limit=${limit}`;
  const [wsResp, feedResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(feedUrl, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  const events = feedResp.ok ? feedResp.data ?? [] : [];
  const nextCursor = events.length >= limit ? String(events[events.length - 1].id) : null;
  return {
    workspace: wsResp.data,
    events,
    feedError: feedResp.ok ? null : feedResp.error || "failed to load activity feed",
    nextCursor,
    currentBefore: before || null
  };
};
export {
  load
};
