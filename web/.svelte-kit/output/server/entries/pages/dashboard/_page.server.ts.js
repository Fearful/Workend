import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../chunks/api.js";
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
export {
  actions,
  load
};
