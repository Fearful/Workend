import * as server from '../entries/pages/projects/_id_/monorepo/_page.server.ts.js';

export const index = 18;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/projects/_id_/monorepo/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/projects/[id]/monorepo/+page.server.ts";
export const imports = ["_app/immutable/nodes/18.BsX2qSBf.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/wpI5rr64.js"];
export const stylesheets = ["_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/18.Dnd9vBfP.css","_app/immutable/assets/Panel.qbNo31SY.css"];
export const fonts = [];
