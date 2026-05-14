import { redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const token = cookies.get(SESSION_COOKIE);
  const cookieHeader = token ? `${SESSION_COOKIE}=${token}` : void 0;
  const [workspacesResult, runsResult, healthResult, comparisonResult, incidentsResult, storageResult] = await Promise.all([
    apiFetch("/api/workspaces", { cookie: cookieHeader }),
    apiFetch("/api/me/runs", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/health", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/comparison?days=7", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/incidents", { cookie: cookieHeader }),
    apiFetch("/api/widgets/workspaces/storage", { cookie: cookieHeader })
  ]);
  return {
    workspaces: workspacesResult.ok ? workspacesResult.data ?? [] : [],
    recentRuns: runsResult.ok ? runsResult.data ?? [] : [],
    error: workspacesResult.ok ? null : workspacesResult.error || "failed to load workspaces",
    health: healthResult.ok ? healthResult.data ?? [] : [],
    comparison: comparisonResult.ok ? comparisonResult.data ?? [] : [],
    incidents: incidentsResult.ok ? incidentsResult.data ?? [] : [],
    storage: storageResult.ok ? storageResult.data ?? [] : []
  };
};
export {
  load
};
