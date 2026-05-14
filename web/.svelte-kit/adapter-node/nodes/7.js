import * as server from '../entries/pages/issues/_id_/_page.server.ts.js';

export const index = 7;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/issues/_id_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/issues/[id]/+page.server.ts";
export const imports = ["_app/immutable/nodes/7.B9lkA62X.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/C5qpB67O.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/wpI5rr64.js"];
export const stylesheets = ["_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/7.Bh465XZN.css"];
export const fonts = [];
