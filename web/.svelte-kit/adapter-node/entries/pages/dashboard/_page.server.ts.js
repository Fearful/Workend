import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [cardsResult, pinsResult, flakyResult] = await Promise.all([
    apiFetch("/api/me/dashboard", { cookie: cookieHeader }),
    apiFetch("/api/me/pinned-tasks", { cookie: cookieHeader }),
    apiFetch("/api/me/flaky-tasks", { cookie: cookieHeader })
  ]);
  return {
    cards: cardsResult.ok ? cardsResult.data ?? [] : [],
    pinned: pinsResult.ok ? pinsResult.data ?? [] : [],
    flaky: flakyResult.ok ? flakyResult.data ?? [] : [],
    error: cardsResult.ok ? null : cardsResult.error || "failed to load dashboard"
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
