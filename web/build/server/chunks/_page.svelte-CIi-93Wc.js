import { ac as ensure_array_like, a7 as attr, a9 as attr_class, a8 as escape_html, ad as stringify, a2 as derived } from './renderer-D0X3o35U.js';
import './root-BPP4VMgG.js';
import './state.svelte-DaEZ9E9Z.js';
import './client-Bp0-LT_C.js';
import { c as formatDuration, s as shortSha } from './utils2-BUPlP7zG.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { T as Tooltip } from './Tooltip-Bg85MZ7d.js';
import { S as SectionHeader } from './SectionHeader-BsGXUHV-.js';
import './index-rPPP9l7g.js';
import './index-server-BFLhAcPs.js';

function FilterPills($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { options, value, label } = $$props;
    $$renderer2.push(`<div class="filters svelte-gzmvqe" role="tablist"${attr("aria-label", label)}>`);
    if (label) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="label svelte-gzmvqe">${escape_html(label)}</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <!--[-->`);
    const each_array = ensure_array_like(options);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let o = each_array[$$index];
      $$renderer2.push(`<button type="button" role="tab"${attr_class(`pill ${stringify(value === o.value ? "active" : "")}`, "svelte-gzmvqe")}${attr("aria-selected", value === o.value)}>${escape_html(o.label)} `);
      if (o.count != null) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<span class="count svelte-gzmvqe">${escape_html(o.count)}</span>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></button>`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
function DensityToggle($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let density = "comfortable";
    $$renderer2.push(`<div class="dt svelte-fslimw" role="group" aria-label="Density"><button type="button"${attr_class("dt-btn svelte-fslimw", void 0, { "active": density === "comfortable" })}${attr("aria-pressed", density === "comfortable")} title="Comfortable density">≡</button> <button type="button"${attr_class("dt-btn svelte-fslimw", void 0, { "active": density === "compact" })}${attr("aria-pressed", density === "compact")} title="Compact density">☰</button></div>`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    let density = "comfortable";
    let counts = derived(() => {
      const c = { "": data.runs.length };
      for (const r of data.runs) c[r.status] = (c[r.status] || 0) + 1;
      return c;
    });
    let filterOptions = derived(() => [
      { value: "", label: "All", count: counts()[""] || 0 },
      {
        value: "succeeded",
        label: "Succeeded",
        count: counts()["succeeded"] || 0
      },
      {
        value: "failed",
        label: "Failed",
        count: counts()["failed"] || 0
      },
      {
        value: "running",
        label: "Running",
        count: counts()["running"] || 0
      },
      {
        value: "cancelled",
        label: "Cancelled",
        count: counts()["cancelled"] || 0
      }
    ]);
    SectionHeader($$renderer2, { title: "Run history" });
    $$renderer2.push(`<!----> <div class="toolbar svelte-1b889ak">`);
    FilterPills($$renderer2, {
      options: filterOptions(),
      value: data.filter || ""
    });
    $$renderer2.push(`<!----> <div class="toolbar-actions svelte-1b889ak">`);
    DensityToggle($$renderer2);
    $$renderer2.push(`<!----></div></div> `);
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
        $$renderer2.push(`<a${attr("href", `/runs/${r.id}`)}${attr_class("run-row svelte-1b889ak", void 0, { "compact": density === "compact" })}>`);
        StatusPill($$renderer2, { status: r.status, size: "sm" });
        $$renderer2.push(`<!----> <span class="run-name svelte-1b889ak">${escape_html(r.task_name)} <span class="dim svelte-1b889ak">(${escape_html(r.task_source)})</span> `);
        if (r.branch) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="branch-tag svelte-1b889ak"${attr("title", `Ran on branch ${r.branch}`)}>${escape_html(r.branch)}</span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <span class="mobile-only run-stack-mobile svelte-1b889ak"><span class="run-meta svelte-1b889ak">${escape_html(formatDuration(r.started_at, r.finished_at))} · `);
        TimeAgo($$renderer2, { value: r.created_at });
        $$renderer2.push(`<!----></span></span></span> `);
        Tooltip($$renderer2, {
          text: r.commit_sha || "no commit",
          children: ($$renderer3) => {
            $$renderer3.push(`<span class="run-meta svelte-1b889ak">${escape_html(shortSha(r.commit_sha))}</span>`);
          }
        });
        $$renderer2.push(`<!----> <span${attr_class("run-meta svelte-1b889ak", void 0, { "exit-bad": r.exit_code != null && r.exit_code !== 0 })}>${escape_html(r.exit_code != null ? `exit ${r.exit_code}` : "—")}</span> <span class="run-meta svelte-1b889ak">${escape_html(formatDuration(r.started_at, r.finished_at))}</span> <span class="run-meta svelte-1b889ak">`);
        TimeAgo($$renderer2, { value: r.created_at });
        $$renderer2.push(`<!----></span></a>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-CIi-93Wc.js.map
