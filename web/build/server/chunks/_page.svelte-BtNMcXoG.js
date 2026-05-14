import { a8 as escape_html, a7 as attr, ac as ensure_array_like, a9 as attr_class, af as clsx, ae as attr_style, ad as stringify, a2 as derived } from './renderer-D0X3o35U.js';
import './root-BPP4VMgG.js';
import './state.svelte-DaEZ9E9Z.js';
import './client-Bp0-LT_C.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { S as SectionHeader } from './SectionHeader-BsGXUHV-.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { F as FlashMessage } from './FlashMessage-CReJ7kuH.js';
import './index-rPPP9l7g.js';
import './Tooltip-Bg85MZ7d.js';
import './index-server-BFLhAcPs.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    function fmtDur(sec) {
      if (sec < 60) return `${Math.round(sec)}s`;
      if (sec < 3600) return `${Math.floor(sec / 60)}m ${Math.round(sec % 60)}s`;
      return `${Math.floor(sec / 3600)}h ${Math.floor(sec % 3600 / 60)}m`;
    }
    let maxDay = derived(() => {
      let m = 0;
      for (const b of data.summary?.daily_runs ?? []) if (b.total > m) m = b.total;
      return Math.max(1, m);
    });
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        {
          label: data.workspace.name,
          href: `/workspaces/${data.workspace.id}`
        },
        { label: "dashboard" }
      ]
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-1xr8ys0"><h1 class="svelte-1xr8ys0">${escape_html(data.workspace.name)}</h1></div> <div class="ws-tabs svelte-1xr8ys0"><a class="ws-tab svelte-1xr8ys0"${attr("href", `/workspaces/${data.workspace.id}`)}>Overview</a> <a class="ws-tab active svelte-1xr8ys0"${attr("href", `/workspaces/${data.workspace.id}/dashboard`)}>Dashboard</a></div> <div class="controls svelte-1xr8ys0"><span class="svelte-1xr8ys0">Window:</span> <!--[-->`);
    const each_array = ensure_array_like([7, 30, 90]);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let d = each_array[$$index];
      $$renderer2.push(`<button${attr_class("pill svelte-1xr8ys0", void 0, { "active": data.days === d })}>${escape_html(d)} days</button>`);
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.summaryError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.summaryError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.summary) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="stats svelte-1xr8ys0"><div class="stat svelte-1xr8ys0"><div class="stat-num svelte-1xr8ys0">${escape_html(data.summary.totals.projects)}</div> <div class="stat-label svelte-1xr8ys0">Projects</div></div> <div class="stat svelte-1xr8ys0"><div class="stat-num svelte-1xr8ys0">${escape_html(data.summary.totals.members)}</div> <div class="stat-label svelte-1xr8ys0">Members</div></div> <div class="stat svelte-1xr8ys0"><div class="stat-num svelte-1xr8ys0">${escape_html(data.summary.totals.runs_in_window.toLocaleString())}</div> <div class="stat-label svelte-1xr8ys0">Runs</div></div> <div class="stat svelte-1xr8ys0"><div class="stat-num svelte-1xr8ys0"><span${attr_class(clsx(data.summary.totals.failed_in_window > 0 ? "bad" : ""), "svelte-1xr8ys0")}>${escape_html(data.summary.totals.failed_in_window)}</span></div> <div class="stat-label svelte-1xr8ys0">Failed</div></div> <div class="stat svelte-1xr8ys0"><div class="stat-num svelte-1xr8ys0"><span${attr_class(clsx(data.summary.totals.success_rate >= 0.9 ? "ok" : "bad"), "svelte-1xr8ys0")}>${escape_html(Math.round((data.summary.totals.success_rate || 0) * 100))}%</span></div> <div class="stat-label svelte-1xr8ys0">Success rate</div></div> <div class="stat svelte-1xr8ys0"><div class="stat-num svelte-1xr8ys0">${escape_html(data.summary.totals.active_now)}</div> <div class="stat-label svelte-1xr8ys0">Active now</div></div></div> <div class="grid-2 svelte-1xr8ys0"><section class="panel svelte-1xr8ys0"><h2 class="svelte-1xr8ys0">Slowest tasks</h2> `);
      if (data.summary.slowest_tasks.length === 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div style="color: var(--text-dim); font-size: var(--fs-sm); padding: var(--space-2) 0;">Not enough data yet.</div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<!--[-->`);
        const each_array_1 = ensure_array_like(data.summary.slowest_tasks);
        for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
          let t = each_array_1[$$index_1];
          $$renderer2.push(`<a class="row svelte-1xr8ys0"${attr("href", `/projects/${t.project_id}/runs`)}><span></span> <span><span class="name svelte-1xr8ys0">${escape_html(t.task_name)}</span> <div class="sub svelte-1xr8ys0">${escape_html(t.project_name)} · ${escape_html(t.task_source)} · ${escape_html(t.runs)} runs</div></span> <span class="meta svelte-1xr8ys0">${escape_html(fmtDur(t.avg_seconds))} avg</span></a>`);
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--></section> <section class="panel svelte-1xr8ys0"><h2 class="svelte-1xr8ys0">Most failing tasks</h2> `);
      if (data.summary.failing_tasks.length === 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div style="color: var(--text-dim); font-size: var(--fs-sm); padding: var(--space-2) 0;">No failures in window.</div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<!--[-->`);
        const each_array_2 = ensure_array_like(data.summary.failing_tasks);
        for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
          let t = each_array_2[$$index_2];
          $$renderer2.push(`<a class="row svelte-1xr8ys0"${attr("href", `/projects/${t.project_id}/runs?status=failed`)}><span></span> <span><span class="name svelte-1xr8ys0">${escape_html(t.task_name)}</span> <div class="sub svelte-1xr8ys0">${escape_html(t.project_name)} · ${escape_html(t.task_source)}</div></span> <span class="meta bad svelte-1xr8ys0">${escape_html(Math.round(t.failure_rate * 100))}% · ${escape_html(t.failures)}/${escape_html(t.total)}</span></a>`);
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]--></section></div> <section class="panel svelte-1xr8ys0" style="margin-bottom: var(--space-5);"><h2 class="svelte-1xr8ys0">Run volume</h2> `);
      if (data.summary.daily_runs.length === 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div style="color: var(--text-dim); font-size: var(--fs-sm); padding: var(--space-2) 0;">No runs in window.</div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<div class="runs-bar-grid svelte-1xr8ys0"><!--[-->`);
        const each_array_3 = ensure_array_like(data.summary.daily_runs);
        for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
          let b = each_array_3[$$index_3];
          $$renderer2.push(`<div class="runs-bar-row svelte-1xr8ys0"><span>${escape_html(b.date.slice(5))}</span> <div class="runs-bar-track svelte-1xr8ys0"><div class="runs-bar-ok svelte-1xr8ys0"${attr_style(`width: ${stringify(b.succeeded / maxDay() * 100)}%;`)}></div> <div class="runs-bar-bad svelte-1xr8ys0"${attr_style(`width: ${stringify(b.failed / maxDay() * 100)}%;`)}></div></div> <span class="runs-bar-amount svelte-1xr8ys0">${escape_html(b.total)}</span></div>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      }
      $$renderer2.push(`<!--]--></section> `);
      SectionHeader($$renderer2, { title: "Projects health" });
      $$renderer2.push(`<!----> <div class="health-list svelte-1xr8ys0"><!--[-->`);
      const each_array_4 = ensure_array_like(data.summary.projects_health);
      for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
        let p = each_array_4[$$index_4];
        $$renderer2.push(`<a class="health-row svelte-1xr8ys0"${attr("href", `/projects/${p.project_id}`)}>`);
        StatusPill($$renderer2, { status: p.status, size: "sm" });
        $$renderer2.push(`<!----> <span class="health-name svelte-1xr8ys0">${escape_html(p.project_name)}</span> <span class="health-meta svelte-1xr8ys0">${escape_html(p.runs_in_window)} runs</span> <span${attr_class(`health-meta ${stringify(p.failure_rate >= 0.5 ? "bad" : p.failure_rate >= 0.1 ? "warn" : "")}`, "svelte-1xr8ys0")}>${escape_html(p.runs_in_window > 0 ? `${Math.round(p.failure_rate * 100)}% fail` : "—")}</span> <span class="health-meta svelte-1xr8ys0">`);
        if (p.last_run_at) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`last `);
          TimeAgo($$renderer2, { value: p.last_run_at });
          $$renderer2.push(`<!----> · ${escape_html(p.last_run_status ?? "")}`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`no runs`);
        }
        $$renderer2.push(`<!--]--></span></a>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BtNMcXoG.js.map
