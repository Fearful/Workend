import * as server from '../entries/pages/workspaces/_id_/_page.server.ts.js';

export const index = 24;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/workspaces/_id_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/workspaces/[id]/+page.server.ts";
export const imports = ["_app/immutable/nodes/24.BYXFyrXT.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/VBdYAuAB.js","_app/immutable/chunks/BvqIPo9c.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/BO09PiA2.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/24.CwuqKXKE.css","_app/immutable/assets/EmptyState.ByauYn1J.css"];
export const fonts = [];
