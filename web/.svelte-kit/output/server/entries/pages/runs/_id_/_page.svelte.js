import { e as escape_html, a as attr, c as ensure_array_like } from "../../../../chunks/renderer.js";
import { o as onDestroy } from "../../../../chunks/index-server.js";
import "@sveltejs/kit/internal";
import "../../../../chunks/exports.js";
import "../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../chunks/root.js";
import "../../../../chunks/state.svelte.js";
import { f as formatRelative, c as formatDuration, s as shortSha } from "../../../../chunks/utils2.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
import { P as Panel } from "../../../../chunks/Panel.js";
import { S as StatusDot } from "../../../../chunks/StatusDot.js";
import { B as Badge } from "../../../../chunks/Badge.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let commentDraft = "";
    function bodyParts(body) {
      const parts = [];
      const re = /@([A-Za-z0-9._-]+)/g;
      let last = 0;
      let m;
      while ((m = re.exec(body)) !== null) {
        if (m.index > last) parts.push({ text: body.slice(last, m.index), mention: false });
        parts.push({ text: m[0], mention: true });
        last = m.index + m[0].length;
      }
      if (last < body.length) parts.push({ text: body.slice(last), mention: false });
      return parts;
    }
    function shortSha$1(sha) {
      return shortSha(sha, 12);
    }
    let liveLog = "";
    let liveStatus = "queued";
    let cancelling = false;
    function stopStream() {
    }
    onDestroy(stopStream);
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        {
          label: "back to project",
          href: `/projects/${data.run.project_id}`
        },
        { label: `run ${data.run.id.slice(0, 8)}` }
      ]
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-1eidjaf"><h1 class="svelte-1eidjaf">`);
    StatusDot($$renderer2, { status: liveStatus, size: 10 });
    $$renderer2.push(`<!----> ${escape_html(data.run.task_name)} <span class="dim svelte-1eidjaf">(${escape_html(data.run.task_source)})</span> `);
    if (data.run.timed_out) {
      $$renderer2.push("<!--[0-->");
      Badge($$renderer2, {
        variant: "warning",
        size: "sm",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->timed out`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.run.attempt > 1) {
      $$renderer2.push("<!--[0-->");
      Badge($$renderer2, {
        variant: "info",
        size: "sm",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->retry · attempt ${escape_html(data.run.attempt)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></h1> <div class="actions svelte-1eidjaf">`);
    {
      $$renderer2.push("<!--[1-->");
      $$renderer2.push(`<button type="button" class="ghost svelte-1eidjaf"${attr("title", "Notify me on this device when this run finishes")} aria-label="Toggle desktop notification">Notify ${escape_html("off")}</button> <button type="button" class="danger svelte-1eidjaf"${attr("disabled", cancelling, true)}>${escape_html("Cancel")}</button>`);
    }
    $$renderer2.push(`<!--]--></div></div> <div class="layout svelte-1eidjaf"><div class="main-col svelte-1eidjaf"><section class="log-container svelte-1eidjaf"><div class="log-header svelte-1eidjaf"><h2 class="svelte-1eidjaf">Log</h2> <div class="log-controls svelte-1eidjaf">`);
    {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="svelte-1eidjaf"><span class="spinner svelte-1eidjaf"></span>${escape_html(liveStatus)}</span>`);
    }
    $$renderer2.push(`<!--]--> <span class="svelte-1eidjaf">${escape_html(liveLog.length)} bytes</span> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div> `);
    {
      $$renderer2.push("<!--[1-->");
      $$renderer2.push(`<div class="empty-log svelte-1eidjaf">Waiting for output…</div>`);
    }
    $$renderer2.push(`<!--]--></section> `);
    Panel($$renderer2, {
      title: "Comments",
      children: ($$renderer3) => {
        if (data.comments.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="empty-comments svelte-1eidjaf">No comments yet.</div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_1 = ensure_array_like(data.comments);
          for (let $$index_2 = 0, $$length = each_array_1.length; $$index_2 < $$length; $$index_2++) {
            let c = each_array_1[$$index_2];
            $$renderer3.push(`<div class="comment svelte-1eidjaf"><div class="comment-head svelte-1eidjaf"><strong class="comment-author svelte-1eidjaf">${escape_html(c.user_display_name || "someone")}</strong> <span class="comment-time svelte-1eidjaf">${escape_html(formatRelative(c.created_at))}</span> `);
            if (data.user && c.user_id === data.user.id) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<form method="POST" action="?/deleteComment" class="comment-delete-form svelte-1eidjaf"><input type="hidden" name="id"${attr("value", c.id)} class="svelte-1eidjaf"/> <button type="submit" class="ghost comment-delete-btn svelte-1eidjaf" title="Delete">×</button></form>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--></div> <div class="comment-body svelte-1eidjaf"><!--[-->`);
            const each_array_2 = ensure_array_like(bodyParts(c.body));
            for (let i = 0, $$length2 = each_array_2.length; i < $$length2; i++) {
              let part = each_array_2[i];
              if (part.mention) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<span class="mention svelte-1eidjaf">${escape_html(part.text)}</span>`);
              } else {
                $$renderer3.push("<!--[-1-->");
                $$renderer3.push(`<span class="svelte-1eidjaf">${escape_html(part.text)}</span>`);
              }
              $$renderer3.push(`<!--]-->`);
            }
            $$renderer3.push(`<!--]--></div></div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
        $$renderer3.push(`<!--]--> <form method="POST" action="?/comment" class="comment-form svelte-1eidjaf"><textarea name="body" rows="3" placeholder="Add a comment. Use @display-name to mention a workspace member." class="svelte-1eidjaf">`);
        const $$body = escape_html(commentDraft);
        if ($$body) {
          $$renderer3.push(`${$$body}`);
        }
        $$renderer3.push(`</textarea> `);
        if (form?.commentError) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="comment-form-error svelte-1eidjaf">${escape_html(form.commentError)}</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> <div class="comment-form-actions svelte-1eidjaf"><button type="submit" class="svelte-1eidjaf">Post comment</button></div></form>`);
      }
    });
    $$renderer2.push(`<!----></div> <div class="meta-col svelte-1eidjaf">`);
    Panel($$renderer2, {
      title: "Run details",
      children: ($$renderer3) => {
        $$renderer3.push(`<div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Status</span><span class="value svelte-1eidjaf">${escape_html(liveStatus)}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Exit code</span><span class="value svelte-1eidjaf">${escape_html("—")}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Commit</span><span class="value svelte-1eidjaf">${escape_html(shortSha$1(data.run.commit_sha))}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Duration</span><span class="value svelte-1eidjaf">${escape_html(formatDuration(data.run.started_at, data.run.finished_at))}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Started</span><span class="value svelte-1eidjaf">${escape_html(data.run.started_at ? new Date(data.run.started_at).toLocaleString() : "—")}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Finished</span><span class="value svelte-1eidjaf">${escape_html(data.run.finished_at ? new Date(data.run.finished_at).toLocaleString() : "—")}</span></div> `);
        if (data.run.params && (data.run.params.env && Object.keys(data.run.params.env).length > 0 || data.run.params.args && data.run.params.args.length > 0)) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Inputs</span> <span class="value value-pre svelte-1eidjaf">${escape_html([
            ...data.run.params.env ? Object.entries(data.run.params.env).map(([k, v]) => `${k}=${v}`) : [],
            ...data.run.params.args ?? []
          ].join("\n"))}</span></div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!----></div></div>`);
  });
}
export {
  _page as default
};
