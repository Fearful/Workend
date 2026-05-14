import { a as attr, e as escape_html, c as ensure_array_like, s as stringify, d as derived } from "../../../../../chunks/renderer.js";
import { B as Breadcrumb } from "../../../../../chunks/Breadcrumb.js";
import { E as EmptyState } from "../../../../../chunks/EmptyState.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
import { M as Modal } from "../../../../../chunks/Modal.js";
import { T as TimeAgo } from "../../../../../chunks/TimeAgo.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let editingSecretId = null;
    let confirmDeleteId = null;
    function cancelDelete() {
      confirmDeleteId = null;
    }
    let deletingSecret = derived(() => confirmDeleteId ? data.secrets.find((s) => s.id === confirmDeleteId) : null);
    Breadcrumb($$renderer2, {
      segments: (
        // Reset UI states after successful actions
        [
          { label: "workspaces", href: "/" },
          {
            label: data.workspace.name,
            href: `/workspaces/${data.workspace.id}`
          },
          { label: "secrets" }
        ]
      )
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-5sbird"><h1 class="svelte-5sbird">Secrets</h1></div> <nav class="ws-tabs svelte-5sbird" aria-label="Workspace sections"><a class="ws-tab svelte-5sbird"${attr("href", `/workspaces/${data.workspace.id}`)}>Overview</a> <a class="ws-tab svelte-5sbird"${attr("href", `/workspaces/${data.workspace.id}/dashboard`)}>Dashboard</a> <a class="ws-tab svelte-5sbird"${attr("href", `/workspaces/${data.workspace.id}/activity`)}>Activity</a> <a class="ws-tab active svelte-5sbird"${attr("href", `/workspaces/${data.workspace.id}/secrets`)}>Secrets</a> <a class="ws-tab svelte-5sbird"${attr("href", `/workspaces/${data.workspace.id}/roles`)}>Roles</a></nav> `);
    if (form?.error) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(form.error)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.created) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Secret created.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.updated) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Secret updated.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.deleted) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Secret deleted.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.secretsError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.secretsError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <div class="toolbar svelte-5sbird"><span class="toolbar-info svelte-5sbird">${escape_html(data.secrets.length)} secret${escape_html(data.secrets.length !== 1 ? "s" : "")}</span> <button>${escape_html("Add secret")}</button></div> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.secrets.length === 0 && true) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "⬡",
        message: "No secrets yet. Secrets are encrypted key-value pairs available to pipelines."
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="secret-list svelte-5sbird"><!--[-->`);
      const each_array = ensure_array_like(data.secrets);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let secret = each_array[$$index];
        $$renderer2.push(`<div class="secret-row svelte-5sbird"><div class="secret-name-col svelte-5sbird"><div class="secret-name svelte-5sbird">${escape_html(secret.name)}</div> <div class="secret-meta svelte-5sbird">by ${escape_html(secret.created_by)} · updated `);
        TimeAgo($$renderer2, { value: secret.updated_at });
        $$renderer2.push(`<!----></div></div> <span class="secret-value-mask svelte-5sbird">********</span> <div class="secret-actions svelte-5sbird"><button class="ghost">Update</button> <button class="ghost">Delete</button></div></div> `);
        if (editingSecretId === secret.id) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<div class="edit-form svelte-5sbird"><form method="POST" action="?/updateSecret"><input type="hidden" name="secret_id"${attr("value", secret.id)}/> <div class="edit-form-inner svelte-5sbird"><div class="field svelte-5sbird"><label${attr("for", `edit-value-${stringify(secret.id)}`)}>New value for <strong>${escape_html(secret.name)}</strong></label> <textarea${attr("id", `edit-value-${stringify(secret.id)}`)} name="value" required="" placeholder="New secret value..." class="svelte-5sbird"></textarea></div> <button type="submit">Save</button> <button type="button" class="ghost">Cancel</button></div></form></div>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    Modal($$renderer2, {
      open: confirmDeleteId !== null,
      title: "Delete secret",
      onClose: cancelDelete,
      children: ($$renderer3) => {
        if (deletingSecret()) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p>Delete <strong style="font-family: var(--font-mono);">${escape_html(deletingSecret().name)}</strong>? This cannot be undone. Pipelines using this secret will fail.</p> <form method="POST" action="?/deleteSecret"><input type="hidden" name="secret_id"${attr("value", deletingSecret().id)}/> <div style="display: flex; justify-content: flex-end; gap: var(--space-2); margin-top: var(--space-4);"><button type="button" class="ghost">Cancel</button> <button type="submit" style="background: var(--danger); border-color: var(--danger);">Delete</button></div></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}
export {
  _page as default
};
