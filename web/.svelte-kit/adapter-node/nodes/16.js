import * as server from '../entries/pages/projects/_id_/runs/_page.server.ts.js';

export const index = 16;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/projects/_id_/runs/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/projects/[id]/runs/+page.server.ts";
export const imports = ["_app/immutable/nodes/16.CNoxma9g.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/CgWWeIEW.js","_app/immutable/chunks/Dd7pVJ5Y.js","_app/immutable/chunks/059uK4jn.js","_app/immutable/chunks/8aH_EF2d.js","_app/immutable/chunks/Di6J7QjO.js"];
export const stylesheets = ["_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/16.BCsy4Npt.css"];
export const fonts = [];
