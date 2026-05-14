import * as server from '../entries/pages/workspaces/_id_/_page.server.ts.js';

export const index = 31;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/workspaces/_id_/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/workspaces/[id]/+page.server.ts";
export const imports = ["_app/immutable/nodes/31.k7Zg7_2p.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/31.BQ9SYp--.css"];
export const fonts = [];
