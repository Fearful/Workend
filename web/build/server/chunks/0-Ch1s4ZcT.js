import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  let unreadMentions = 0;
  if (locals.user) {
    const cookie = cookies.get(SESSION_COOKIE);
    const r = await apiFetch("/api/me/mentions/count", {
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (r.ok && r.data) unreadMentions = r.data.unread;
  }
  return { user: locals.user, unreadMentions };
};

var _layout_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 0;
let component_cache;
const component = async () => component_cache ??= (await import('./_layout.svelte-DZ0AYxuW.js')).default;
const server_id = "src/routes/+layout.server.ts";
const imports = ["_app/immutable/nodes/0.zlB7ghG4.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/DuSA90i8.js","_app/immutable/chunks/DdjgLl7P.js","_app/immutable/chunks/CxpEhr-0.js"];
const stylesheets = ["_app/immutable/assets/0.BkaopCgQ.css"];
const fonts = [];

export { component, fonts, imports, index, _layout_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=0-Ch1s4ZcT.js.map
