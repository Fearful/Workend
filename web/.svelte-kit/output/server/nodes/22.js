import * as server from '../entries/pages/settings/notifications/_page.server.ts.js';

export const index = 22;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/settings/notifications/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/settings/notifications/+page.server.ts";
export const imports = ["_app/immutable/nodes/22.oa3IdDB1.js","_app/immutable/chunks/C4JmIpy3.js","_app/immutable/chunks/apxFJXnE.js","_app/immutable/chunks/CvynrFVy.js","_app/immutable/chunks/CvI2x4sI.js","_app/immutable/chunks/Bn0VCKPQ.js","_app/immutable/chunks/cLUGwKvc.js","_app/immutable/chunks/VBdYAuAB.js","_app/immutable/chunks/DC4q0MGt.js","_app/immutable/chunks/Dx6EV4QR.js","_app/immutable/chunks/BZxKEKJE.js","_app/immutable/chunks/ChBkQGhY.js","_app/immutable/chunks/BiQCKUjL.js","_app/immutable/chunks/BO09PiA2.js","_app/immutable/chunks/Di6J7QjO.js","_app/immutable/chunks/Cxf1vVyU.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.Dd2v_-2B.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/22.BAs25zcE.css"];
export const fonts = [];
