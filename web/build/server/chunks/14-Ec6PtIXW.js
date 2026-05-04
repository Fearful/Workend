import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const branchesResult = await apiFetch(`/api/projects/${params.id}/branches`, { cookie: cookieHeader });
  return {
    branchesSource: branchesResult.ok ? branchesResult.data?.source ?? "ls-remote" : "ls-remote",
    branches: branchesResult.ok ? branchesResult.data?.branches ?? [] : [],
    branchesError: branchesResult.ok ? null : branchesResult.error || "failed to list branches",
    canCreatePR: branchesResult.ok && branchesResult.data?.source === "provider"
  };
};
const actions = {
  switch: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    if (!name) return fail(400, { error: "branch name required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/projects/${params.id}/branch`, {
      method: "POST",
      body: { name },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "switch failed" });
    throw redirect(303, `/projects/${params.id}`);
  },
  createPR: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const source = String(data.get("source") || "").trim();
    const target = String(data.get("target") || "").trim();
    const title = String(data.get("title") || "").trim();
    const body = String(data.get("body") || "");
    if (!source || !target) return fail(400, { error: "source and target required" });
    if (source === target) return fail(400, { error: "source and target must differ" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/projects/${params.id}/pull-requests`, {
      method: "POST",
      body: { source, target, title, body },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "PR creation failed" });
    }
    return { prURL: result.data.url, prNumber: result.data.number };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 14;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-CrMTI2KJ.js')).default;
const server_id = "src/routes/projects/[id]/branches/+page.server.ts";
const imports = ["_app/immutable/nodes/14.BHHQsvj7.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/Dmb8Bvf1.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/m1uqkJwC.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/Di6J7QjO.js"];
const stylesheets = ["_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/14.BSFcvH9A.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=14-Ec6PtIXW.js.map
