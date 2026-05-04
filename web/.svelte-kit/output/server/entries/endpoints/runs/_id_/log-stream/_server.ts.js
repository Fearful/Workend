const API_URL = process.env.WORKEND_API_URL || "http://api:8080";
const SESSION_COOKIE = "workend_session";
const GET = async ({ params, cookies, request }) => {
  const cookie = cookies.get(SESSION_COOKIE);
  const headers = {
    Accept: "text/event-stream"
  };
  if (cookie) headers["Cookie"] = `${SESSION_COOKIE}=${cookie}`;
  const upstream = await fetch(`${API_URL}/api/runs/${params.id}/log/stream`, {
    headers,
    signal: request.signal
  });
  if (!upstream.ok || !upstream.body) {
    return new Response("upstream error", { status: upstream.status || 502 });
  }
  return new Response(upstream.body, {
    status: 200,
    headers: {
      "Content-Type": "text/event-stream",
      "Cache-Control": "no-cache",
      "Connection": "keep-alive",
      "X-Accel-Buffering": "no"
    }
  });
};
export {
  GET
};
