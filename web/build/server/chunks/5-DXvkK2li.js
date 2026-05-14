import { f as fail, r as redirect } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [
    cardsResult,
    pinsResult,
    flakyResult,
    runPulseResult,
    heatmapResult,
    myQueueResult,
    velocityResult,
    sandboxResult,
    quotaResult
  ] = await Promise.all([
    apiFetch("/api/me/dashboard", { cookie: cookieHeader }),
    apiFetch("/api/me/pinned-tasks", { cookie: cookieHeader }),
    apiFetch("/api/me/flaky-tasks", { cookie: cookieHeader }),
    apiFetch("/api/widgets/dashboard/run-pulse", { cookie: cookieHeader }),
    apiFetch("/api/widgets/dashboard/failure-heatmap?days=7", { cookie: cookieHeader }),
    apiFetch("/api/widgets/dashboard/my-queue", { cookie: cookieHeader }),
    apiFetch("/api/widgets/dashboard/sprint-velocity", { cookie: cookieHeader }),
    apiFetch("/api/widgets/dashboard/sandbox-status", { cookie: cookieHeader }),
    apiFetch("/api/widgets/dashboard/quota-meter", { cookie: cookieHeader })
  ]);
  const defaultQueue = { pending_approvals: [], expiring_sandboxes: [], unread_mentions: 0 };
  const defaultVelocity = {
    this_week: { runs: 0, passed: 0, failed: 0, avg_duration_ms: 0 },
    last_week: { runs: 0, passed: 0, failed: 0, avg_duration_ms: 0 },
    deltas: { runs: 0, passed: 0, failed: 0, avg_duration_ms: 0 }
  };
  return {
    cards: cardsResult.ok ? cardsResult.data ?? [] : [],
    pinned: pinsResult.ok ? pinsResult.data ?? [] : [],
    flaky: flakyResult.ok ? flakyResult.data ?? [] : [],
    error: cardsResult.ok ? null : cardsResult.error || "failed to load dashboard",
    runPulse: runPulseResult.ok ? runPulseResult.data?.runs ?? [] : [],
    heatmap: heatmapResult.ok ? heatmapResult.data?.buckets ?? [] : [],
    myQueue: myQueueResult.ok ? myQueueResult.data ?? defaultQueue : defaultQueue,
    velocity: velocityResult.ok ? velocityResult.data ?? defaultVelocity : defaultVelocity,
    sandboxes: sandboxResult.ok ? sandboxResult.data?.sandboxes ?? [] : [],
    quota: quotaResult.ok ? quotaResult.data?.workspaces ?? [] : []
  };
};
const actions = {
  runPinned: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    if (!taskID) return fail(400, { error: "task_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/tasks/${taskID}/runs`, {
      method: "POST",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "run failed to start" });
    }
    throw redirect(303, `/runs/${result.data.id}`);
  },
  unpin: async ({ request, cookies }) => {
    const data = await request.formData();
    const taskID = String(data.get("task_id") || "");
    if (!taskID) return fail(400, { error: "task_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    await apiFetch(`/api/tasks/${taskID}/pin`, {
      method: "DELETE",
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    return { unpinned: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 5;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte-BoCR3mDg.js')).default;
const server_id = "src/routes/dashboard/+page.server.ts";
const imports = ["_app/immutable/nodes/5.CTY-uoY1.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/BoB1CPyg.js","_app/immutable/chunks/CT9QbQKf.js","_app/immutable/chunks/C-j7aE8P.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/Cz5G4ddd.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/PageHeader.Ct_CmObI.css","_app/immutable/assets/Panel.qbNo31SY.css","_app/immutable/assets/StatusPill.DBhu1QxG.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/5.CXXyi9n1.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=5-DXvkK2li.js.map
