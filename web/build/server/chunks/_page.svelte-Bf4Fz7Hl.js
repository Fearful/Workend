import { ac as ensure_array_like, a7 as attr, a8 as escape_html, a2 as derived } from './renderer-D0X3o35U.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { M as Modal } from './Modal-BnjRYgDm.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';
import './Tooltip-Bg85MZ7d.js';
import './index-server-BFLhAcPs.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let showCreate = false;
    let extendTarget = null;
    let extendHours = "2";
    let destroyTarget = null;
    function isActive(status) {
      return status === "running" || status === "starting" || status === "ready";
    }
    function isDead(status) {
      return status === "destroyed" || status === "expired" || status === "error";
    }
    function timeRemaining(expiresAt) {
      const diff = new Date(expiresAt).getTime() - Date.now();
      if (diff <= 0) return "expired";
      const hours = Math.floor(diff / 36e5);
      const minutes = Math.floor(diff % 36e5 / 6e4);
      if (hours > 0) return `${hours}h ${minutes}m`;
      return `${minutes}m`;
    }
    let activeSandboxes = derived(() => data.sandboxes.filter((s) => !isDead(s.status)));
    let deadSandboxes = derived(() => data.sandboxes.filter((s) => isDead(s.status)));
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
      let actions = function($$renderer3) {
        $$renderer3.push(`<button type="button">New Sandbox</button>`);
      };
      PageHeader($$renderer2, { title: "Sandboxes", actions });
    }
    $$renderer2.push(`<!----> `);
    if (data.sandboxes.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "📦",
        message: "No sandboxes yet. Create one to get an isolated dev environment."
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      if (activeSandboxes().length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="sandbox-list svelte-1oyniso"><!--[-->`);
        const each_array = ensure_array_like(activeSandboxes());
        for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
          let s = each_array[$$index];
          $$renderer2.push(`<div class="sandbox-card svelte-1oyniso"><div class="card-main svelte-1oyniso"><div class="card-header svelte-1oyniso">`);
          StatusPill($$renderer2, { status: s.status });
          $$renderer2.push(`<!----> <a${attr("href", `/sandboxes/${s.id}`)} class="card-branch svelte-1oyniso">${escape_html(s.branch || "main")}</a> `);
          Badge($$renderer2, {
            variant: "muted",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->${escape_html(s.project_id.slice(0, 8))}`);
            }
          });
          $$renderer2.push(`<!----></div> `);
          if (s.url && isActive(s.status)) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<a${attr("href", s.url)} target="_blank" rel="noopener noreferrer" class="card-url svelte-1oyniso">${escape_html(s.url)}</a>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--> <div class="card-meta svelte-1oyniso"><span>Created `);
          TimeAgo($$renderer2, { value: s.created_at });
          $$renderer2.push(`<!----></span> <span class="meta-sep svelte-1oyniso">*</span> <span class="time-remaining svelte-1oyniso">${escape_html(timeRemaining(s.expires_at))} remaining</span> `);
          if (s.port) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="meta-sep svelte-1oyniso">*</span> <span>Port ${escape_html(s.port)}</span>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></div></div> <div class="card-actions svelte-1oyniso"><button type="button" class="ghost svelte-1oyniso">Extend</button> <button type="button" class="danger svelte-1oyniso">Destroy</button></div></div>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      if (deadSandboxes().length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="section-label svelte-1oyniso">Expired / Destroyed</div> <div class="sandbox-list svelte-1oyniso"><!--[-->`);
        const each_array_1 = ensure_array_like(deadSandboxes());
        for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
          let s = each_array_1[$$index_1];
          $$renderer2.push(`<div class="sandbox-card dimmed svelte-1oyniso"><div class="card-main svelte-1oyniso"><div class="card-header svelte-1oyniso">`);
          StatusPill($$renderer2, { status: s.status });
          $$renderer2.push(`<!----> <a${attr("href", `/sandboxes/${s.id}`)} class="card-branch svelte-1oyniso">${escape_html(s.branch || "main")}</a> `);
          Badge($$renderer2, {
            variant: "muted",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->${escape_html(s.project_id.slice(0, 8))}`);
            }
          });
          $$renderer2.push(`<!----></div> <div class="card-meta svelte-1oyniso"><span>Created `);
          TimeAgo($$renderer2, { value: s.created_at });
          $$renderer2.push(`<!----></span> `);
          if (s.destroyed_at) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="meta-sep svelte-1oyniso">*</span> <span>Destroyed `);
            TimeAgo($$renderer2, { value: s.destroyed_at });
            $$renderer2.push(`<!----></span>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></div></div> <div class="card-actions svelte-1oyniso"></div></div>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--> `);
    Modal($$renderer2, {
      open: showCreate,
      title: "New Sandbox",
      width: 480,
      onClose: () => showCreate = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/create"><div class="field"><label for="sb-project">Project ID</label> <input id="sb-project" name="project_id" type="text" required="" placeholder="e.g. proj_abc123"/></div> <div class="field"><label for="sb-branch">Branch</label> <input id="sb-branch" name="branch" type="text" placeholder="main"/> <p class="modal-hint svelte-1oyniso">Defaults to "main" if blank.</p></div> <div class="field"><label for="sb-expires">Expires in (hours)</label> <input id="sb-expires" name="expires_hours" type="number" min="1" max="72" value="4"/> <p class="modal-hint svelte-1oyniso">1-72 hours. Defaults to 4.</p></div> <div class="modal-actions svelte-1oyniso"><button type="button" class="ghost">Cancel</button> <button type="submit">Create</button></div></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: extendTarget !== null,
      title: "Extend Sandbox",
      width: 400,
      onClose: () => extendTarget = null,
      children: ($$renderer3) => {
        if (extendTarget) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<form method="POST" action="?/extend"><input type="hidden" name="id"${attr("value", extendTarget.id)}/> <p class="confirm-text svelte-1oyniso">Extend sandbox on branch <strong class="svelte-1oyniso">${escape_html(extendTarget.branch || "main")}</strong>.</p> <div class="field"><label for="ext-hours">Additional hours</label> <input id="ext-hours" name="hours" type="number" min="1" max="72"${attr("value", extendHours)}/> <p class="modal-hint svelte-1oyniso">1-72 hours to add.</p></div> <div class="modal-actions svelte-1oyniso"><button type="button" class="ghost">Cancel</button> <button type="submit">Extend</button></div></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: destroyTarget !== null,
      title: "Destroy Sandbox",
      width: 400,
      onClose: () => destroyTarget = null,
      children: ($$renderer3) => {
        if (destroyTarget) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<form method="POST" action="?/destroy"><input type="hidden" name="id"${attr("value", destroyTarget.id)}/> <p class="confirm-text svelte-1oyniso">This will permanently destroy the sandbox on branch <strong class="svelte-1oyniso">${escape_html(destroyTarget.branch || "main")}</strong>.
        All data inside the container will be lost.</p> <div class="modal-actions svelte-1oyniso"><button type="button" class="ghost">Cancel</button> <button type="submit" class="danger">Destroy</button></div></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-Bf4Fz7Hl.js.map
