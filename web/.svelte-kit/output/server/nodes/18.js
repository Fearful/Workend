import * as server from '../entries/pages/projects/_id_/trends/_page.server.ts.js';

export const index = 18;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/projects/_id_/trends/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/projects/[id]/trends/+page.server.ts";
export const imports = ["_app/immutable/nodes/18.8OMNqkNq.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/CYuiovvJ.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/Di6J7QjO.js"];
export const stylesheets = ["_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/18.wycGJu-N.css"];
export const fonts = [];
