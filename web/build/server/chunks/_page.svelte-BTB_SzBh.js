import { a8 as escape_html, ac as ensure_array_like, a7 as attr } from './renderer-D0X3o35U.js';
import { o as onDestroy } from './index-server-BFLhAcPs.js';
import './root-BPP4VMgG.js';
import './state.svelte-DaEZ9E9Z.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { c as formatDuration } from './utils2-BUPlP7zG.js';
import './Tooltip-Bg85MZ7d.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    onDestroy(() => {
    });
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        { label: `pipeline run ${data.pipelineRun.id.slice(0, 8)}` }
      ]
    });
    $$renderer2.push(`<!----> <div class="hero svelte-14esf8o"><div class="hero-main svelte-14esf8o"><div class="hero-line svelte-14esf8o">`);
    StatusPill($$renderer2, { status: data.pipelineRun.status });
    $$renderer2.push(`<!----> <h1 class="svelte-14esf8o">Pipeline run</h1></div> <div class="hero-meta svelte-14esf8o"><span class="hero-meta-item svelte-14esf8o"><span class="hero-meta-label svelte-14esf8o">Steps</span> <strong class="svelte-14esf8o">${escape_html((data.pipelineRun.children ?? []).length)}</strong></span> <span class="hero-sep svelte-14esf8o">·</span> <span class="hero-meta-item svelte-14esf8o"><span class="hero-meta-label svelte-14esf8o">Duration</span> <strong class="svelte-14esf8o">${escape_html(formatDuration(data.pipelineRun.started_at, data.pipelineRun.finished_at))}</strong></span> <span class="hero-sep svelte-14esf8o">·</span> <span class="hero-meta-item svelte-14esf8o"><span class="hero-meta-label svelte-14esf8o">Started</span> <strong class="svelte-14esf8o">`);
    TimeAgo($$renderer2, {
      value: data.pipelineRun.started_at ?? data.pipelineRun.created_at
    });
    $$renderer2.push(`<!----></strong></span></div></div></div> <section class="step-list svelte-14esf8o"><!--[-->`);
    const each_array = ensure_array_like(data.pipelineRun.children ?? []);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let c = each_array[$$index];
      $$renderer2.push(`<a${attr("href", `/runs/${c.run_id}`)} class="step-row svelte-14esf8o"><span class="step-pos svelte-14esf8o">#${escape_html(c.step + 1)}</span> `);
      StatusPill($$renderer2, { status: c.status, size: "sm" });
      $$renderer2.push(`<!----> <span class="step-task svelte-14esf8o">${escape_html(c.task_name || c.run_id.slice(0, 8))}</span> <span class="step-meta svelte-14esf8o">view log →</span> <span class="arrow svelte-14esf8o">→</span></a>`);
    }
    $$renderer2.push(`<!--]--> `);
    if ((data.pipelineRun.children ?? []).length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div style="padding: var(--space-5); color: var(--text-dim); text-align: center;">No steps run yet.</div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></section>`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BTB_SzBh.js.map
