import * as server from '../entries/pages/workspaces/_id_/projects/new/_page.server.ts.js';

export const index = 25;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/workspaces/_id_/projects/new/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/workspaces/[id]/projects/new/+page.server.ts";
export const imports = ["_app/immutable/nodes/25.C6SWb2eB.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/Dmb8Bvf1.js","_app/immutable/chunks/BsbC8Uq8.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/VBdYAuAB.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/BiQCKUjL.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/25.Bns4E-Mf.css"];
export const fonts = [];
