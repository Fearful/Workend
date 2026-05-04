import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const branchesResult = await apiFetch(`/api/projects/${params.id}/branches`, { cookie: cookieHeader });
  return {
    branchesSource: branchesResult.ok ? branchesResult.data?.source ?? "ls-remote" : "ls-remote",
    branches: branchesResult.ok ? branchesResult.data?.branches ?? [] : [],
    branchesError: branchesResult.ok ? null : branchesResult.error || "failed to list branches",
    canCreatePR: branchesResult.ok && branchesResult.data?.source === "provider"
  };
};
const actions = {
  switch: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const name = String(data.get("name") || "").trim();
    if (!name) return fail(400, { error: "branch name required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/projects/${params.id}/branch`, {
      method: "POST",
      body: { name },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok) return fail(result.status, { error: result.error || "switch failed" });
    throw redirect(303, `/projects/${params.id}`);
  },
  createPR: async ({ request, params, cookies }) => {
    const data = await request.formData();
    const source = String(data.get("source") || "").trim();
    const target = String(data.get("target") || "").trim();
    const title = String(data.get("title") || "").trim();
    const body = String(data.get("body") || "");
    if (!source || !target) return fail(400, { error: "source and target required" });
    if (source === target) return fail(400, { error: "source and target must differ" });
    const cookie = cookies.get(SESSION_COOKIE);
    const result = await apiFetch(`/api/projects/${params.id}/pull-requests`, {
      method: "POST",
      body: { source, target, title, body },
      cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
    });
    if (!result.ok || !result.data) {
      return fail(result.status, { error: result.error || "PR creation failed" });
    }
    return { prURL: result.data.url, prNumber: result.data.number };
  }
};
export {
  actions,
  load
};
