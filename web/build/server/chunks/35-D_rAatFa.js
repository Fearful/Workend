import { f as fail, r as redirect, e as error } from './index-BkmUvga9.js';
import { a as apiFetch } from './api-chcxdA-p.js';

const SESSION_COOKIE = "workend_session";
const load = async ({ params, locals, cookies }) => {
  if (!locals.user) throw redirect(303, "/login");
  const cookie = cookies.get(SESSION_COOKIE);
  const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
  const [wsResp, rolesResp, membersResp, keysResp] = await Promise.all([
    apiFetch(`/api/workspaces/${params.id}`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/roles`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/members`, { cookie: cookieHeader }),
    apiFetch(`/api/workspaces/${params.id}/signing-keys`, { cookie: cookieHeader })
  ]);
  if (wsResp.status === 404) throw error(404, "workspace not found");
  if (!wsResp.ok || !wsResp.data) throw error(500, wsResp.error || "failed to load workspace");
  return {
    workspace: wsResp.data,
    roles: rolesResp.ok ? rolesResp.data ?? [] : [],
    rolesError: rolesResp.ok ? null : rolesResp.error || "failed to load roles",
    members: membersResp.ok ? membersResp.data ?? [] : [],
    signingKeys: keysResp.ok ? keysResp.data ?? [] : [],
    keysError: keysResp.ok ? null : keysResp.error || "failed to load signing keys"
  };
};
const actions = {
  createRole: async ({ request, params, cookies }) => {
    const fd = await request.formData();
    const name = String(fd.get("name") || "").trim();
    const description = String(fd.get("description") || "").trim();
    const permEntries = fd.getAll("permissions");
    if (!name) return fail(400, { roleError: "Role name is required", roleName: name });
    const permissions = [];
    for (const entry of permEntries) {
      const [resource, action] = String(entry).split(":");
      if (resource && action) permissions.push({ resource, action });
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/roles`, {
      method: "POST",
      body: { name, description, permissions },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { roleError: result.error || "Failed to create role", roleName: name });
    return { roleCreated: true };
  },
  updateRole: async ({ request, cookies }) => {
    const fd = await request.formData();
    const roleId = String(fd.get("role_id") || "");
    const name = String(fd.get("name") || "").trim();
    const description = String(fd.get("description") || "").trim();
    const permEntries = fd.getAll("permissions");
    if (!roleId) return fail(400, { roleError: "Role ID required" });
    const permissions = [];
    for (const entry of permEntries) {
      const [resource, action] = String(entry).split(":");
      if (resource && action) permissions.push({ resource, action });
    }
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const body = {};
    if (name) body.name = name;
    if (description) body.description = description;
    body.permissions = permissions;
    const result = await apiFetch(`/api/roles/${roleId}`, {
      method: "PATCH",
      body,
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { roleError: result.error || "Failed to update role" });
    return { roleUpdated: true };
  },
  deleteRole: async ({ request, cookies }) => {
    const fd = await request.formData();
    const roleId = String(fd.get("role_id") || "");
    if (!roleId) return fail(400, { roleError: "Role ID required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/roles/${roleId}`, {
      method: "DELETE",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { roleError: result.error || "Failed to delete role" });
    return { roleDeleted: true };
  },
  assignRole: async ({ request, params, cookies }) => {
    const fd = await request.formData();
    const userId = String(fd.get("user_id") || "");
    const roleId = String(fd.get("role_id") || "");
    if (!userId || !roleId) return fail(400, { assignError: "User and role are required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/workspaces/${params.id}/members/${userId}/role`, {
      method: "POST",
      body: { role_id: roleId },
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { assignError: result.error || "Failed to assign role" });
    return { roleAssigned: true };
  },
  generateKey: async ({ params, cookies }) => {
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(
      `/api/workspaces/${params.id}/signing-keys`,
      { method: "POST", cookie: cookieHeader }
    );
    if (!result.ok) return fail(result.status, { keyError: result.error || "Failed to generate key" });
    return { keyGenerated: true, privateKey: result.data?.private_key || "" };
  },
  revokeKey: async ({ request, cookies }) => {
    const fd = await request.formData();
    const keyId = String(fd.get("key_id") || "");
    if (!keyId) return fail(400, { keyError: "Key ID required" });
    const cookie = cookies.get(SESSION_COOKIE);
    const cookieHeader = cookie ? `${SESSION_COOKIE}=${cookie}` : void 0;
    const result = await apiFetch(`/api/signing-keys/${keyId}/revoke`, {
      method: "POST",
      cookie: cookieHeader
    });
    if (!result.ok) return fail(result.status, { keyError: result.error || "Failed to revoke key" });
    return { keyRevoked: true };
  }
};

var _page_server_ts = /*#__PURE__*/Object.freeze({
  __proto__: null,
  actions: actions,
  load: load
});

const index = 35;
let component_cache;
const component = async () => component_cache ??= (await import('./_page.svelte--M0UvskB.js')).default;
const server_id = "src/routes/workspaces/[id]/roles/+page.server.ts";
const imports = ["_app/immutable/nodes/35.DOFJmtmA.js","_app/immutable/chunks/CWj6FrbW.js","_app/immutable/chunks/BCGREYIq.js","_app/immutable/chunks/glOr0qzD.js","_app/immutable/chunks/BcfnGdFX.js","_app/immutable/chunks/D8xXPBaV.js","_app/immutable/chunks/BS2x50Pa.js","_app/immutable/chunks/Bx7T9aDp.js","_app/immutable/chunks/YSzAdjSj.js","_app/immutable/chunks/wpI5rr64.js","_app/immutable/chunks/Bq1FU2Pv.js","_app/immutable/chunks/CHpGRhK0.js","_app/immutable/chunks/ilPNOASy.js","_app/immutable/chunks/vbEortrD.js","_app/immutable/chunks/B8IgtE21.js","_app/immutable/chunks/B5rGamLH.js","_app/immutable/chunks/DZ5gTY8C.js","_app/immutable/chunks/Di6_auBm.js","_app/immutable/chunks/B2UVO_nM.js"];
const stylesheets = ["_app/immutable/assets/Breadcrumb.BmuUfcGr.css","_app/immutable/assets/EmptyState.ByauYn1J.css","_app/immutable/assets/Badge.Cj2_Wv2F.css","_app/immutable/assets/Modal.CHlCiKdq.css","_app/immutable/assets/SectionHeader.Bb9ZRXXP.css","_app/immutable/assets/Tooltip.BrYf-THh.css","_app/immutable/assets/TimeAgo.BYfizd1M.css","_app/immutable/assets/35.BCknMLQH.css"];
const fonts = [];

export { component, fonts, imports, index, _page_server_ts as server, server_id, stylesheets };
//# sourceMappingURL=35-D_rAatFa.js.map
