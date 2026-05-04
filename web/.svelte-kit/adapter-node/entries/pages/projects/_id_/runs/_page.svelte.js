import { b as attr_class, s as stringify, c as ensure_array_like, a as attr, f as attr_style, e as escape_html } from "../../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../../chunks/exports.js";
import "../../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../../chunks/root.js";
import "../../../../../chunks/state.svelte.js";
import "../../../../../chunks/client.js";
import { a as statusColor, s as shortSha, c as formatDuration } from "../../../../../chunks/utils2.js";
import { E as EmptyState } from "../../../../../chunks/EmptyState.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    $$renderer2.push(`<h2 class="section-title svelte-1b889ak">Run history</h2> <div class="filters svelte-1b889ak"><span class="filter-label svelte-1b889ak">Filter:</span> <button${attr_class(`pill ${stringify(data.filter === "" ? "active" : "")}`, "svelte-1b889ak")}>All</button> <button${attr_class(`pill ${stringify(data.filter === "succeeded" ? "active" : "")}`, "svelte-1b889ak")}>Succeeded</button> <button${attr_class(`pill ${stringify(data.filter === "failed" ? "active" : "")}`, "svelte-1b889ak")}>Failed</button> <button${attr_class(`pill ${stringify(data.filter === "cancelled" ? "active" : "")}`, "svelte-1b889ak")}>Cancelled</button> <button${attr_class(`pill ${stringify(data.filter === "running" ? "active" : "")}`, "svelte-1b889ak")}>Running</button></div> `);
    if (data.runs.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        children: ($$renderer3) => {
          if (data.filter) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`No runs matching <strong>${escape_html(data.filter)}</strong>.`);
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`No runs yet.`);
          }
          $$renderer3.push(`<!--]-->`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(data.runs);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let r = each_array[$$index];
        $$renderer2.push(`<a${attr("href", `/runs/${r.id}`)} class="run-row svelte-1b889ak"><span class="dot svelte-1b889ak"${attr_style(`background: ${stringify(statusColor(r.status))}`)}></span> <span class="run-name svelte-1b889ak">${escape_html(r.task_name)} <span class="dim svelte-1b889ak">(${escape_html(r.task_source)})</span></span> <span class="run-meta svelte-1b889ak">${escape_html(shortSha(r.commit_sha))}</span> <span class="run-meta svelte-1b889ak">${escape_html(r.exit_code ?? "—")}</span> <span class="run-meta svelte-1b889ak">${escape_html(formatDuration(r.started_at, r.finished_at))}</span> <span class="run-meta svelte-1b889ak">${escape_html(new Date(r.created_at).toLocaleString())}</span></a>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
