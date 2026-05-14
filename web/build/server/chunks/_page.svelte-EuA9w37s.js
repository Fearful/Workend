import { a8 as escape_html, a7 as attr, ac as ensure_array_like } from './renderer-D0X3o35U.js';
import { o as onDestroy } from './index-server-BFLhAcPs.js';
import './root-BPP4VMgG.js';
import './state.svelte-DaEZ9E9Z.js';
import { c as formatDuration, b as formatRelative, s as shortSha } from './utils2-BUPlP7zG.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { P as Panel } from './Panel-B1HKSjmN.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { B as Badge } from './Badge-imHJGUZA.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { T as Tooltip } from './Tooltip-Bg85MZ7d.js';
import { M as Modal } from './Modal-BnjRYgDm.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let commentDraft = "";
    let showShareModal = false;
    let shareCopied = null;
    let showSignModal = false;
    let showProvenanceModal = false;
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
    function formatBytes(b) {
      if (b < 1024) return `${b} B`;
      if (b < 1024 * 1024) return `${(b / 1024).toFixed(1)} KB`;
      if (b < 1024 * 1024 * 1024) return `${(b / 1024 / 1024).toFixed(1)} MB`;
      return `${(b / 1024 / 1024 / 1024).toFixed(2)} GB`;
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
    $$renderer2.push(`<!----> <div class="hero svelte-1eidjaf"><div class="hero-main svelte-1eidjaf"><div class="hero-line svelte-1eidjaf">`);
    StatusPill($$renderer2, { status: liveStatus });
    $$renderer2.push(`<!----> <h1 class="svelte-1eidjaf">${escape_html(data.run.task_name)} <span class="dim svelte-1eidjaf">(${escape_html(data.run.task_source)})</span></h1> `);
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
          $$renderer3.push(`<!---->attempt ${escape_html(data.run.attempt)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> <div class="hero-meta svelte-1eidjaf">`);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <span class="hero-meta-item svelte-1eidjaf"><span class="hero-meta-label svelte-1eidjaf">Duration</span><strong class="svelte-1eidjaf">${escape_html(formatDuration(data.run.started_at, data.run.finished_at))}</strong></span> `);
    if (data.run.commit_sha) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="hero-sep svelte-1eidjaf">·</span> `);
      Tooltip($$renderer2, {
        text: data.run.commit_sha,
        children: ($$renderer3) => {
          $$renderer3.push(`<span class="hero-meta-item svelte-1eidjaf"><span class="hero-meta-label svelte-1eidjaf">Commit</span><strong class="mono svelte-1eidjaf">${escape_html(data.run.commit_sha.slice(0, 7))}</strong></span>`);
        }
      });
      $$renderer2.push(`<!---->`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <span class="hero-sep svelte-1eidjaf">·</span> <span class="hero-meta-item svelte-1eidjaf"><span class="hero-meta-label svelte-1eidjaf">Started</span><strong class="svelte-1eidjaf">`);
    TimeAgo($$renderer2, { value: data.run.started_at });
    $$renderer2.push(`<!----></strong></span></div></div> <div class="actions svelte-1eidjaf">`);
    {
      $$renderer2.push("<!--[1-->");
      $$renderer2.push(`<button type="button" class="ghost svelte-1eidjaf"${attr("title", "Notify me on this device when this run finishes")} aria-label="Toggle desktop notification">🔔 ${escape_html("Off")}</button> <button type="button" class="danger svelte-1eidjaf"${attr("disabled", cancelling, true)}>${escape_html("Cancel run")}</button>`);
    }
    $$renderer2.push(`<!--]--></div></div> <div class="layout svelte-1eidjaf"><div class="main-col svelte-1eidjaf"><section class="panel log-section svelte-1eidjaf"><div class="log-section-head svelte-1eidjaf"><h2 class="svelte-1eidjaf">Log <span class="log-bytes svelte-1eidjaf">${escape_html(liveLog.length.toLocaleString())} bytes</span></h2> `);
    {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="live-badge svelte-1eidjaf">live · ${escape_html(liveStatus)}</span>`);
    }
    $$renderer2.push(`<!--]--></div> `);
    {
      $$renderer2.push("<!--[1-->");
      $$renderer2.push(`<div class="empty-log svelte-1eidjaf">Waiting for output…</div>`);
    }
    $$renderer2.push(`<!--]--></section> `);
    if (data.artifacts.length > 0) {
      $$renderer2.push("<!--[0-->");
      {
        let actions = function($$renderer3) {
          $$renderer3.push(`<span style="color: var(--text-dim); font-size: var(--fs-xs); font-family: var(--font-mono);" class="svelte-1eidjaf">${escape_html(data.artifacts.length)} file${escape_html(data.artifacts.length === 1 ? "" : "s")}</span>`);
        };
        Panel($$renderer2, {
          title: "Artifacts",
          actions,
          children: ($$renderer3) => {
            $$renderer3.push(`<ul class="artifact-list svelte-1eidjaf"><!--[-->`);
            const each_array_1 = ensure_array_like(data.artifacts);
            for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
              let a = each_array_1[$$index_1];
              $$renderer3.push(`<li class="svelte-1eidjaf"><a${attr("href", `/api/artifacts/${a.id}/download`)} class="artifact-name svelte-1eidjaf">${escape_html(a.relative_path)}</a> <span class="artifact-size svelte-1eidjaf">${escape_html(formatBytes(a.size_bytes))}</span> `);
              if (a.mime_type) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<span class="artifact-mime svelte-1eidjaf">${escape_html(a.mime_type)}</span>`);
              } else {
                $$renderer3.push("<!--[-1-->");
              }
              $$renderer3.push(`<!--]--></li>`);
            }
            $$renderer3.push(`<!--]--></ul>`);
          }
        });
      }
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    Panel($$renderer2, {
      title: "Comments",
      children: ($$renderer3) => {
        if (data.comments.length === 0) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<div class="empty-comments svelte-1eidjaf">No comments yet.</div>`);
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`<!--[-->`);
          const each_array_2 = ensure_array_like(data.comments);
          for (let $$index_3 = 0, $$length = each_array_2.length; $$index_3 < $$length; $$index_3++) {
            let c = each_array_2[$$index_3];
            $$renderer3.push(`<div class="comment svelte-1eidjaf"><div class="comment-head svelte-1eidjaf"><strong class="comment-author svelte-1eidjaf">${escape_html(c.user_display_name || "someone")}</strong> <span class="comment-time svelte-1eidjaf">${escape_html(formatRelative(c.created_at))}</span> `);
            if (data.user && c.user_id === data.user.id) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<form method="POST" action="?/deleteComment" class="comment-delete-form svelte-1eidjaf"><input type="hidden" name="id"${attr("value", c.id)} class="svelte-1eidjaf"/> <button type="submit" class="ghost comment-delete-btn svelte-1eidjaf" title="Delete">×</button></form>`);
            } else {
              $$renderer3.push("<!--[-1-->");
            }
            $$renderer3.push(`<!--]--></div> <div class="comment-body svelte-1eidjaf"><!--[-->`);
            const each_array_3 = ensure_array_like(bodyParts(c.body));
            for (let i = 0, $$length2 = each_array_3.length; i < $$length2; i++) {
              let part = each_array_3[i];
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
        $$renderer3.push(`<div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Status</span><span class="value svelte-1eidjaf">`);
        StatusPill($$renderer3, { status: liveStatus, size: "sm" });
        $$renderer3.push(`<!----></span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Exit code</span><span class="value svelte-1eidjaf">${escape_html("—")}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Commit</span><span class="value svelte-1eidjaf">`);
        if (data.run.commit_sha) {
          $$renderer3.push("<!--[0-->");
          Tooltip($$renderer3, {
            text: data.run.commit_sha,
            children: ($$renderer4) => {
              $$renderer4.push(`<span class="svelte-1eidjaf">${escape_html(shortSha$1(data.run.commit_sha))}</span>`);
            }
          });
        } else {
          $$renderer3.push("<!--[-1-->");
          $$renderer3.push(`—`);
        }
        $$renderer3.push(`<!--]--></span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Duration</span><span class="value svelte-1eidjaf">${escape_html(formatDuration(data.run.started_at, data.run.finished_at))}</span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Started</span><span class="value svelte-1eidjaf">`);
        TimeAgo($$renderer3, { value: data.run.started_at });
        $$renderer3.push(`<!----></span></div> <div class="row svelte-1eidjaf"><span class="label svelte-1eidjaf">Finished</span><span class="value svelte-1eidjaf">`);
        TimeAgo($$renderer3, { value: data.run.finished_at });
        $$renderer3.push(`<!----></span></div> `);
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
    $$renderer2.push(`<!----> `);
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<button type="button" class="ghost svelte-1eidjaf" style="padding: 0.25rem 0.5rem; font-size: var(--fs-xs);">Share</button>`);
      };
      Panel($$renderer2, {
        title: "Sharing",
        actions,
        children: ($$renderer3) => {
          if (data.shares.length === 0) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div class="empty-shares svelte-1eidjaf">No share links. Click Share to generate one.</div>`);
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`<!--[-->`);
            const each_array_4 = ensure_array_like(data.shares);
            for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
              let s = each_array_4[$$index_4];
              $$renderer3.push(`<div class="share-row svelte-1eidjaf"><span class="share-token svelte-1eidjaf">${escape_html(s.token.slice(0, 16))}...</span> <span class="share-meta svelte-1eidjaf">expires `);
              TimeAgo($$renderer3, { value: s.expires_at });
              $$renderer3.push(`<!----></span> <div class="share-actions svelte-1eidjaf">`);
              if (shareCopied === s.token) {
                $$renderer3.push("<!--[0-->");
                $$renderer3.push(`<span class="copy-ok svelte-1eidjaf">Copied</span>`);
              } else {
                $$renderer3.push("<!--[-1-->");
                $$renderer3.push(`<button type="button" class="ghost svelte-1eidjaf">Copy</button>`);
              }
              $$renderer3.push(`<!--]--> <form method="POST" action="?/revokeShare" class="inline-form svelte-1eidjaf"><input type="hidden" name="share_id"${attr("value", s.id)} class="svelte-1eidjaf"/> <button type="submit" class="ghost svelte-1eidjaf" title="Revoke this link" style="color: var(--danger-text);">Revoke</button></form></div></div>`);
            }
            $$renderer3.push(`<!--]-->`);
          }
          $$renderer3.push(`<!--]--> `);
          if (form?.shareError) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p class="modal-error svelte-1eidjaf" style="margin-top: var(--space-2);">${escape_html(form.shareError)}</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]-->`);
        }
      });
    }
    $$renderer2.push(`<!----> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div> `);
    Modal($$renderer2, {
      open: showShareModal,
      title: "Share Run",
      width: 440,
      onClose: () => showShareModal = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/share" class="svelte-1eidjaf"><p style="color: var(--text-muted); font-size: var(--fs-sm); margin: 0 0 var(--space-3) 0;" class="svelte-1eidjaf">Generate a shareable link for this run. Anyone with the link can view the run details and log.</p> <div class="field svelte-1eidjaf"><label for="share-expires" class="svelte-1eidjaf">Link expiry (hours)</label> <input id="share-expires" name="expires_hours" type="number" min="1" max="720" value="168" class="svelte-1eidjaf"/> <p class="modal-hint svelte-1eidjaf">1-720 hours (default 168 = 7 days). Leave blank for server default.</p></div> <div class="modal-actions svelte-1eidjaf"><button type="button" class="ghost svelte-1eidjaf">Cancel</button> <button type="submit" class="svelte-1eidjaf">Generate Link</button></div></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: showSignModal,
      title: "Sign Run",
      width: 520,
      onClose: () => showSignModal = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<form method="POST" action="?/sign" class="svelte-1eidjaf"><p style="color: var(--text-muted); font-size: var(--fs-sm); margin: 0 0 var(--space-3) 0;" class="svelte-1eidjaf">Cryptographically sign this run to create an immutable provenance record. The signature proves this run was reviewed and approved.</p> <div class="field svelte-1eidjaf"><label for="sign-key-id" class="svelte-1eidjaf">Key ID</label> <input id="sign-key-id" name="key_id" type="text" required="" placeholder="e.g. my-signing-key-2026" class="svelte-1eidjaf"/> <p class="modal-hint svelte-1eidjaf">Identifier for the signing key.</p></div> <div class="field svelte-1eidjaf"><label for="sign-private-key" class="svelte-1eidjaf">Private Key</label> <textarea id="sign-private-key" name="private_key" rows="6" required="" placeholder="Paste your PEM-encoded private key..." style="font-family: var(--font-mono); font-size: var(--fs-xs);" class="svelte-1eidjaf"></textarea> <p class="modal-hint svelte-1eidjaf">Your private key is sent to the server for signing and is not stored.</p></div> <div class="modal-actions svelte-1eidjaf"><button type="button" class="ghost svelte-1eidjaf">Cancel</button> <button type="submit" class="svelte-1eidjaf">Sign</button></div></form>`);
      }
    });
    $$renderer2.push(`<!----> `);
    {
      let footer = function($$renderer3) {
        $$renderer3.push(`<button type="button" class="ghost svelte-1eidjaf">Close</button>`);
      };
      Modal($$renderer2, {
        open: showProvenanceModal,
        title: "SLSA Provenance",
        width: 640,
        onClose: () => showProvenanceModal = false,
        footer,
        children: ($$renderer3) => {
          {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-4) 0; text-align: center;" class="svelte-1eidjaf">No provenance data available.</div>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
      });
    }
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-EuA9w37s.js.map
