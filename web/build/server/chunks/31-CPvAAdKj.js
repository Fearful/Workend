import { f as fail, e as error, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const wid = params.id;
  const [
    wsResp,
    projResp,
    membersResp,
    activityResp,
    issuesResp,
    gridResp,
    teamResp,
    depsResp,
    secretsSumResp,
    pipeResp,
    sandboxResp
  ] = await Promise.all([
    apiFetch(`/api/workspaces/${wid}`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${wid}/projects`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${wid}/members`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${wid}/activity?limit=30`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${wid}/recent-issues`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/workspace/${wid}/project-grid`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/workspace/${wid}/team-presence`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/workspace/${wid}/dependency-overview`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/workspace/${wid}/secrets-summary`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/workspace/${wid}/pipeline-board`, { cookie: cookieHeader }),
    apiFetch(`/api/widgets/workspace/${wid}/sandbox-preview-rack`, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  return {
    workspace: wsResp.data,
    projects: projResp.ok ? projResp.data ?? [] : [],
    projectsError: projResp.ok ? null : projResp.error || "failed to load projects",
    members: membersResp.ok ? membersResp.data ?? [] : [],
    activity: activityResp.ok ? activityResp.data ?? [] : [],
    recentIssues: issuesResp.ok ? issuesResp.data ?? [] : [],
    projectGrid: gridResp.ok ? gridResp.data?.projects ?? [] : [],
    teamPresence: teamResp.ok ? teamResp.data?.members ?? [] : [],
    depOverview: depsResp.ok ? depsResp.data ?? null : null,
    secretsSummary: secretsSumResp.ok ? secretsSumResp.data ?? null : null,
    pipelineBoard: pipeResp.ok ? pipeResp.data ?? null : null,
    sandboxItems: sandboxResp.ok ? sandboxResp.data?.items ?? [] : []
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

const index = 31;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-Bjl7oeQ_.js')).default;
const server_id = "src/routes/workspaces/[id]/+page.server.ts";
const imports = ["_app/immutable/nodes/31.k7Zg7_2p.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/31.BQ9SYp--.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=31-CPvAAdKj.js.map
