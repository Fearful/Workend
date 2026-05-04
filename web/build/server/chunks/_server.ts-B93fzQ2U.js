import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const POST = async ({ params, request, cookies }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const data = await request.formData();
  const to = String(data.get("to_column") || "").trim();
  if (!to) return new Response("to_column required", { status: 400 });
  const r = await apiFetch(`/api/issues/${params.id}/move`, {
    method: "PATCH",
    body: { to_column: to },
    cookie: cookieHeader
  });
  return new Response(r.error ?? "", { status: r.ok ? 204 : r.status });
};

export { POST };
//# sourceMappingURL=_server.ts-B93fzQ2U.js.map
