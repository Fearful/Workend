import { a7 as attr, ac as ensure_array_like, a8 as escape_html, a9 as attr_class, ad as stringify, a2 as derived } from './renderer-D0X3o35U.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { M as Modal } from './Modal-BnjRYgDm.js';
import { S as SectionHeader } from './SectionHeader-BsGXUHV-.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import './Tooltip-Bg85MZ7d.js';
import './index-server-BFLhAcPs.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    const RESOURCES = [
      "project",
      "run",
      "pipeline",
      "secret",
      "member",
      "role",
      "workspace"
    ];
    const ACTIONS = ["read", "create", "update", "delete"];
    let editingRoleId = null;
    let deleteRoleId = null;
    let editingRole = derived(() => editingRoleId ? data.roles.find((r) => r.id === editingRoleId) : null);
    let deletingRole = derived(() => deleteRoleId ? data.roles.find((r) => r.id === deleteRoleId) : null);
    function hasPermission(permissions, resource, action) {
      return permissions.some((p) => p.resource === resource && p.action === action);
    }
    let showPrivateKey = false;
    let confirmRevokeKeyId = null;
    let revokingKey = derived(() => confirmRevokeKeyId ? data.signingKeys.find((k) => k.id === confirmRevokeKeyId) : null);
    Breadcrumb($$renderer2, {
      segments: (
        // ---------- Effects ----------
        [
          { label: "workspaces", href: "/" },
          {
            label: data.workspace.name,
            href: `/workspaces/${data.workspace.id}`
          },
          { label: "roles" }
        ]
      )
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-is1c29"><h1 class="svelte-is1c29">Roles &amp; Access</h1></div> <nav class="ws-tabs svelte-is1c29" aria-label="Workspace sections"><a class="ws-tab svelte-is1c29"${attr("href", `/workspaces/${data.workspace.id}`)}>Overview</a> <a class="ws-tab svelte-is1c29"${attr("href", `/workspaces/${data.workspace.id}/dashboard`)}>Dashboard</a> <a class="ws-tab svelte-is1c29"${attr("href", `/workspaces/${data.workspace.id}/activity`)}>Activity</a> <a class="ws-tab svelte-is1c29"${attr("href", `/workspaces/${data.workspace.id}/secrets`)}>Secrets</a> <a class="ws-tab active svelte-is1c29"${attr("href", `/workspaces/${data.workspace.id}/roles`)}>Roles</a></nav> `);
    if (form?.roleError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(form.roleError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.assignError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(form.assignError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.keyError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(form.keyError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.roleCreated) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Role created.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.roleUpdated) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Role updated.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.roleDeleted) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Role deleted.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.roleAssigned) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Role assigned.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.keyRevoked) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Signing key revoked.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.rolesError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.rolesError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<button>${escape_html("Create role")}</button>`);
      };
      SectionHeader($$renderer2, { title: "Roles", actions });
    }
    $$renderer2.push(`<!----> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.roles.length === 0 && true) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "◎",
        message: "No roles defined. Create custom roles to control access."
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array_3 = ensure_array_like(data.roles);
      for (let $$index_6 = 0, $$length = each_array_3.length; $$index_6 < $$length; $$index_6++) {
        let role = each_array_3[$$index_6];
        $$renderer2.push(`<div class="role-card svelte-is1c29"><div class="role-head svelte-is1c29"><span class="role-name svelte-is1c29">${escape_html(role.name)} `);
        if (role.is_system) {
          $$renderer2.push("<!--[0-->");
          Badge($$renderer2, {
            variant: "muted",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->System`);
            }
          });
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></span> `);
        if (!role.is_system) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<div class="role-actions svelte-is1c29"><button class="ghost">Edit</button> <button class="ghost">Delete</button></div>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></div> `);
        if (role.description) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<div class="role-desc svelte-is1c29">${escape_html(role.description)}</div>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <div class="perm-grid svelte-is1c29"><div class="perm-header svelte-is1c29"></div> <!--[-->`);
        const each_array_4 = ensure_array_like(ACTIONS);
        for (let $$index_3 = 0, $$length2 = each_array_4.length; $$index_3 < $$length2; $$index_3++) {
          let action = each_array_4[$$index_3];
          $$renderer2.push(`<div class="perm-header svelte-is1c29" style="text-align: center;">${escape_html(action)}</div>`);
        }
        $$renderer2.push(`<!--]--> <!--[-->`);
        const each_array_5 = ensure_array_like(RESOURCES);
        for (let $$index_5 = 0, $$length2 = each_array_5.length; $$index_5 < $$length2; $$index_5++) {
          let resource = each_array_5[$$index_5];
          $$renderer2.push(`<div class="perm-resource svelte-is1c29">${escape_html(resource)}</div> <!--[-->`);
          const each_array_6 = ensure_array_like(ACTIONS);
          for (let $$index_4 = 0, $$length3 = each_array_6.length; $$index_4 < $$length3; $$index_4++) {
            let action = each_array_6[$$index_4];
            $$renderer2.push(`<div class="perm-cell svelte-is1c29">`);
            if (hasPermission(role.permissions, resource, action)) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="perm-check svelte-is1c29">✓</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
              $$renderer2.push(`<span class="perm-empty svelte-is1c29">—</span>`);
            }
            $$renderer2.push(`<!--]--></div>`);
          }
          $$renderer2.push(`<!--]-->`);
        }
        $$renderer2.push(`<!--]--></div></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--> <div class="section svelte-is1c29">`);
    SectionHeader($$renderer2, { title: "Member roles" });
    $$renderer2.push(`<!----> `);
    if (data.members.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, { message: "No members to assign roles to." });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="member-list svelte-is1c29"><!--[-->`);
      const each_array_7 = ensure_array_like(data.members);
      for (let $$index_8 = 0, $$length = each_array_7.length; $$index_8 < $$length; $$index_8++) {
        let member = each_array_7[$$index_8];
        $$renderer2.push(`<div class="member-assign-row svelte-is1c29"><span class="member-assign-name svelte-is1c29">${escape_html(member.display_name)}</span> <span class="member-assign-email svelte-is1c29">${escape_html(member.email)}</span> <form method="POST" action="?/assignRole" class="role-assign-form svelte-is1c29"><input type="hidden" name="user_id"${attr("value", member.user_id)}/> <select name="role_id" class="role-assign-select svelte-is1c29"><!--[-->`);
        const each_array_8 = ensure_array_like(data.roles);
        for (let $$index_7 = 0, $$length2 = each_array_8.length; $$index_7 < $$length2; $$index_7++) {
          let role = each_array_8[$$index_7];
          $$renderer2.option({ value: role.id }, ($$renderer3) => {
            $$renderer3.push(`${escape_html(role.name)}`);
          });
        }
        $$renderer2.push(`<!--]--></select> <button type="submit" class="ghost">Assign</button></form></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--></div> <div class="section svelte-is1c29">`);
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<form method="POST" action="?/generateKey" style="display: inline;"><button type="submit">Generate key</button></form>`);
      };
      SectionHeader($$renderer2, { title: "Signing keys", actions });
    }
    $$renderer2.push(`<!----> `);
    if (data.keysError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.keysError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.signingKeys.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "⚿",
        message: "No signing keys. Generate a key pair to sign deployments."
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="key-list svelte-is1c29"><!--[-->`);
      const each_array_9 = ensure_array_like(data.signingKeys);
      for (let $$index_9 = 0, $$length = each_array_9.length; $$index_9 < $$length; $$index_9++) {
        let key = each_array_9[$$index_9];
        $$renderer2.push(`<div${attr_class("key-row svelte-is1c29", void 0, { "key-revoked": key.revoked_at !== null })}><div class="svelte-is1c29"><div class="key-hash svelte-is1c29">${escape_html(key.key_hash)}</div> <div class="key-meta svelte-is1c29">by ${escape_html(key.created_by)} · `);
        TimeAgo($$renderer2, { value: key.created_at });
        $$renderer2.push(`<!----></div></div> `);
        if (key.revoked_at) {
          $$renderer2.push("<!--[0-->");
          Badge($$renderer2, {
            variant: "danger",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->Revoked`);
            }
          });
        } else {
          $$renderer2.push("<!--[-1-->");
          Badge($$renderer2, {
            variant: "success",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->Active`);
            }
          });
        }
        $$renderer2.push(`<!--]--> <span style="font-family: var(--font-mono); font-size: var(--fs-xs); color: var(--text-dim);" class="svelte-is1c29">${escape_html(key.public_key.slice(0, 24))}...</span> `);
        if (!key.revoked_at) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<button class="ghost svelte-is1c29">Revoke</button>`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<span class="svelte-is1c29"></span>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--></div>  `);
    Modal($$renderer2, {
      open: editingRoleId !== null && editingRole() !== void 0,
      title: "Edit role",
      width: 560,
      onClose: () => {
        editingRoleId = null;
      },
      children: ($$renderer3) => {
        if (editingRole()) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<form method="POST" action="?/updateRole"><input type="hidden" name="role_id"${attr("value", editingRole().id)}/> <div class="field" style="margin-bottom: var(--space-3);"><label for="edit-role-name">Name</label> <input id="edit-role-name" name="name" type="text"${attr("value", editingRole().name)}/></div> <div class="field" style="margin-bottom: var(--space-3);"><label for="edit-role-desc">Description</label> <input id="edit-role-desc" name="description" type="text"${attr("value", editingRole().description)}/></div> <span id="perm-label-edit" style="display: block; margin-bottom: var(--space-2); font-size: var(--fs-sm); font-weight: var(--fw-medium);">Permissions</span> <div class="perm-grid-form svelte-is1c29" role="group" aria-labelledby="perm-label-edit"><div class="perm-header svelte-is1c29"></div> <!--[-->`);
          const each_array_10 = ensure_array_like(ACTIONS);
          for (let $$index_10 = 0, $$length = each_array_10.length; $$index_10 < $$length; $$index_10++) {
            let action = each_array_10[$$index_10];
            $$renderer3.push(`<div class="perm-header svelte-is1c29" style="text-align: center;">${escape_html(action)}</div>`);
          }
          $$renderer3.push(`<!--]--> <!--[-->`);
          const each_array_11 = ensure_array_like(RESOURCES);
          for (let $$index_12 = 0, $$length = each_array_11.length; $$index_12 < $$length; $$index_12++) {
            let resource = each_array_11[$$index_12];
            $$renderer3.push(`<div class="perm-resource svelte-is1c29">${escape_html(resource)}</div> <!--[-->`);
            const each_array_12 = ensure_array_like(ACTIONS);
            for (let $$index_11 = 0, $$length2 = each_array_12.length; $$index_11 < $$length2; $$index_11++) {
              let action = each_array_12[$$index_11];
              $$renderer3.push(`<label class="perm-cell svelte-is1c29"><input type="checkbox" name="permissions"${attr("value", `${stringify(resource)}:${stringify(action)}`)}${attr("checked", hasPermission(editingRole().permissions, resource, action), true)}${attr("aria-label", `${stringify(resource)} ${stringify(action)}`)} class="svelte-is1c29"/></label>`);
            }
            $$renderer3.push(`<!--]-->`);
          }
          $$renderer3.push(`<!--]--></div> <div class="form-footer svelte-is1c29"><button type="button" class="ghost">Cancel</button> <button type="submit">Save changes</button></div></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: deleteRoleId !== null && deletingRole() !== void 0,
      title: "Delete role",
      onClose: () => {
        deleteRoleId = null;
      },
      children: ($$renderer3) => {
        if (deletingRole()) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p>Delete the role <strong>${escape_html(deletingRole().name)}</strong>? Members with this role will lose these permissions.</p> <form method="POST" action="?/deleteRole"><input type="hidden" name="role_id"${attr("value", deletingRole().id)}/> <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4);"><button type="button" class="ghost">Cancel</button> <button type="submit" style="background: var(--danger); border-color: var(--danger);">Delete</button></div></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: confirmRevokeKeyId !== null && revokingKey() !== void 0,
      title: "Revoke signing key",
      onClose: () => {
        confirmRevokeKeyId = null;
      },
      children: ($$renderer3) => {
        if (revokingKey()) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p>Revoke key <strong style="font-family: var(--font-mono);">${escape_html(revokingKey().key_hash)}</strong>? Deployments signed with this key will no longer be verified.</p> <form method="POST" action="?/revokeKey"><input type="hidden" name="key_id"${attr("value", revokingKey().id)}/> <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4);"><button type="button" class="ghost">Cancel</button> <button type="submit" style="background: var(--danger); border-color: var(--danger);">Revoke</button></div></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: showPrivateKey && !!form?.privateKey,
      title: "Private key generated",
      width: 560,
      onClose: () => {
        showPrivateKey = false;
      },
      children: ($$renderer3) => {
        $$renderer3.push(`<div class="key-warning svelte-is1c29">Copy this private key now. It will not be shown again.</div> <div class="private-key-display svelte-is1c29">${escape_html(form?.privateKey)}</div> <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-3);"><button>Copy to clipboard</button> <button class="ghost">Done</button></div>`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte--M0UvskB.js.map
