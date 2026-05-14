import { redirect, error } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const prResult = await apiFetch(`/api/pipeline-runs/${params.id}`, { cookie: cookieHeader });
  if (prResult.status === 404) throw error(404, "pipeline run not found");
  if (!prResult.ok || !prResult.data) throw error(500, prResult.error || "failed to load");
  let pipeline = null;
  return { pipelineRun: prResult.data, pipeline };
};
export {
  load
};
