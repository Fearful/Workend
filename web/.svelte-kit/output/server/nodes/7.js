import * as server from '../entries/pages/issues/_id_/_page.server.ts.js';

export const index = 7;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/issues/_id_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/issues/[id]/+page.server.ts";
export const imports = ["_app/immutable/nodes/7.BYJEXiFz.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/Dmb8Bvf1.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/Cxf1vVyU.js"];
export const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/7.Bh465XZN.css"];
export const fonts = [];
