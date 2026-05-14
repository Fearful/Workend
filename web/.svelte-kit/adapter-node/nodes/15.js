import * as server from '../entries/pages/projects/_id_/board/_page.server.ts.js';

export const index = 15;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/projects/_id_/board/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/projects/[id]/board/+page.server.ts";
export const imports = ["_app/immutable/nodes/15.DCfAA06p.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/B2UVO_nM.js"];
export const stylesheets = ["_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/15.Btu-YQeZ.css"];
export const fonts = [];
