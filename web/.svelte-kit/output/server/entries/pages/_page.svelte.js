import { a as attr, e as escape_html, c as ensure_array_like, b as attr_class, s as stringify, f as attr_style, d as derived } from "../../chunks/renderer.js";
import { P as PageHeader } from "../../chunks/PageHeader.js";
/* empty css                                               */
import { B as Badge } from "../../chunks/Badge.js";
import { E as EmptyState } from "../../chunks/EmptyState.js";
import { F as FlashMessage } from "../../chunks/FlashMessage.js";
import { S as StatusPill } from "../../chunks/StatusPill.js";
import { T as TimeAgo } from "../../chunks/TimeAgo.js";
import { S as SectionHeader } from "../../chunks/SectionHeader.js";
function RunRow($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { run, showLocation = true } = $$props;
    $$renderer2.push(`<a${attr("href", `/runs/${run.id}`)} class="run-row svelte-1n4sxwt">`);
    StatusPill($$renderer2, { status: run.status, size: "sm" });
    $$renderer2.push(`<!----> <div class="run-info"><div class="run-task svelte-1n4sxwt">${escape_html(run.task_name)} <span class="dim svelte-1n4sxwt">(${escape_html(run.task_source)})</span></div> `);
    if (showLocation && run.workspace_name && run.project_name) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="run-loc svelte-1n4sxwt">${escape_html(run.workspace_name)} / ${escape_html(run.project_name)}</div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (run.exit_code != null && run.exit_code !== 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="run-meta exit-bad svelte-1n4sxwt">exit ${escape_html(run.exit_code)}</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<span class="run-meta svelte-1n4sxwt"></span>`);
    }
    $$renderer2.push(`<!--]--> <span class="run-meta svelte-1n4sxwt">`);
    TimeAgo($$renderer2, { value: run.created_at });
    $$renderer2.push(`<!----></span></a>`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    let comparisonOpen = false;
    let comparisonData = derived(() => data.comparison);
    function healthColor(failRate) {
      if (failRate < 10) return "success";
      if (failRate <= 30) return "warning";
      return "danger";
    }
    function healthLabel(failRate) {
      if (failRate < 10) return "Healthy";
      if (failRate <= 30) return "Warning";
      return "Critical";
    }
    function severityVariant(severity) {
      const s = severity.toLowerCase();
      if (s === "critical") return "danger";
      if (s === "high") return "warning";
      if (s === "medium") return "warning";
      return "muted";
    }
    function storageMax(items) {
      let m = 0;
      for (const s of items) {
        const total = s.projects + s.tasks + s.runs + s.artifacts;
        if (total > m) m = total;
      }
      return m || 1;
    }
    let activeIncidents = derived(() => data.incidents.filter((i) => i.incident_count > 0));
    {
      let actions = function($$renderer3) {
        $$renderer3.push(`<a href="/workspaces/new"><button>New workspace</button></a>`);
      };
      PageHeader($$renderer2, { title: "Workspaces", actions });
    }
    $$renderer2.push(`<!----> `);
    if (data.error) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.error)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (activeIncidents().length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="incident-banner svelte-1uha8ag"><p class="incident-banner-title svelte-1uha8ag">Open incidents</p> <!--[-->`);
      const each_array = ensure_array_like(activeIncidents());
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let inc = each_array[$$index];
        $$renderer2.push(`<div class="incident-item svelte-1uha8ag">`);
        Badge($$renderer2, {
          variant: severityVariant(inc.worst_severity),
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(inc.worst_severity)}`);
          }
        });
        $$renderer2.push(`<!----> <span>${escape_html(inc.workspace_name)} — ${escape_html(inc.incident_count)} incident${escape_html(inc.incident_count === 1 ? "" : "s")}</span></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.health.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="section svelte-1uha8ag">`);
      SectionHeader($$renderer2, { title: "Workspace health" });
      $$renderer2.push(`<!----> <div class="health-grid svelte-1uha8ag"><!--[-->`);
      const each_array_1 = ensure_array_like(data.health);
      for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
        let h = each_array_1[$$index_1];
        $$renderer2.push(`<div class="health-card svelte-1uha8ag"><div class="health-header svelte-1uha8ag"><span class="health-name svelte-1uha8ag">${escape_html(h.workspace_name)}</span> `);
        Badge($$renderer2, {
          variant: healthColor(h.fail_rate_7d),
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(healthLabel(h.fail_rate_7d))}`);
          }
        });
        $$renderer2.push(`<!----></div> <div class="health-stats svelte-1uha8ag"><div class="health-stat svelte-1uha8ag"><div class="health-stat-value svelte-1uha8ag">${escape_html(h.member_count)}</div> <div class="health-stat-label svelte-1uha8ag">Members</div></div> <div class="health-stat svelte-1uha8ag"><div class="health-stat-value svelte-1uha8ag">${escape_html(h.project_count)}</div> <div class="health-stat-label svelte-1uha8ag">Projects</div></div> <div class="health-stat svelte-1uha8ag"><div class="health-stat-value svelte-1uha8ag">${escape_html(h.active_runs)}</div> <div class="health-stat-label svelte-1uha8ag">Active runs</div></div></div> <div class="fail-bar svelte-1uha8ag"><div${attr_class(`fail-bar-fill ${stringify(healthColor(h.fail_rate_7d))}`, "svelte-1uha8ag")}${attr_style(`width: ${stringify(Math.min(h.fail_rate_7d, 100))}%`)}></div></div> <div class="health-footer svelte-1uha8ag"><span>7d fail rate: ${escape_html(h.fail_rate_7d.toFixed(1))}%</span> <span>`);
        TimeAgo($$renderer2, { value: h.last_activity });
        $$renderer2.push(`<!----></span></div></div>`);
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.workspaces.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "◇",
        message: "No workspaces yet. Create your first one to get started.",
        actionHref: "/workspaces/new",
        actionLabel: "New workspace"
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="grid svelte-1uha8ag"><!--[-->`);
      const each_array_2 = ensure_array_like(data.workspaces);
      for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
        let ws = each_array_2[$$index_2];
        $$renderer2.push(`<a${attr("href", `/workspaces/${ws.id}`)} class="card svelte-1uha8ag"><h3 class="svelte-1uha8ag">${escape_html(ws.name)}</h3> <p class="svelte-1uha8ag">${escape_html(ws.description || "—")}</p> <div class="meta svelte-1uha8ag">created `);
        TimeAgo($$renderer2, { value: ws.created_at });
        $$renderer2.push(`<!----></div></a>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    if (data.recentRuns.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="section svelte-1uha8ag">`);
      SectionHeader($$renderer2, { title: "Recent runs" });
      $$renderer2.push(`<!----> <!--[-->`);
      const each_array_3 = ensure_array_like(data.recentRuns);
      for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
        let r = each_array_3[$$index_3];
        RunRow($$renderer2, { run: r });
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (comparisonData().length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="section svelte-1uha8ag"><button type="button" class="collapsible-toggle svelte-1uha8ag"><span${attr_class("toggle-arrow svelte-1uha8ag", void 0, { "open": comparisonOpen })}>▶</span> Workspace comparison</button> `);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.storage.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="section svelte-1uha8ag">`);
      SectionHeader($$renderer2, { title: "Storage overview" });
      $$renderer2.push(`<!----> <div class="storage-grid svelte-1uha8ag"><!--[-->`);
      const each_array_6 = ensure_array_like(data.storage);
      for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
        let s = each_array_6[$$index_6];
        const maxTotal = storageMax(data.storage);
        $$renderer2.push(`<div class="storage-card svelte-1uha8ag"><div class="storage-name svelte-1uha8ag">${escape_html(s.workspace_name)}</div> <div class="storage-row svelte-1uha8ag"><span class="storage-label svelte-1uha8ag">Projects</span> <div class="storage-bar-track svelte-1uha8ag"><div class="storage-bar-fill bar-projects svelte-1uha8ag"${attr_style(`width: ${stringify(s.projects / maxTotal * 100)}%`)}></div></div> <span class="storage-count svelte-1uha8ag">${escape_html(s.projects)}</span></div> <div class="storage-row svelte-1uha8ag"><span class="storage-label svelte-1uha8ag">Tasks</span> <div class="storage-bar-track svelte-1uha8ag"><div class="storage-bar-fill bar-tasks svelte-1uha8ag"${attr_style(`width: ${stringify(s.tasks / maxTotal * 100)}%`)}></div></div> <span class="storage-count svelte-1uha8ag">${escape_html(s.tasks)}</span></div> <div class="storage-row svelte-1uha8ag"><span class="storage-label svelte-1uha8ag">Runs</span> <div class="storage-bar-track svelte-1uha8ag"><div class="storage-bar-fill bar-runs svelte-1uha8ag"${attr_style(`width: ${stringify(s.runs / maxTotal * 100)}%`)}></div></div> <span class="storage-count svelte-1uha8ag">${escape_html(s.runs)}</span></div> <div class="storage-row svelte-1uha8ag"><span class="storage-label svelte-1uha8ag">Artifacts</span> <div class="storage-bar-track svelte-1uha8ag"><div class="storage-bar-fill bar-artifacts svelte-1uha8ag"${attr_style(`width: ${stringify(s.artifacts / maxTotal * 100)}%`)}></div></div> <span class="storage-count svelte-1uha8ag">${escape_html(s.artifacts)}</span></div></div>`);
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
