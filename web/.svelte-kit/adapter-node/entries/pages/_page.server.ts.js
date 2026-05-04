import { redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const token = cookies.get(SESSION_COOKIE);
  const cookieHeader = token ? `${SESSION_COOKIE}=${token}` : void 0;
  const [workspacesResult, runsResult] = await Promise.all([
    apiFetch("/api/workspaces", { cookie: cookieHeader }),
    apiFetch("/api/me/runs", { cookie: cookieHeader })
  ]);
  return {
    workspaces: workspacesResult.ok ? workspacesResult.data ?? [] : [],
    recentRuns: runsResult.ok ? runsResult.data ?? [] : [],
    error: workspacesResult.ok ? null : workspacesResult.error || "failed to load workspaces"
  };
};
export {
  load
};
