import * as server from '../entries/pages/runs/_id_/compare/_other_/_page.server.ts.js';

export const index = 20;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/runs/_id_/compare/_other_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/runs/[id]/compare/[other]/+page.server.ts";
export const imports = ["_app/immutable/nodes/20.CjUXcWpS.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/VBdYAuAB.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/BvqIPo9c.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/BiQCKUjL.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/20.hCPyTSHD.css"];
export const fonts = [];
