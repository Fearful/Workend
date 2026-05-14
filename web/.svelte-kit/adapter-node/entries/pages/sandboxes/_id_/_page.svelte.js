import { e as escape_html, a as attr, c as ensure_array_like, b as attr_class, d as derived } from "../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/root.js";
import "../../../../chunks/state.svelte.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
import { P as Panel } from "../../../../chunks/Panel.js";
import { S as StatusPill } from "../../../../chunks/StatusPill.js";
/* empty css                                                     */
import { M as Modal } from "../../../../chunks/Modal.js";
import { T as TimeAgo } from "../../../../chunks/TimeAgo.js";
import { F as FlashMessage } from "../../../../chunks/FlashMessage.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let showExtend = false;
    let showDestroy = false;
    let activeSessionId = null;
    let termHistory = [];
    function isActive(status) {
      return status === "running" || status === "starting" || status === "ready";
    }
    function timeRemaining(expiresAt) {
      const diff = new Date(expiresAt).getTime() - Date.now();
      if (diff <= 0) return "expired";
      const hours = Math.floor(diff / 36e5);
      const minutes = Math.floor(diff % 36e5 / 6e4);
      if (hours > 0) return `${hours}h ${minutes}m remaining`;
      return `${minutes}m remaining`;
    }
    let openSessions = derived(() => data.shellSessions.filter((s) => s.status === "open" || s.status === "active"));
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
    Breadcrumb($$renderer2, {
      segments: [
        { label: "sandboxes", href: "/sandboxes" },
        {
          label: `${data.sandbox.branch || "main"} (${data.sandbox.id.slice(0, 8)})`
        }
      ]
    });
    $$renderer2.push(`<!----> <div class="hero svelte-prr3wg"><div class="hero-main svelte-prr3wg"><div class="hero-line svelte-prr3wg">`);
    StatusPill($$renderer2, { status: data.sandbox.status });
    $$renderer2.push(`<!----> <h1 class="svelte-prr3wg">${escape_html(data.sandbox.branch || "main")}</h1></div> <div class="hero-meta svelte-prr3wg"><span>Project <a${attr("href", `/projects/${data.sandbox.project_id}`)}>${escape_html(data.sandbox.project_id.slice(0, 8))}</a></span> <span class="hero-sep svelte-prr3wg">*</span> <span class="countdown svelte-prr3wg">${escape_html(timeRemaining(data.sandbox.expires_at))}</span> <span class="hero-sep svelte-prr3wg">*</span> <span>Created `);
    TimeAgo($$renderer2, { value: data.sandbox.created_at });
    $$renderer2.push(`<!----></span></div></div> <div class="actions svelte-prr3wg">`);
    if (isActive(data.sandbox.status)) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<button type="button" class="ghost">Extend</button> <button type="button" class="danger">Destroy</button>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div> <div class="layout svelte-prr3wg"><div class="main-col svelte-prr3wg">`);
    Panel($$renderer2, {
      title: "Shell",
      padding: "compact",
      children: ($$renderer3) => {
        if (!isActive(data.sandbox.status)) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="no-shell svelte-prr3wg">Shell is unavailable. Sandbox is ${escape_html(data.sandbox.status)}.</div>`);
        } else if (openSessions().length === 0 && !activeSessionId) {
          $$renderer3.push("<!--[1-->");
          $$renderer3.push(`<div class="no-shell svelte-prr3wg">No active shell session. <form method="POST" action="?/createShell" class="inline-form" style="display: inline;"><button type="submit">Create session</button></form></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<div class="shell-header svelte-prr3wg"><div class="shell-tabs svelte-prr3wg"><!--[-->`);
          const each_array = ensure_array_like(openSessions());
          for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
            let sess = each_array[$$index];
            $$renderer3.push(`<button type="button"${attr_class("shell-tab svelte-prr3wg", void 0, { "active": activeSessionId === sess.id })}>${escape_html(sess.id.slice(0, 8))}</button>`);
          }
          $$renderer3.push(`<!--]--> <form method="POST" action="?/createShell" class="inline-form" style="display: inline;"><button type="submit" class="shell-tab svelte-prr3wg" title="New session">+</button></form></div> `);
          {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--></div> <div class="term-output svelte-prr3wg">`);
          if (termHistory.length === 0) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div class="term-empty svelte-prr3wg">Ready. Type a command below.</div>`);
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`<!--[-->`);
            const each_array_1 = ensure_array_like(termHistory);
            for (let i = 0, $$length = each_array_1.length; i < $$length; i++) {
              let entry = each_array_1[i];
              $$renderer3.push(`<div class="term-entry svelte-prr3wg"><div><span class="term-prompt svelte-prr3wg">$</span><span class="term-cmd svelte-prr3wg">${escape_html(entry.command)}</span></div> `);
              if (entry.output) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<div class="term-result svelte-prr3wg">${escape_html(entry.output)}</div>`);
              } else {
                $$renderer3.push("<!--[-1-->");
              }
              $$renderer3.push(`<!--]--> `);
              if (entry.exitCode !== 0) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<div class="term-exit-error svelte-prr3wg">exit ${escape_html(entry.exitCode)}</div>`);
              } else {
                $$renderer3.push("<!--[-1-->");
              }
              $$renderer3.push(`<!--]--></div>`);
            }
            $$renderer3.push(`<!--]-->`);
          }
          $$renderer3.push(`<!--]--></div> `);
          {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----> `);
    if (data.shellSessions.length > 0) {
      $$renderer2.push("<!--[0-->");
      Panel($$renderer2, {
        title: "Shell Sessions",
        children: ($$renderer3) => {
          $$renderer3.push(`<div class="session-list svelte-prr3wg"><!--[-->`);
          const each_array_2 = ensure_array_like(data.shellSessions);
          for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
            let sess = each_array_2[$$index_2];
            $$renderer3.push(`<div class="session-item svelte-prr3wg"><span class="session-id svelte-prr3wg">${escape_html(sess.id.slice(0, 12))}</span> `);
            StatusPill($$renderer3, {
              status: sess.status === "open" || sess.status === "active" ? "running" : "succeeded",
              size: "sm",
              label: sess.status
            });
            $$renderer3.push(`<!----> `);
            TimeAgo($$renderer3, { value: sess.created_at });
            $$renderer3.push(`<!----></div>`);
          }
          $$renderer3.push(`<!--]--></div>`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> <div class="meta-col svelte-prr3wg">`);
    Panel($$renderer2, {
      title: "Sandbox details",
      children: ($$renderer3) => {
        $$renderer3.push(`<div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Status</span><span class="value svelte-prr3wg">`);
        StatusPill($$renderer3, { status: data.sandbox.status, size: "sm" });
        $$renderer3.push(`<!----></span></div> <div class="row svelte-prr3wg"><span class="label svelte-prr3wg">URL</span><span class="value svelte-prr3wg">`);
        if (data.sandbox.url && isActive(data.sandbox.status)) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<a${attr("href", data.sandbox.url)} target="_blank" rel="noopener noreferrer">${escape_html(data.sandbox.url)}</a>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`${escape_html(data.sandbox.url || "---")}`);
        }
        $$renderer3.push(`<!--]--></span></div> <div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Port</span><span class="value svelte-prr3wg">${escape_html(data.sandbox.port || "---")}</span></div> <div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Container</span><span class="value svelte-prr3wg">${escape_html(data.sandbox.container_id ? data.sandbox.container_id.slice(0, 12) : "---")}</span></div> <div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Branch</span><span class="value svelte-prr3wg">${escape_html(data.sandbox.branch || "main")}</span></div> <div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Created</span><span class="value svelte-prr3wg">`);
        TimeAgo($$renderer3, { value: data.sandbox.created_at });
        $$renderer3.push(`<!----></span></div> <div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Expires</span><span class="value svelte-prr3wg">`);
        TimeAgo($$renderer3, { value: data.sandbox.expires_at });
        $$renderer3.push(`<!----></span></div> `);
        if (data.sandbox.destroyed_at) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="row svelte-prr3wg"><span class="label svelte-prr3wg">Destroyed</span><span class="value svelte-prr3wg">`);
          TimeAgo($$renderer3, { value: data.sandbox.destroyed_at });
          $$renderer3.push(`<!----></span></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----></div></div> `);
    Modal($$renderer2, {
      open: showExtend,
      title: "Extend Sandbox",
      width: 400,
      onClose: () => showExtend = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/extend"><div class="field"><label for="ext-hours">Additional hours</label> <input id="ext-hours" name="hours" type="number" min="1" max="72" value="2"/> <p class="modal-hint svelte-prr3wg">1-72 hours to add to the current expiry.</p></div> <div class="modal-actions svelte-prr3wg"><button type="button" class="ghost">Cancel</button> <button type="submit">Extend</button></div></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: showDestroy,
      title: "Destroy Sandbox",
      width: 400,
      onClose: () => showDestroy = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/destroy"><p class="confirm-text svelte-prr3wg">This will permanently destroy this sandbox. All data inside the container will be lost.</p> <div class="modal-actions svelte-prr3wg"><button type="button" class="ghost">Cancel</button> <button type="submit" class="danger">Destroy</button></div></form>`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}
export {
  _page as default
};
