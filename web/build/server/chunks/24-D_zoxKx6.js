import { f as fail, e as error, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

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

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 24;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-DPBbFC3c.js')).default;
const server_id = "src/routes/workspaces/[id]/+page.server.ts";
const imports = ["_app/immutable/nodes/24.BYXFyrXT.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/VBdYAuAB.js","_app/immutable/chunks/BvqIPo9c.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/BO09PiA2.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/24.CwuqKXKE.css","_app/immutable/assets/EmptyState.ByauYn1J.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=24-D_zoxKx6.js.map
