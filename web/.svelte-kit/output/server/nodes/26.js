import * as server from '../entries/pages/workspaces/new/_page.server.ts.js';

export const index = 26;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/workspaces/new/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/workspaces/new/+page.server.ts";
export const imports = ["_app/immutable/nodes/26.DN7zoh-Z.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/Cxf1vVyU.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js"];
export const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/26._Y1PPEyG.css"];
export const fonts = [];
