import { fail, error, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [wsResp, projResp, membersResp, activityResp, issuesResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/projects`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/members`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/activity?limit=30`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/recent-issues`, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  return {
    workspace: wsResp.data,
    projects: projResp.ok ? projResp.data ?? [] : [],
    projectsError: projResp.ok ? null : projResp.error || "failed to load projects",
    members: membersResp.ok ? membersResp.data ?? [] : [],
    activity: activityResp.ok ? activityResp.data ?? [] : [],
    recentIssues: issuesResp.ok ? issuesResp.data ?? [] : []
  };
};
const actions = {
  delete: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/workspaces/${params.id}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) throw error(result.status, result.error || "delete failed");
    throw redirect(303, "/");
  },
  addMember: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const data = await request.formData();
    const email = String(data.get("email") || "").trim();
    const role = String(data.get("role") || "member");
    if (!email) return fail(400, { memberError: "email required", email });
    const result = await apiFetch(`/api/workspaces/${params.id}/members`, {
      method: "POST",
      body: { email, role },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { memberError: result.error || "add failed", email });
    return { memberAdded: true };
  },
  removeMember: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const userID = String((await request.formData()).get("user_id") || "");
    if (!userID) return fail(400, { memberError: "user_id required" });
    const result = await apiFetch(`/api/workspaces/${params.id}/members/${userID}`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { memberError: result.error || "remove failed" });
    return { memberRemoved: true };
  }
};
export {
  actions,
  load
};
