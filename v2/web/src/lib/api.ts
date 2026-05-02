/**
 * Server-side API client. Always called from .server.ts files; never imported
 * by .svelte components directly. Forwards the browser's cookie to the API
 * so the API can identify the user.
 */
const API_URL = process.env.WORKEND_API_URL || 'http://api:8080';

export interface ApiResult<T> {
  ok: boolean;
  status: number;
  data?: T;
  error?: string;
  setCookie?: string;
}

export async function apiFetch<T = unknown>(
  path: string,
  opts: {
    method?: string;
    body?: unknown;
    cookie?: string;
    userAgent?: string;
  } = {}
): Promise<ApiResult<T>> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json'
  };
  if (opts.cookie) headers['Cookie'] = opts.cookie;
  if (opts.userAgent) headers['User-Agent'] = opts.userAgent;

  const res = await fetch(`${API_URL}${path}`, {
    method: opts.method || 'GET',
    headers,
    body: opts.body ? JSON.stringify(opts.body) : undefined
  });

  const setCookie = res.headers.get('set-cookie') || undefined;
  const contentType = res.headers.get('content-type') || '';

  if (!res.ok) {
    const text = await res.text();
    return { ok: false, status: res.status, error: text.trim(), setCookie };
  }

  if (res.status === 204 || !contentType.includes('application/json')) {
    return { ok: true, status: res.status, setCookie };
  }

  const data = (await res.json()) as T;
  return { ok: true, status: res.status, data, setCookie };
}

/**
 * Parse a Set-Cookie header into the pieces SvelteKit's `cookies.set` needs.
 * The Go API issues a single workend_session cookie on auth; we just need
 * value, expires, and the security flags.
 */
export function parseSetCookie(header: string): {
  name: string;
  value: string;
  options: {
    path: string;
    httpOnly: boolean;
    secure: boolean;
    sameSite: 'lax' | 'strict' | 'none';
    expires?: Date;
    maxAge?: number;
  };
} | null {
  const parts = header.split(';').map((p) => p.trim());
  const [nameValue, ...attrs] = parts;
  const eq = nameValue.indexOf('=');
  if (eq < 0) return null;
  const name = nameValue.slice(0, eq);
  const value = nameValue.slice(eq + 1);

  const options = {
    path: '/',
    httpOnly: false,
    secure: false,
    sameSite: 'lax' as 'lax' | 'strict' | 'none',
    expires: undefined as Date | undefined,
    maxAge: undefined as number | undefined
  };

  for (const attr of attrs) {
    const lower = attr.toLowerCase();
    if (lower === 'httponly') options.httpOnly = true;
    else if (lower === 'secure') options.secure = true;
    else if (lower.startsWith('path=')) options.path = attr.slice(5);
    else if (lower.startsWith('expires=')) options.expires = new Date(attr.slice(8));
    else if (lower.startsWith('max-age=')) options.maxAge = parseInt(attr.slice(8), 10);
    else if (lower.startsWith('samesite=')) {
      const v = attr.slice(9).toLowerCase();
      if (v === 'lax' || v === 'strict' || v === 'none') options.sameSite = v;
    }
  }

  return { name, value, options };
}
