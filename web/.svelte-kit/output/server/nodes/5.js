import * as server from '../entries/pages/dashboard/_page.server.ts.js';

export const index = 5;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/dashboard/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/dashboard/+page.server.ts";
export const imports = ["_app/immutable/nodes/5.CTY-uoY1.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
export const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/5.CXXyi9n1.css"];
export const fonts = [];
