import * as server from '../entries/pages/runs/_id_/compare/_other_/_page.server.ts.js';

export const index = 25;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/runs/_id_/compare/_other_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/runs/[id]/compare/[other]/+page.server.ts";
export const imports = ["_app/immutable/nodes/25.BOSskJoq.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/Bq1FU2Pv.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/25.hCPyTSHD.css"];
export const fonts = [];
