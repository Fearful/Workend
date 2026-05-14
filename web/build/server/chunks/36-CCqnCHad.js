import { f as fail, r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [wsResp, secretsResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/secrets`, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  return {
    workspace: wsResp.data,
    secrets: secretsResp.ok ? secretsResp.data ?? [] : [],
    secretsError: secretsResp.ok ? null : secretsResp.error || "failed to load secrets"
  };
};
const actions = {
  createSecret: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    const value = String(data.get("value") || "");
    if (!name || !value) return fail(400, { error: "Name and value are required", name });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/secrets`, {
      method: "POST",
      body: { name, value },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "Failed to create secret", name });
    return { created: true };
  },
  updateSecret: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const secretId = String(data.get("secret_id") || "");
    const value = String(data.get("value") || "");
    if (!secretId || !value) return fail(400, { error: "Secret ID and value are required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/secrets/${secretId}`, {
      method: "PUT",
      body: { value },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "Failed to update secret" });
    return { updated: true };
  },
  deleteSecret: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const secretId = String(data.get("secret_id") || "");
    if (!secretId) return fail(400, { error: "Secret ID required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/secrets/${secretId}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "Failed to delete secret" });
    return { deleted: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 36;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BcK_Prn0.js')).default;
const server_id = "src/routes/workspaces/[id]/secrets/+page.server.ts";
const imports = ["_app/immutable/nodes/36.B6jbsTVh.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/36.BZGUCT9h.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=36-CCqnCHad.js.map
