import { e as escape_html, c as ensure_array_like, b as attr_class, s as stringify, a as attr } from "../../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../../chunks/exports.js";
import "../../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../../chunks/root.js";
import "../../../../../chunks/state.svelte.js";
import { s as shortSha } from "../../../../../chunks/utils2.js";
import { B as Badge } from "../../../../../chunks/Badge.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
import { M as Modal } from "../../../../../chunks/Modal.js";
import { E as EmptyState } from "../../../../../chunks/EmptyState.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let dragName = null;
    let dropName = null;
    let prModal = null;
    let switchModal = null;
    let switchPending = false;
    function shortSha$1(sha) {
      return shortSha(sha, 8);
    }
    $$renderer2.push(`<h2 class="section-title svelte-wtiglm">Branches <span class="source-pill svelte-wtiglm">${escape_html(data.branchesSource === "provider" ? "OAuth" : "ls-remote")}</span></h2> <p class="hint svelte-wtiglm">Click a branch to switch the project to it (re-clones in the background). `);
    if (data.canCreatePR) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`Drag a branch onto another to open a pull/merge request from source → target.`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`Connect an OAuth provider in <a href="/settings">Settings</a> to enable drag-to-PR.`);
    }
    $$renderer2.push(`<!--]--></p> `);
    if (data.branchesError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.branchesError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
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
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.prURL) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->PR #${escape_html(form.prNumber)} opened. <a${attr("href", form.prURL)} target="_blank" rel="noopener">View on provider →</a>`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <section class="branch-list svelte-wtiglm">`);
    if (data.branches.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, { message: "No branches found." });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(data.branches);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let b = each_array[$$index];
        $$renderer2.push(`<div${attr_class(`branch-row ${stringify(dragName === b.name ? "dragging" : "")} ${stringify(dropName === b.name ? "drop-target" : "")}`, "svelte-wtiglm")} draggable="true" role="button" tabindex="0"><span class="grip svelte-wtiglm" aria-hidden="true">⋮⋮</span> <span class="name svelte-wtiglm">${escape_html(b.name)}</span> `);
        if (b.default) {
          $$renderer2.push("<!--[0-->");
          Badge($$renderer2, {
            variant: "success",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->default`);
            }
          });
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<span class="svelte-wtiglm"></span>`);
        }
        $$renderer2.push(`<!--]--> `);
        if (b.protected) {
          $$renderer2.push("<!--[0-->");
          Badge($$renderer2, {
            variant: "warning",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->protected`);
            }
          });
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<span class="svelte-wtiglm"></span>`);
        }
        $$renderer2.push(`<!--]--> <span class="sha svelte-wtiglm">${escape_html(shortSha$1(b.commit_sha))}</span></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></section> `);
    Modal($$renderer2, {
      open: switchModal !== null,
      title: "Switch branch?",
      width: 420,
      onClose: () => switchModal = null,
      children: ($$renderer3) => {
        if (switchModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p style="margin: 0 0 var(--space-3) 0;">Switch this project to <strong>${escape_html(switchModal.name)}</strong>? The repo will be re-cloned.</p> `);
          {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div style="display:flex; gap: var(--space-2); justify-content:flex-end;"><button type="button" class="ghost"${attr("disabled", switchPending, true)}>Cancel</button> <button type="button"${attr("disabled", switchPending, true)}>${escape_html("Switch")}</button></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: prModal !== null,
      title: prModal ? `Open PR: ${prModal.source} → ${prModal.target}` : "",
      width: 560,
      onClose: () => prModal = null,
      children: ($$renderer3) => {
        if (prModal) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<form method="POST" action="?/createPR" class="pr-modal-form svelte-wtiglm"><input type="hidden" name="source"${attr("value", prModal.source)}/> <input type="hidden" name="target"${attr("value", prModal.target)}/> <div class="field"><label for="pr-title">Title</label> <input id="pr-title" name="title"${attr("value", prModal.title)} required=""/></div> <div class="field"><label for="pr-body">Description (optional)</label> <textarea id="pr-body" name="body" rows="6">`);
          const $$body = escape_html(prModal.body);
          if ($$body) {
            $$renderer3.push(`${$$body}`);
          }
          $$renderer3.push(`</textarea></div> <div style="display:flex; gap: var(--space-2); justify-content:flex-end;"><button type="button" class="ghost">Cancel</button> <button type="submit">Open PR</button></div></form>`);
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
