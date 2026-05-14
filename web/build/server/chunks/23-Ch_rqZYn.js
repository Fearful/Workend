import { r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, url, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const taskID = url.searchParams.get("task_id") || "";
  const daysRaw = url.searchParams.get("days") || "30";
  const days = Math.max(1, Math.min(365, parseInt(daysRaw, 10) || 30));
  const trendsPath = taskID ? `/api/projects/${params.id}/trends?days=${days}&task_id=${encodeURIComponent(taskID)}` : `/api/projects/${params.id}/trends?days=${days}`;
  const [tasksResult, trendsResult, activityResult] = await Promise.all([
    apiFetch(`/api/projects/${params.id}/tasks`, { cookie: cookieHeader }),
    apiFetch(trendsPath, { cookie: cookieHeader }),
    apiFetch(`/api/projects/${params.id}/activity?days=365`, { cookie: cookieHeader })
  ]);
  return {
    tasks: tasksResult.ok ? tasksResult.data ?? [] : [],
    trends: trendsResult.ok ? trendsResult.data ?? { days, points: [] } : { days, points: [] },
    activity: activityResult.ok ? activityResult.data ?? { days: 365, cells: [] } : { days: 365, cells: [] },
    filterTaskID: taskID,
    days
  };
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  load: load
});

const index = 23;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-DmVfX4AL.js')).default;
const server_id = "src/routes/projects/[id]/trends/+page.server.ts";
const imports = ["_app/immutable/nodes/23.CWwcOGGD.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Dv-DGh57.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/B8IgtE21.js"];
const stylesheets = ["_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/23.C8Tt1xp9.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=23-Ch_rqZYn.js.map
