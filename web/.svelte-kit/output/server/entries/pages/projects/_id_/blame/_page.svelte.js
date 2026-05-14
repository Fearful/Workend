import { a as attr, e as escape_html, c as ensure_array_like, b as attr_class, s as stringify } from "../../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../../chunks/exports.js";
import "../../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../../chunks/root.js";
import "../../../../../chunks/state.svelte.js";
/* empty css                                                        */
import { S as StatusPill } from "../../../../../chunks/StatusPill.js";
import { M as Modal } from "../../../../../chunks/Modal.js";
import { T as TimeAgo } from "../../../../../chunks/TimeAgo.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
import { c as formatDuration } from "../../../../../chunks/utils2.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let syncPending = false;
    let detailModal = null;
    function rowClass(c) {
      if (c.run_count === 0) return "";
      if (c.fail_count > 0) return "row-fail";
      return "row-pass";
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
    $$renderer2.push(`<!--]--> <div class="page-header svelte-1ksg60r"><h2 class="svelte-1ksg60r">Blame timeline</h2> <button type="button"${attr("disabled", syncPending, true)}>${escape_html("Sync commits")}</button></div> `);
    if (data.blameTimeline.length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="empty svelte-1ksg60r"><p>No commits synced yet. Click "Sync commits" to pull commit history and correlate with runs.</p></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="table-wrap svelte-1ksg60r"><table class="svelte-1ksg60r"><thead><tr><th class="svelte-1ksg60r">SHA</th><th class="svelte-1ksg60r">Author</th><th class="svelte-1ksg60r">Message</th><th class="svelte-1ksg60r">Date</th><th class="num svelte-1ksg60r">Files</th><th class="num svelte-1ksg60r">+/-</th><th class="num svelte-1ksg60r">Runs</th><th class="num svelte-1ksg60r">Avg</th></tr></thead><tbody><!--[-->`);
      const each_array = ensure_array_like(data.blameTimeline);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let c = each_array[$$index];
        $$renderer2.push(`<tr${attr_class(`clickable ${stringify(rowClass(c))}`, "svelte-1ksg60r")} role="button" tabindex="0"><td class="sha svelte-1ksg60r">${escape_html(c.sha.slice(0, 8))}</td><td class="author svelte-1ksg60r">${escape_html(c.author)}</td><td class="msg svelte-1ksg60r">${escape_html(c.message.split("\n")[0].slice(0, 80))}</td><td class="date svelte-1ksg60r">`);
        TimeAgo($$renderer2, { value: c.committed_at });
        $$renderer2.push(`<!----></td><td class="num svelte-1ksg60r">${escape_html(c.files_changed)}</td><td class="num svelte-1ksg60r"><span class="ins svelte-1ksg60r">+${escape_html(c.insertions)}</span> <span class="del svelte-1ksg60r">-${escape_html(c.deletions)}</span></td><td class="num svelte-1ksg60r">`);
        if (c.run_count > 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="runs-cell svelte-1ksg60r">${escape_html(c.run_count)} `);
          if (c.pass_count > 0) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="pass-count svelte-1ksg60r">${escape_html(c.pass_count)}p</span>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--> `);
          if (c.fail_count > 0) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="fail-count svelte-1ksg60r">${escape_html(c.fail_count)}f</span>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<span class="dim svelte-1ksg60r">--</span>`);
        }
        $$renderer2.push(`<!--]--></td><td class="num svelte-1ksg60r">`);
        if (c.avg_duration_ms > 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`${escape_html(c.avg_duration_ms < 1e3 ? `${Math.round(c.avg_duration_ms)}ms` : `${(c.avg_duration_ms / 1e3).toFixed(1)}s`)}`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<span class="dim svelte-1ksg60r">--</span>`);
        }
        $$renderer2.push(`<!--]--></td></tr>`);
      }
      $$renderer2.push(`<!--]--></tbody></table></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    Modal($$renderer2, {
      open: detailModal !== null,
      title: detailModal ? `Commit ${detailModal.sha.slice(0, 8)}` : "",
      width: 640,
      onClose: () => detailModal = null,
      children: ($$renderer3) => {
        if (detailModal) {
          $$renderer3.push("<!--[0-->");
          if (detailModal.loading) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div style="color: var(--text-dim); font-style: italic; padding: var(--space-3) 0;">Loading...</div>`);
          } else if (detailModal.error) {
            $$renderer3.push("<!--[1-->");
            $$renderer3.push(`<p class="modal-error svelte-1ksg60r">${escape_html(detailModal.error)}</p>`);
          } else if (detailModal.commit) {
            $$renderer3.push("<!--[2-->");
            $$renderer3.push(`<div class="detail-row svelte-1ksg60r"><span class="detail-label svelte-1ksg60r">SHA</span><span class="detail-value svelte-1ksg60r">${escape_html(detailModal.commit.sha)}</span></div> <div class="detail-row svelte-1ksg60r"><span class="detail-label svelte-1ksg60r">Author</span><span class="detail-value detail-msg svelte-1ksg60r">${escape_html(detailModal.commit.author)} &lt;${escape_html(detailModal.commit.author_email)}></span></div> <div class="detail-row svelte-1ksg60r"><span class="detail-label svelte-1ksg60r">Date</span><span class="detail-value svelte-1ksg60r">${escape_html(new Date(detailModal.commit.committed_at).toLocaleString())}</span></div> <div class="detail-row svelte-1ksg60r"><span class="detail-label svelte-1ksg60r">Message</span><span class="detail-value detail-msg svelte-1ksg60r">${escape_html(detailModal.commit.message)}</span></div> <div class="detail-row svelte-1ksg60r"><span class="detail-label svelte-1ksg60r">Changes</span><span class="detail-value svelte-1ksg60r">${escape_html(detailModal.commit.files_changed)} files, <span class="ins svelte-1ksg60r">+${escape_html(detailModal.commit.insertions)}</span> <span class="del svelte-1ksg60r">-${escape_html(detailModal.commit.deletions)}</span></span></div> `);
            if (detailModal.runs.length > 0) {
              $$renderer3.push("<!--[0-->");
              $$renderer3.push(`<h3 style="margin: var(--space-4) 0 var(--space-2) 0; font-size: 0.875rem; color: var(--text-muted);">Associated runs (${escape_html(detailModal.runs.length)})</h3> <!--[-->`);
              const each_array_1 = ensure_array_like(detailModal.runs);
              for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
                let r = each_array_1[$$index_1];
                $$renderer3.push(`<a${attr("href", `/runs/${r.id}`)} class="run-link svelte-1ksg60r">`);
                StatusPill($$renderer3, { status: r.status, size: "sm" });
                $$renderer3.push(`<!----> <span class="run-meta svelte-1ksg60r">${escape_html(formatDuration(r.started_at, r.finished_at))}</span> <span class="run-meta svelte-1ksg60r">${escape_html(r.started_at ? new Date(r.started_at).toLocaleString() : "--")}</span></a>`);
              }
              $$renderer3.push(`<!--]-->`);
            } else {
              $$renderer3.push("<!--[-1-->");
              $$renderer3.push(`<p style="color: var(--text-dim); font-size: 0.875rem; margin-top: var(--space-3);">No runs associated with this commit.</p>`);
            }
            $$renderer3.push(`<!--]-->`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> <div class="modal-actions svelte-1ksg60r"><button type="button" class="ghost">Close</button></div>`);
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
