const API_URL = process.env.WORKEND_API_URL || "http://api:8080";
async function apiFetch(path, opts = {}) {
  const headers = {
    "Content-Type": "application/json"
  };
  if (opts.cookie) headers["Cookie"] = opts.cookie;
  if (opts.userAgent) headers["User-Agent"] = opts.userAgent;
  const res = await fetch(`${API_URL}${path}`, {
    method: opts.method || "GET",
    headers,
    body: opts.body ? JSON.stringify(opts.body) : void 0
  });
  const setCookie = res.headers.get("set-cookie") || void 0;
  const contentType = res.headers.get("content-type") || "";
  if (!res.ok) {
    const text = await res.text();
    return { ok: false, status: res.status, error: text.trim(), setCookie };
  }
  if (res.status === 204 || !contentType.includes("application/json")) {
    return { ok: true, status: res.status, setCookie };
  }
  const data = await res.json();
  return { ok: true, status: res.status, data, setCookie };
}
function parseSetCookie(header) {
  const parts = header.split(";").map((p) => p.trim());
  const [nameValue, ...attrs] = parts;
  const eq = nameValue.indexOf("=");
  if (eq < 0) return null;
  const name = nameValue.slice(0, eq);
  const value = nameValue.slice(eq + 1);
  const options = {
    path: "/",
    httpOnly: false,
    secure: false,
    sameSite: "lax",
    expires: void 0,
    maxAge: void 0
  };
  for (const attr of attrs) {
    const lower = attr.toLowerCase();
    if (lower === "httponly") options.httpOnly = true;
    else if (lower === "secure") options.secure = true;
    else if (lower.startsWith("path=")) options.path = attr.slice(5);
    else if (lower.startsWith("expires=")) options.expires = new Date(attr.slice(8));
    else if (lower.startsWith("max-age=")) options.maxAge = parseInt(attr.slice(8), 10);
    else if (lower.startsWith("samesite=")) {
      const v = attr.slice(9).toLowerCase();
      if (v === "lax" || v === "strict" || v === "none") options.sameSite = v;
    }
  }
  return { name, value, options };
}
export {
  apiFetch as a,
  parseSetCookie as p
};
