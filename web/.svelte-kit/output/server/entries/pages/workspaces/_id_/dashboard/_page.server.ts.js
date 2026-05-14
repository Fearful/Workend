import { redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const days = Math.max(1, Math.min(365, parseInt(url.searchParams.get("days") || "30", 10) || 30));
  const [wsResp, sumResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/dashboard?days=${days}`, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  return {
    workspace: wsResp.data,
    summary: sumResp.ok ? sumResp.data ?? null : null,
    summaryError: sumResp.ok ? null : sumResp.error || "failed to load dashboard",
    days
  };
};
export {
  load
};
