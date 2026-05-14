import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const previewsResult = await apiFetch(
    `/api/projects/${params.id}/previews`,
    { cookie: cookieHeader }
  );
  return {
    previews: previewsResult.ok ? previewsResult.data ?? [] : []
  };
};
const actions = {
  createPreview: async ({ params, request, cookies }) => {
    const formData = await request.formData();
    const branch = String(formData.get("branch") || "").trim();
    if (!branch) return fail(400, { error: "branch required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/projects/${params.id}/previews`, {
      method: "POST",
      body: { branch },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "preview creation failed" });
    return { previewCreated: true };
  },
  deletePreview: async ({ request, cookies }) => {
    const formData = await request.formData();
    const previewID = String(formData.get("preview_id") || "");
    if (!previewID) return fail(400, { error: "preview_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/previews/${previewID}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { deleted: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 20;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BFwwquLC.js')).default;
const server_id = "src/routes/projects/[id]/previews/+page.server.ts";
const imports = ["_app/immutable/nodes/20.BiMzlWCF.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/20.ucwMtO09.css","_app/immutable/assets/Panel.qbNo31SY.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=20-Dk4Qb6as.js.map
