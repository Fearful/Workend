import { c as ensure_array_like, e as escape_html, a as attr, b as attr_class, s as stringify, f as attr_style, d as derived } from "../../../chunks/renderer.js";
import { b as freshnessClass, a as statusColor, f as formatRelative } from "../../../chunks/utils2.js";
import { P as PageHeader } from "../../../chunks/PageHeader.js";
import { S as StatusDot } from "../../../chunks/StatusDot.js";
import { B as Badge } from "../../../chunks/Badge.js";
import { E as EmptyState } from "../../../chunks/EmptyState.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    let groupedByWorkspace = derived(() => {
      const groups = {};
      for (const c of data.cards) {
        if (!groups[c.workspace_name]) groups[c.workspace_name] = [];
        groups[c.workspace_name].push(c);
      }
      return groups;
    });
    let stats = derived(() => {
      const running = data.cards.filter((c) => c.last_run_status === "running" || c.last_run_status === "queued").length;
      const failing = data.cards.filter((c) => c.has_failing_recent_run).length;
      return { total: data.cards.length, running, failing };
    });
    PageHeader($$renderer2, { title: "Dashboard" });
    $$renderer2.push(`<!----> <div class="layout svelte-x1i5gj"><div class="main-content svelte-x1i5gj">`);
    if (data.cards.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "◇",
        message: "No projects yet. Create a workspace and add some.",
        actionHref: "/",
        actionLabel: "Browse workspaces"
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(Object.entries(groupedByWorkspace()));
      for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
        let [wsName, cards] = each_array[$$index_1];
        $$renderer2.push(`<h2 class="svelte-x1i5gj">${escape_html(wsName)}</h2> <div class="grid svelte-x1i5gj"><!--[-->`);
        const each_array_1 = ensure_array_like(cards);
        for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
          let c = each_array_1[$$index];
          $$renderer2.push(`<a${attr("href", `/projects/${c.project_id}`)}${attr_class(`card ${stringify(c.has_failing_recent_run ? "alert" : "")}`, "svelte-x1i5gj")}><div class="card-head svelte-x1i5gj">`);
          StatusDot($$renderer2, { status: c.status });
          $$renderer2.push(`<!----> <span class="card-name svelte-x1i5gj">${escape_html(c.project_name)}</span> `);
          if (c.has_failing_recent_run) {
            $$renderer2.push("<!--[0-->");
            Badge($$renderer2, {
              variant: "danger",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->recent failure`);
              }
            });
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></div> `);
          if (c.last_commit_message) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<div class="commit-line svelte-x1i5gj">${escape_html(c.last_commit_message)}</div>`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--> <div class="meta-grid svelte-x1i5gj"><span class="label svelte-x1i5gj">Status</span><span class="value svelte-x1i5gj">${escape_html(c.status)}</span> <span class="label svelte-x1i5gj">Tasks</span><span class="value svelte-x1i5gj">${escape_html(c.task_count)}</span> <span class="label svelte-x1i5gj">Lang</span><span class="value svelte-x1i5gj">${escape_html(c.top_language || "—")}</span> <span class="label svelte-x1i5gj">Lines</span><span class="value svelte-x1i5gj">${escape_html(c.total_lines ? c.total_lines.toLocaleString() : "—")}</span> <span class="label svelte-x1i5gj">Last sync</span> `);
          if (freshnessClass(c.last_synced_at) === "fresh-good") {
            $$renderer2.push("<!--[0-->");
            Badge($$renderer2, {
              variant: "success",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(formatRelative(c.last_synced_at))}`);
              }
            });
          } else if (freshnessClass(c.last_synced_at) === "fresh-warn") {
            $$renderer2.push("<!--[1-->");
            Badge($$renderer2, {
              variant: "warning",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(formatRelative(c.last_synced_at))}`);
              }
            });
          } else {
            $$renderer2.push("<!--[-1-->");
            Badge($$renderer2, {
              variant: "muted",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(formatRelative(c.last_synced_at))}`);
              }
            });
          }
          $$renderer2.push(`<!--]--> <span class="label svelte-x1i5gj">Last run</span> `);
          if (c.last_run_status) {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="value svelte-x1i5gj"${attr_style(`color: ${stringify(statusColor(c.last_run_status))};`)}>${escape_html(c.last_run_task_name)} · ${escape_html(c.last_run_status)}</span>`);
          } else {
            $$renderer2.push("<!--[-1-->");
            $$renderer2.push(`<span class="value svelte-x1i5gj">never</span>`);
          }
          $$renderer2.push(`<!--]--></div></a>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div> <aside class="sidebar svelte-x1i5gj">`);
    if (data.cards.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-x1i5gj">Overview</h2> <div class="stats-grid svelte-x1i5gj"><div class="stat-tile svelte-x1i5gj"><div class="stat-num svelte-x1i5gj">${escape_html(stats().total)}</div> <div class="stat-label svelte-x1i5gj">Projects</div></div> <div class="stat-tile svelte-x1i5gj"><div class="stat-num warning svelte-x1i5gj">${escape_html(stats().running)}</div> <div class="stat-label svelte-x1i5gj">Running</div></div> <div class="stat-tile svelte-x1i5gj"><div class="stat-num danger svelte-x1i5gj">${escape_html(stats().failing)}</div> <div class="stat-label svelte-x1i5gj">Failing</div></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.pinned.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-x1i5gj">Pinned tasks</h2> <div class="pinned-grid svelte-x1i5gj"><!--[-->`);
      const each_array_2 = ensure_array_like(data.pinned);
      for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
        let p = each_array_2[$$index_2];
        $$renderer2.push(`<div class="pin-card svelte-x1i5gj"><div class="pin-head svelte-x1i5gj"><span class="pin-task svelte-x1i5gj">${escape_html(p.task_name)}</span> <span class="pin-source svelte-x1i5gj">${escape_html(p.task_source)}</span></div> <div class="pin-loc svelte-x1i5gj"><a${attr("href", `/projects/${p.project_id}`)} class="svelte-x1i5gj">${escape_html(p.workspace_name)} / ${escape_html(p.project_name)}</a></div> <div class="pin-actions svelte-x1i5gj"><form method="POST" action="?/runPinned" class="pin-run-form svelte-x1i5gj"><input type="hidden" name="task_id"${attr("value", p.task_id)}/> <button type="submit" class="svelte-x1i5gj">Run</button></form> <form method="POST" action="?/unpin" class="svelte-x1i5gj"><input type="hidden" name="task_id"${attr("value", p.task_id)}/> <button type="submit" class="ghost" title="Unpin" aria-label="Unpin">★</button></form></div></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.flaky.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="warn svelte-x1i5gj">Flaky tasks</h2> <div class="flaky-grid svelte-x1i5gj"><!--[-->`);
      const each_array_3 = ensure_array_like(data.flaky);
      for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
        let f = each_array_3[$$index_3];
        $$renderer2.push(`<a${attr("href", `/projects/${f.project_id}/trends?task_id=${f.task_id}`)} class="flaky-card svelte-x1i5gj"><div class="flaky-head svelte-x1i5gj"><span class="flaky-task svelte-x1i5gj">${escape_html(f.task_name)}</span> <span class="flaky-source svelte-x1i5gj">${escape_html(f.task_source)}</span></div> <div class="flaky-loc svelte-x1i5gj">${escape_html(f.workspace_name)} / ${escape_html(f.project_name)}</div> <div class="flaky-stats svelte-x1i5gj"><span><strong class="flaky-flips svelte-x1i5gj">${escape_html(f.flips)}</strong> flips / ${escape_html(f.total)} runs</span> <span>${escape_html(Math.round(f.success_rate * 100))}% success</span></div></a>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></aside></div>`);
  });
}
export {
  _page as default
};
