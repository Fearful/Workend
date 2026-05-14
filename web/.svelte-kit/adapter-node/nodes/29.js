import * as server from '../entries/pages/settings/notifications/_page.server.ts.js';

export const index = 29;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/settings/notifications/_page.svelte.js')).default;
export { server };
export const server_id = "src/routes/settings/notifications/+page.server.ts";
export const imports = ["_app/immutable/nodes/29.C0yOeIdY.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/Dv-DGh57.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js"];
export const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/29.jxliwq_n.css"];
export const fonts = [];
