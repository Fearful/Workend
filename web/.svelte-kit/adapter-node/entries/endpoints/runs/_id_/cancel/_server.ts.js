import { json } from "@sveltejs/kit";
import { a as apiFetch } from "../../../../../chunks/api.js";
const SESSION_COOKIE = "workend_session";
const POST = async ({ params, cookies }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch(`/api/runs/${params.id}/cancel`, {
    method: "POST",
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0
  });
  if (!result.ok) return json({ error: result.error || "cancel failed" }, { status: result.status });
  return json({ ok: true });
};
export {
  POST
};
