import { redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const projResult = await apiFetch(`/api/projects/${params.id}`, { cookie: cookieHeader });
  if (projResult.status === 404) throw error(404, "project not found");
  if (!projResult.ok || !projResult.data) throw error(500, projResult.error || "failed to load");
  const wsResult = await apiFetch(
    `/api/workspaces/${projResult.data.workspace_id}`,
    { cookie: cookieHeader }
  );
  return {
    project: projResult.data,
    workspace: wsResult.ok ? wsResult.data ?? null : null
  };
};
export {
  load
};
