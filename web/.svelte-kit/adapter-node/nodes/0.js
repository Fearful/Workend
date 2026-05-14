import * as server from '../entries/pages/_layout.server.ts.js';

export const index = 0;
let component_cache;
export const component = async () => component_cache ??= (await import('../entries/pages/_layout.svelte.js')).default;
export { server };
export const server_id = "src/routes/+layout.server.ts";
export const imports = ["_app/immutable/nodes/0.zlB7ghG4.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/CxpEhr-0.js"];
export const stylesheets = ["_app/immutable/assets/0.BkaopCgQ.css"];
export const fonts = [];
