import { j as json } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

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

export { POST };
//# sourceMappingURL=_server.ts-DVQY26Di.js.map
