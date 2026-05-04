import { error, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const runResult = await apiFetch(`/api/runs/${params.id}`, { cookie: cookieHeader });
  if (runResult.status === 404) throw error(404, "run not found");
  if (!runResult.ok || !runResult.data) throw error(500, runResult.error || "failed to load");
  const recentResult = await apiFetch(`/api/projects/${runResult.data.project_id}/runs`, { cookie: cookieHeader });
  const recentRuns = recentResult.ok && recentResult.data ? recentResult.data.filter((r) => r.task_id === runResult.data.task_id && r.id !== params.id).slice(0, 10) : [];
  const commentsResult = await apiFetch(`/api/runs/${params.id}/comments`, { cookie: cookieHeader });
  const comments = commentsResult.ok ? commentsResult.data ?? [] : [];
  let log = "";
  const isTerminal = runResult.data.status === "succeeded" || runResult.data.status === "failed" || runResult.data.status === "cancelled";
  if (isTerminal) {
    try {
      const apiURL = process.env.WORKEND_API_URL || "http://api:8080";
      const r = await fetch(`${apiURL}/api/runs/${params.id}/log`, {
        headers: cookieHeader ? { Cookie: cookieHeader } : {}
      });
      if (r.ok) log = await r.text();
    } catch {
    }
  }
  return { run: runResult.data, log, recentRuns, comments };
};
const actions = {
  rerun: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const original = await apiFetch(`/api/runs/${params.id}`, { cookie: cookieHeader });
    if (!original.ok || !original.data) throw error(500, "failed to load original run");
    const newRun = await apiFetch(`/api/tasks/${original.data.task_id}/runs`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!newRun.ok || !newRun.data) throw error(newRun.status, newRun.error || "rerun failed");
    throw redirect(303, `/runs/${newRun.data.id}`);
  },
  comment: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const body = String(data.get("body") || "").trim();
    if (!body) return { commentError: "comment body required", commentDraft: body };
    const result = await apiFetch(`/api/runs/${params.id}/comments`, {
      method: "POST",
      body: { body },
      cookie: cookieHeader
    });
    if (!result.ok) {
      return { commentError: result.error || "failed to post comment", commentDraft: body };
    }
    return { commented: true };
  },
  deleteComment: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const id = String(data.get("id") || "");
    if (!id) return { commentError: "id required" };
    const result = await apiFetch(`/api/runs/${params.id}/comments/${id}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) {
      return { commentError: result.error || "failed to delete" };
    }
    return { commentDeleted: true };
  },
  approve: async ({ params, request, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const data = await request.formData();
    const approved = String(data.get("approved") || "") === "true";
    const note = String(data.get("note") || "");
    const r = await apiFetch(`/api/runs/${params.id}/approve`, {
      method: "POST",
      body: { approved, note },
      cookie: cookieHeader
    });
    if (!r.ok) return { approvalError: r.error || "failed" };
    return { approvalDecided: approved };
  },
  rerunPinned: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const original = await apiFetch(`/api/runs/${params.id}`, { cookie: cookieHeader });
    if (!original.ok || !original.data) throw error(500, "failed to load original run");
    if (!original.data.commit_sha) throw error(400, "original run has no recorded commit");
    const newRun = await apiFetch(`/api/tasks/${original.data.task_id}/runs`, {
      method: "POST",
      body: { at_commit: original.data.commit_sha },
      cookie: cookieHeader
    });
    if (!newRun.ok || !newRun.data) throw error(newRun.status, newRun.error || "rerun failed");
    throw redirect(303, `/runs/${newRun.data.id}`);
  }
};
export {
  actions,
  load
};
