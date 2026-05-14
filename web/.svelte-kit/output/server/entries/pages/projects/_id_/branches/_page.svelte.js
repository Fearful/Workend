import { a as attr, e as escape_html, c as ensure_array_like, b as attr_class, s as stringify, d as derived } from "../../../../../chunks/renderer.js";
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
import { S as SectionHeader } from "../../../../../chunks/SectionHeader.js";
import { T as Tooltip } from "../../../../../chunks/Tooltip.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let dragName = null;
    let dropName = null;
    let prModal = null;
    let switchModal = null;
    let switchPending = false;
    let filter = "";
    let showBots = false;
    function isBot(name) {
      return name.startsWith("dependabot/") || name.startsWith("renovate/");
    }
    let filteredBranches = derived(() => {
      const q = filter.trim().toLowerCase();
      return data.branches.filter((b) => {
        if (q && !b.name.toLowerCase().includes(q)) return false;
        if (!q && isBot(b.name)) return false;
        return true;
      });
    });
    let botCount = derived(() => data.branches.filter((b) => isBot(b.name)).length);
    function shortSha$1(sha) {
      return shortSha(sha, 8);
    }
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<span class="source-pill svelte-wtiglm">${escape_html(data.branchesSource === "provider" ? "OAuth" : "ls-remote")}</span>`);
      };
      SectionHeader($$renderer2, { title: "Branches", actions });
    }
    $$renderer2.push(`<!----> <p class="hint svelte-wtiglm">Click a branch to switch the project to it (re-clones in the background). `);
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
    $$renderer2.push(`<!--]--> <div class="filter-bar svelte-wtiglm"><input type="search" class="filter-input svelte-wtiglm" placeholder="Filter branches…"${attr("value", filter)}/> `);
    if (botCount() > 0 && !filter.trim()) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<label class="show-bots svelte-wtiglm"><input type="checkbox"${attr("checked", showBots, true)} class="svelte-wtiglm"/> <span>Show ${escape_html(botCount())} bot branches</span></label>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <span class="count svelte-wtiglm">${escape_html(filteredBranches().length)} of ${escape_html(data.branches.length)}</span></div> <section class="branch-list svelte-wtiglm">`);
    if (data.branches.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, { message: "No branches found." });
    } else if (filteredBranches().length === 0) {
      $$renderer2.push("<!--[1-->");
      EmptyState($$renderer2, { message: "No branches match." });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(filteredBranches());
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
        $$renderer2.push(`<!--]--> `);
        Tooltip($$renderer2, {
          text: b.commit_sha,
          children: ($$renderer3) => {
            $$renderer3.push(`<span class="sha svelte-wtiglm">${escape_html(shortSha$1(b.commit_sha))}</span>`);
          }
        });
        $$renderer2.push(`<!----></div>`);
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
