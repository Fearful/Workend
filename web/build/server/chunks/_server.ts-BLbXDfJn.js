import { j as json } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const STATE_COOKIE_PREFIX = "workend_oauth_state_";
const POST = async ({ params, cookies, request }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const result = await apiFetch(`/api/auth/${params.provider}/start`, {
    method: "POST",
    cookie: cookie ? `${SESSION_COOKIE}=${cookie}` : void 0,
    userAgent: request.headers.get("user-agent") || ""
  });
  if (!result.ok || !result.data) {
    return json({ error: result.error || "failed" }, { status: result.status });
  }
  if (result.setCookie) {
    const expected = STATE_COOKIE_PREFIX + params.provider;
    const m = result.setCookie.match(new RegExp(`^${expected}=([^;]+)`));
    if (m) {
      cookies.set(expected, m[1], {
        path: "/",
        httpOnly: true,
        sameSite: "lax",
        maxAge: 600
      });
    }
  }
  return json({ url: result.data.url });
};

export { POST };
//# sourceMappingURL=_server.ts-BLbXDfJn.js.map
