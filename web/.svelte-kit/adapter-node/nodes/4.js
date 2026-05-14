import * as server from '../entries/pages/admin/_page.server.ts.js';

export const index = 4;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/admin/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/admin/+page.server.ts";
export const imports = ["_app/immutable/nodes/4.CtiVQOGN.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js"];
export const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/4.jiiUQkUp.css"];
export const fonts = [];
