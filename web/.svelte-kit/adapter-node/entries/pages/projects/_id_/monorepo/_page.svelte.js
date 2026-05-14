import { a as attr, e as escape_html, c as ensure_array_like } from "../../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../../chunks/exports.js";
import "../../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../../chunks/root.js";
import "../../../../../chunks/state.svelte.js";
/* empty css                                                        */
import { B as Badge } from "../../../../../chunks/Badge.js";
import { M as Modal } from "../../../../../chunks/Modal.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let detectPending = false;
    let autoMapPending = false;
    let detailModal = null;
    function badgeVariant(t) {
      switch (t) {
        case "npm":
          return "accent";
        case "go":
          return "info";
        case "cargo":
          return "warning";
        case "python":
          return "success";
        case "gradle":
          return "danger";
        default:
          return "muted";
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
    $$renderer2.push(`<!--]--> <div class="page-header svelte-16xdifz"><h2 class="svelte-16xdifz">Monorepo packages</h2> <div class="header-actions svelte-16xdifz"><button type="button"${attr("disabled", detectPending, true)}>${escape_html("Detect packages")}</button> `);
    if (data.packages.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<button type="button" class="ghost"${attr("disabled", autoMapPending, true)}>${escape_html("Auto-map tasks")}</button>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div> `);
    if (data.packages.length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="empty svelte-16xdifz"><p>No packages detected. Click "Detect packages" to scan for package manifests (package.json, go.mod, Cargo.toml, etc.).</p></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="pkg-grid svelte-16xdifz"><!--[-->`);
      const each_array = ensure_array_like(data.packages);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let pkg = each_array[$$index];
        $$renderer2.push(`<div class="pkg-card svelte-16xdifz" role="button" tabindex="0"><div class="pkg-card-header svelte-16xdifz"><span class="pkg-card-name svelte-16xdifz">${escape_html(pkg.name)}</span> `);
        Badge($$renderer2, {
          variant: badgeVariant(pkg.pkg_type),
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(pkg.pkg_type)}`);
          }
        });
        $$renderer2.push(`<!----></div> <div class="pkg-card-path svelte-16xdifz">${escape_html(pkg.path)}</div> <div class="pkg-card-footer svelte-16xdifz">${escape_html(pkg.task_count)} task${escape_html(pkg.task_count === 1 ? "" : "s")} scoped</div></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    Modal($$renderer2, {
      open: detailModal !== null,
      title: detailModal?.pkg ? detailModal.pkg.name : "Package detail",
      width: 560,
      onClose: () => detailModal = null,
      children: ($$renderer3) => {
        if (detailModal) {
          $$renderer3.push("<!--[0-->");
          if (detailModal.loading) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading...</div>`);
          } else if (detailModal.error) {
            $$renderer3.push("<!--[1-->");
            $$renderer3.push(`<p class="modal-error svelte-16xdifz">${escape_html(detailModal.error)}</p>`);
          } else if (detailModal.pkg) {
            $$renderer3.push("<!--[2-->");
            $$renderer3.push(`<div class="detail-row svelte-16xdifz"><span class="detail-label svelte-16xdifz">Path</span><span class="detail-value svelte-16xdifz">${escape_html(detailModal.pkg.path)}</span></div> <div class="detail-row svelte-16xdifz"><span class="detail-label svelte-16xdifz">Type</span><span class="detail-value svelte-16xdifz">`);
            Badge($$renderer3, {
              variant: badgeVariant(detailModal.pkg.pkg_type),
              size: "sm",
              children: ($$renderer4) => {
                $$renderer4.push(`<!---->${escape_html(detailModal.pkg.pkg_type)}`);
              }
            });
            $$renderer3.push(`<!----></span></div> <h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Scoped tasks (${escape_html(detailModal.scopes.length)})</h3> `);
            if (detailModal.scopes.length > 0) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<!--[-->`);
              const each_array_1 = ensure_array_like(detailModal.scopes);
              for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
                let scope = each_array_1[$$index_1];
                $$renderer3.push(`<div class="scope-item svelte-16xdifz"><span class="scope-task-name svelte-16xdifz">${escape_html(scope.task_name)}</span> <button type="button" class="ghost danger" style="font-size: 0.75rem;">Remove</button></div>`);
              }
              $$renderer3.push(`<!--]-->`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<p style="color: var(--text-dim); font-size: 0.875rem;">No tasks scoped to this package.</p>`);
            }
            $$renderer3.push(`<!--]--> `);
            if (data.tasks.length > 0) {
              $$renderer3.push("<!--[0-->");
              const scopedIDs = new Set(detailModal.scopes.map((s) => s.task_id));
              const available = data.tasks.filter((t) => !scopedIDs.has(t.id));
              if (available.length > 0) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<div class="scope-add-section svelte-16xdifz"><span style="font-size: 0.8125rem; color: var(--text-muted);">Add task scope:</span> <div class="scope-add-list svelte-16xdifz"><!--[-->`);
                const each_array_2 = ensure_array_like(available.slice(0, 15));
                for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
                  let t = each_array_2[$$index_2];
                  $$renderer3.push(`<button type="button" class="ghost" style="font-size: 0.75rem;">+ ${escape_html(t.name)}</button>`);
                }
                $$renderer3.push(`<!--]--> `);
                if (available.length > 15) {
                  $$renderer3.push("<!--[0-->");
                  $$renderer3.push(`<span style="font-size: 0.75rem; color: var(--text-dim);">+${escape_html(available.length - 15)} more</span>`);
                } else {
                  $$renderer3.push("<!--[-1-->");
                }
                $$renderer3.push(`<!--]--></div></div>`);
              } else {
                $$renderer3.push("<!--[-1-->");
              }
              $$renderer3.push(`<!--]-->`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]-->`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-16xdifz"><button type="button" class="ghost">Close</button></div>`);
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
