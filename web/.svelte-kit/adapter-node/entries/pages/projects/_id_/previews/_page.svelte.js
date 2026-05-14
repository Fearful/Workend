import { a as attr, c as ensure_array_like, e as escape_html } from "../../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../../chunks/exports.js";
import "../../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../../chunks/root.js";
import "../../../../../chunks/state.svelte.js";
/* empty css                                                        */
import { S as StatusPill } from "../../../../../chunks/StatusPill.js";
import { B as Badge } from "../../../../../chunks/Badge.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
/* empty css                                                        */
import { T as TimeAgo } from "../../../../../chunks/TimeAgo.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let newBranch = "";
    let actionPending = null;
    let deleteConfirmID = null;
    function previewStatus(status) {
      switch (status) {
        case "deployed":
          return "ready";
        case "deploying":
          return "cloning";
        case "stopped":
          return "cancelled";
        case "failed":
          return "failed";
        default:
          return "pending";
      }
    }
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
    $$renderer2.push(`<!--]--> <div class="page-header svelte-xgadx5"><h2 class="svelte-xgadx5">Deploy previews</h2></div> <form method="POST" action="?/createPreview" class="create-form svelte-xgadx5"><input type="text" name="branch" class="branch-input svelte-xgadx5" placeholder="Branch name (e.g. feature/new-ui)"${attr("value", newBranch)}/> <button type="submit"${attr("disabled", !newBranch.trim(), true)}>Create preview</button></form> `);
    if (data.previews.length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="empty svelte-xgadx5"><p>No deploy previews yet. Create one by entering a branch name above.</p></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(data.previews);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let p = each_array[$$index];
        $$renderer2.push(`<div class="preview-card svelte-xgadx5">`);
        StatusPill($$renderer2, { status: previewStatus(p.status), size: "sm" });
        $$renderer2.push(`<!----> <div class="card-body svelte-xgadx5"><div class="card-branch svelte-xgadx5">`);
        Badge($$renderer2, {
          variant: "info",
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(p.branch)}`);
          }
        });
        $$renderer2.push(`<!----> <span style="font-size: 0.75rem; color: var(--text-dim);">${escape_html(p.status)}</span></div> `);
        if (p.url) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<a${attr("href", p.url)} target="_blank" rel="noopener noreferrer" class="card-url svelte-xgadx5">${escape_html(p.url)}</a>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <div class="card-meta svelte-xgadx5">`);
        if (p.last_deployed_at) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span>Deployed `);
          TimeAgo($$renderer2, { value: p.last_deployed_at });
          $$renderer2.push(`<!----></span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <span>Created `);
        TimeAgo($$renderer2, { value: p.created_at });
        $$renderer2.push(`<!----></span> <label class="auto-deploy-toggle svelte-xgadx5"><input type="checkbox"${attr("checked", p.auto_deploy, true)} class="svelte-xgadx5"/> Auto-deploy</label></div></div> <div class="card-actions svelte-xgadx5"><button type="button" class="ghost svelte-xgadx5"${attr("disabled", actionPending === p.id, true)}>${escape_html(actionPending === p.id ? "Working..." : "Redeploy")}</button> `);
        if (p.status === "deployed" || p.status === "deploying") {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<button type="button" class="ghost svelte-xgadx5"${attr("disabled", actionPending === p.id, true)}>Stop</button>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> `);
        if (deleteConfirmID === p.id) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<div class="delete-confirm svelte-xgadx5"><form method="POST" action="?/deletePreview" class="inline-form"><input type="hidden" name="preview_id"${attr("value", p.id)}/> <button type="submit" class="danger svelte-xgadx5" style="font-size: 0.8125rem;">Confirm</button></form> <button type="button" class="ghost svelte-xgadx5" style="font-size: 0.8125rem;">Cancel</button></div>`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<button type="button" class="ghost danger svelte-xgadx5">Delete</button>`);
        }
        $$renderer2.push(`<!--]--></div></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
