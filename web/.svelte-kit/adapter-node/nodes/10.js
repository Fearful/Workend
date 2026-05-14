import * as server from '../entries/pages/mentions/_page.server.ts.js';

export const index = 10;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/mentions/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/mentions/+page.server.ts";
export const imports = ["_app/immutable/nodes/10.JitZ3q-T.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/YSzAdjSj.js"];
export const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/10.BxgXEenA.css"];
export const fonts = [];
