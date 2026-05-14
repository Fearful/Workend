import * as server from '../entries/pages/projects/_id_/_layout.server.ts.js';

export const index = 2;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/projects/_id_/_layout.svelte.js')).default;
export { server };
export const server_id = "src/routes/projects/[id]/+layout.server.ts";
export const imports = ["_app/immutable/nodes/2.D_-oxuV4.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/CxpEhr-0.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B2UVO_nM.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/wpI5rr64.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/2.DW8Ekywe.css"];
export const fonts = [];
