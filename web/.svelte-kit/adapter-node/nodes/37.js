import * as server from '../entries/pages/workspaces/new/_page.server.ts.js';

export const index = 37;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/workspaces/new/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/workspaces/new/+page.server.ts";
export const imports = ["_app/immutable/nodes/37.Dk1XDMx2.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js"];
export const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/37._Y1PPEyG.css"];
export const fonts = [];
