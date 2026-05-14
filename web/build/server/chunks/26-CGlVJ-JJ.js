import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const result = await apiFetch("/api/me/sandboxes", { cookie: cookieHeader });
  const sandboxes = result.ok ? result.data ?? [] : [];
  return { sandboxes };
};
const actions = {
  create: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const project_id = String(fd.get("project_id") || "").trim();
    const branch = String(fd.get("branch") || "").trim() || void 0;
    const expiresRaw = String(fd.get("expires_hours") || "").trim();
    if (!project_id) return fail(400, { error: "project_id is required" });
    const expires_hours = expiresRaw ? parseInt(expiresRaw, 10) : void 0;
    if (expires_hours !== void 0 && (!Number.isFinite(expires_hours) || expires_hours < 1 || expires_hours > 72)) {
      return fail(400, { error: "expires_hours must be 1-72" });
    }
    const r = await apiFetch("/api/sandboxes", {
      method: "POST",
      body: { project_id, branch, expires_hours },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to create sandbox" });
    if (r.data?.id) throw redirect(303, `/sandboxes/${r.data.id}`);
    return { created: true };
  },
  destroy: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const id = String(fd.get("id") || "");
    if (!id) return fail(400, { error: "id required" });
    const r = await apiFetch(`/api/sandboxes/${id}/destroy`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to destroy sandbox" });
    return { destroyed: true };
  },
  extend: async ({ request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const fd = await request.formData();
    const id = String(fd.get("id") || "");
    const hoursRaw = String(fd.get("hours") || "").trim();
    if (!id) return fail(400, { error: "id required" });
    const hours = parseInt(hoursRaw, 10);
    if (!Number.isFinite(hours) || hours < 1 || hours > 72) {
      return fail(400, { error: "hours must be 1-72" });
    }
    const r = await apiFetch(`/api/sandboxes/${id}/extend`, {
      method: "POST",
      body: { hours },
      cookie: cookieHeader
    });
    if (!r.ok) return fail(r.status, { error: r.error || "failed to extend sandbox" });
    return { extended: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 26;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-Bf4Fz7Hl.js')).default;
const server_id = "src/routes/sandboxes/+page.server.ts";
const imports = ["_app/immutable/nodes/26.Dowi1yGT.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/wpI5rr64.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/26.CzX2RyZj.css","_app/immutable/assets/Panel.qbNo31SY.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=26-CGlVJ-JJ.js.map
