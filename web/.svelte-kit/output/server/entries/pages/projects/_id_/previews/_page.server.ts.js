import { fail, redirect } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const previewsResult = await apiFetch(
    `/api/projects/${params.id}/previews`,
    { cookie: cookieHeader }
  );
  return {
    previews: previewsResult.ok ? previewsResult.data ?? [] : []
  };
};
const actions = {
  createPreview: async ({ params, request, cookies }) => {
    const formData = await request.formData();
    const branch = String(formData.get("branch") || "").trim();
    if (!branch) return fail(400, { error: "branch required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/projects/${params.id}/previews`, {
      method: "POST",
      body: { branch },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "preview creation failed" });
    return { previewCreated: true };
  },
  deletePreview: async ({ request, cookies }) => {
    const formData = await request.formData();
    const previewID = String(formData.get("preview_id") || "");
    if (!previewID) return fail(400, { error: "preview_id required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/previews/${previewID}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { error: result.error || "delete failed" });
    return { deleted: true };
  }
};
export {
  actions,
  load
};
