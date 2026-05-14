import { e as escape_html, a as attr, c as ensure_array_like, f as attr_style, s as stringify, b as attr_class } from "../../../../chunks/renderer.js";
import { b as formatRelative, s as shortSha } from "../../../../chunks/utils2.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
import { S as StatusPill } from "../../../../chunks/StatusPill.js";
import { E as EmptyState } from "../../../../chunks/EmptyState.js";
import { F as FlashMessage } from "../../../../chunks/FlashMessage.js";
import { B as Badge } from "../../../../chunks/Badge.js";
import { S as SectionHeader } from "../../../../chunks/SectionHeader.js";
import { T as TimeAgo } from "../../../../chunks/TimeAgo.js";
import { T as Tooltip } from "../../../../chunks/Tooltip.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    function actionLabel(action) {
      switch (action) {
        case "workspace.create":
          return "created the workspace";
        case "project.create":
          return "added a project";
        case "project.delete":
          return "deleted a project";
        case "project.sync":
          return "synced a project";
        case "run.start":
          return "started a run";
        case "run.cancel":
          return "cancelled a run";
        case "run.complete":
          return "completed a run";
        case "user.login":
          return "signed in";
        case "user.signup":
          return "signed up";
        default:
          return action.replace(/\./g, " ");
      }
    }
    function actionIcon(action) {
      if (action.startsWith("run.")) return "▶";
      if (action.startsWith("project.")) return "◇";
      if (action.startsWith("workspace.")) return "○";
      if (action.startsWith("user.")) return "☉";
      return "·";
    }
    function eventHref(e) {
      if (e.target_kind === "run" && e.target_id) return `/runs/${e.target_id}`;
      if (e.target_kind === "project" && e.target_id) return `/projects/${e.target_id}`;
      return null;
    }
    function pipelineStatusVariant(status) {
      switch (status) {
        case "passed":
        case "succeeded":
          return "success";
        case "running":
        case "queued":
          return "warning";
        case "failed":
          return "danger";
        default:
          return "muted";
      }
    }
    function sandboxStatusVariant(status) {
      switch (status) {
        case "running":
          return "success";
        case "provisioning":
          return "warning";
        case "stopped":
        case "expired":
          return "muted";
        case "error":
          return "danger";
        default:
          return "info";
      }
    }
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        { label: data.workspace.name }
      ]
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-3osbwr"><div><h1 class="svelte-3osbwr">${escape_html(data.workspace.name)} <span class="role-tag svelte-3osbwr">${escape_html(data.workspace.my_role)}</span></h1> <p class="desc svelte-3osbwr">${escape_html(data.workspace.description || "No description")}</p></div> <div class="actions svelte-3osbwr"><a${attr("href", `/workspaces/${data.workspace.id}/projects/new`)}><button>Add project</button></a> `);
    if (data.workspace.my_role === "owner") {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<form method="POST" action="?/delete" class="inline-form"><button type="submit" class="ghost">Delete workspace</button></form>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div> <nav class="ws-tabs svelte-3osbwr" aria-label="Workspace sections"><a class="ws-tab active svelte-3osbwr"${attr("href", `/workspaces/${data.workspace.id}`)}>Overview</a> <a class="ws-tab svelte-3osbwr"${attr("href", `/workspaces/${data.workspace.id}/dashboard`)}>Dashboard</a> <a class="ws-tab svelte-3osbwr"${attr("href", `/workspaces/${data.workspace.id}/activity`)}>Activity</a> <a class="ws-tab svelte-3osbwr"${attr("href", `/workspaces/${data.workspace.id}/secrets`)}>Secrets</a> <a class="ws-tab svelte-3osbwr"${attr("href", `/workspaces/${data.workspace.id}/roles`)}>Roles</a></nav> <div class="widget-grid svelte-3osbwr"><div class="widget svelte-3osbwr"><div class="widget-head svelte-3osbwr"><h3 class="widget-title svelte-3osbwr">Projects</h3> `);
    if (data.projectGrid.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="widget-link svelte-3osbwr" style="color: var(--text-dim); font-size: var(--fs-xs);">${escape_html(data.projectGrid.length)} total</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.projectGrid.length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widget-empty svelte-3osbwr">No projects yet</div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="proj-card-list svelte-3osbwr"><!--[-->`);
      const each_array = ensure_array_like(data.projectGrid.slice(0, 6));
      for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
        let p = each_array[$$index_1];
        $$renderer2.push(`<a${attr("href", `/projects/${p.project_id}`)} class="proj-card svelte-3osbwr"><div class="proj-color-bar svelte-3osbwr"${attr_style(`background: ${stringify(p.status_color || "var(--text-dim)")};`)}></div> <div><div class="proj-card-name svelte-3osbwr">${escape_html(p.project_name)}</div> <div class="proj-card-sub svelte-3osbwr">${escape_html(p.task_count)} tasks</div></div> <div class="proj-card-sparkline svelte-3osbwr"><!--[-->`);
        const each_array_1 = ensure_array_like(p.recent_runs.slice(0, 8));
        for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
          let run = each_array_1[$$index];
          $$renderer2.push(`<span class="spark-dot svelte-3osbwr"${attr_style(`background: ${stringify(run.status === "succeeded" || run.status === "passed" ? "var(--status-success-fg)" : run.status === "failed" ? "var(--status-danger-fg)" : "var(--status-warning-fg)")};`)}></span>`);
        }
        $$renderer2.push(`<!--]--></div></a>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--></div> <div class="widget svelte-3osbwr"><div class="widget-head svelte-3osbwr"><h3 class="widget-title svelte-3osbwr">Team</h3> `);
    if (data.teamPresence.length > 0) {
      $$renderer2.push("<!--[0-->");
      const activeCount = data.teamPresence.filter((m) => m.is_active).length;
      $$renderer2.push(`<span style="color: var(--text-dim); font-size: var(--fs-xs);">${escape_html(activeCount)} active</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.teamPresence.length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widget-empty svelte-3osbwr">No team members</div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="team-list svelte-3osbwr"><!--[-->`);
      const each_array_2 = ensure_array_like(data.teamPresence.slice(0, 8));
      for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
        let m = each_array_2[$$index_2];
        $$renderer2.push(`<div class="team-row svelte-3osbwr"><span${attr_class("presence-dot svelte-3osbwr", void 0, { "active": m.is_active, "inactive": !m.is_active })}></span> <span class="team-name svelte-3osbwr">${escape_html(m.display_name || m.email)} `);
        if (m.active_runs > 0) {
          $$renderer2.push("<!--[0-->");
          Badge($$renderer2, {
            variant: "info",
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->${escape_html(m.active_runs)} running`);
            }
          });
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></span> <span class="team-meta svelte-3osbwr">`);
        if (m.last_activity) {
          $$renderer2.push("<!--[0-->");
          TimeAgo($$renderer2, { value: m.last_activity });
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`--`);
        }
        $$renderer2.push(`<!--]--></span></div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.pipelineBoard) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widget svelte-3osbwr"><div class="widget-head svelte-3osbwr"><h3 class="widget-title svelte-3osbwr">Pipelines</h3></div> <div class="pipe-stats svelte-3osbwr"><div class="pipe-stat svelte-3osbwr"><div class="pipe-stat-num svelte-3osbwr" style="color: var(--text-dim);">${escape_html(data.pipelineBoard.queued)}</div> <div class="pipe-stat-label svelte-3osbwr">Queued</div></div> <div class="pipe-stat svelte-3osbwr"><div class="pipe-stat-num svelte-3osbwr" style="color: var(--status-info-fg);">${escape_html(data.pipelineBoard.running)}</div> <div class="pipe-stat-label svelte-3osbwr">Running</div></div> <div class="pipe-stat svelte-3osbwr"><div class="pipe-stat-num svelte-3osbwr" style="color: var(--status-success-fg);">${escape_html(data.pipelineBoard.passed)}</div> <div class="pipe-stat-label svelte-3osbwr">Passed</div></div> <div class="pipe-stat svelte-3osbwr"><div class="pipe-stat-num svelte-3osbwr" style="color: var(--status-danger-fg);">${escape_html(data.pipelineBoard.failed)}</div> <div class="pipe-stat-label svelte-3osbwr">Failed</div></div></div> `);
      if (data.pipelineBoard.runs.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="pipe-run-list svelte-3osbwr"><!--[-->`);
        const each_array_3 = ensure_array_like(data.pipelineBoard.runs.slice(0, 5));
        for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
          let run = each_array_3[$$index_3];
          $$renderer2.push(`<div class="pipe-run svelte-3osbwr">`);
          Badge($$renderer2, {
            variant: pipelineStatusVariant(run.status),
            size: "sm",
            children: ($$renderer3) => {
              $$renderer3.push(`<!---->${escape_html(run.status)}`);
            }
          });
          $$renderer2.push(`<!----> <span class="pipe-run-name svelte-3osbwr">${escape_html(run.pipeline_name)} / ${escape_html(run.project_name)}</span> <span class="pipe-run-time svelte-3osbwr">`);
          TimeAgo($$renderer2, { value: run.started_at });
          $$renderer2.push(`<!----></span></div>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<div class="widget-empty svelte-3osbwr">No recent runs</div>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.secretsSummary) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widget svelte-3osbwr"><div class="widget-head svelte-3osbwr"><h3 class="widget-title svelte-3osbwr">Secrets</h3> <a${attr("href", `/workspaces/${data.workspace.id}/secrets`)} class="widget-link svelte-3osbwr">Manage</a></div> <div class="stat-grid svelte-3osbwr"><div class="stat-cell svelte-3osbwr"><div class="stat-num svelte-3osbwr">${escape_html(data.secretsSummary.total)}</div> <div class="stat-label svelte-3osbwr">Total</div></div> <div class="stat-cell svelte-3osbwr"><div class="stat-num svelte-3osbwr">${escape_html(data.secretsSummary.creator_count)}</div> <div class="stat-label svelte-3osbwr">Contributors</div></div> <div class="stat-cell svelte-3osbwr"><div class="stat-label svelte-3osbwr" style="margin-top: 0;">Newest</div> <div style="font-size: var(--fs-sm);">`);
      if (data.secretsSummary.newest_updated) {
        $$renderer2.push("<!--[0-->");
        TimeAgo($$renderer2, { value: data.secretsSummary.newest_updated });
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`--`);
      }
      $$renderer2.push(`<!--]--></div></div> <div class="stat-cell svelte-3osbwr"><div class="stat-label svelte-3osbwr" style="margin-top: 0;">Oldest</div> <div style="font-size: var(--fs-sm);">`);
      if (data.secretsSummary.oldest_updated) {
        $$renderer2.push("<!--[0-->");
        TimeAgo($$renderer2, { value: data.secretsSummary.oldest_updated });
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`--`);
      }
      $$renderer2.push(`<!--]--></div></div></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.depOverview) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widget svelte-3osbwr"><div class="widget-head svelte-3osbwr"><h3 class="widget-title svelte-3osbwr">Dependencies</h3></div> <div class="stat-grid svelte-3osbwr"><div class="stat-cell svelte-3osbwr"><div class="stat-num svelte-3osbwr">${escape_html(data.depOverview.total_projects)}</div> <div class="stat-label svelte-3osbwr">Projects</div></div> <div class="stat-cell svelte-3osbwr"><div class="stat-num svelte-3osbwr">${escape_html(data.depOverview.total_deps)}</div> <div class="stat-label svelte-3osbwr">Edges</div></div> <div class="stat-cell svelte-3osbwr"><div class="stat-num svelte-3osbwr"${attr_style(`color: ${stringify(data.depOverview.unresolved_deps > 0 ? "var(--status-danger-fg)" : "var(--text)")};`)}>${escape_html(data.depOverview.unresolved_deps)}</div> <div class="stat-label svelte-3osbwr">Unresolved</div></div> <div class="stat-cell svelte-3osbwr"></div></div> `);
      if (data.depOverview.top_connected.length > 0) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="dep-list svelte-3osbwr"><!--[-->`);
        const each_array_4 = ensure_array_like(data.depOverview.top_connected.slice(0, 4));
        for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
          let dep = each_array_4[$$index_4];
          $$renderer2.push(`<div class="dep-row svelte-3osbwr"><a${attr("href", `/projects/${dep.project_id}`)} style="color: var(--text); font-size: var(--fs-sm);">${escape_html(dep.project_name)}</a> <span class="dep-edges svelte-3osbwr">${escape_html(dep.edge_count)} edges</span></div>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.sandboxItems.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="widget svelte-3osbwr"><div class="widget-head svelte-3osbwr"><h3 class="widget-title svelte-3osbwr">Sandboxes</h3> <span style="color: var(--text-dim); font-size: var(--fs-xs);">${escape_html(data.sandboxItems.length)} active</span></div> <div class="sandbox-list svelte-3osbwr"><!--[-->`);
      const each_array_5 = ensure_array_like(data.sandboxItems.slice(0, 5));
      for (let $$index_5 = 0, $$length = each_array_5.length; $$index_5 < $$length; $$index_5++) {
        let item = each_array_5[$$index_5];
        $$renderer2.push(`<div class="sandbox-row svelte-3osbwr">`);
        Badge($$renderer2, {
          variant: sandboxStatusVariant(item.status),
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(item.status)}`);
          }
        });
        $$renderer2.push(`<!----> <div><div class="sandbox-name svelte-3osbwr">${escape_html(item.project_name)}</div> <div class="sandbox-branch svelte-3osbwr">${escape_html(item.branch)}</div></div> <span class="sandbox-ttl svelte-3osbwr">`);
        if (item.minutes_left != null) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`${escape_html(item.minutes_left)}m left`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`${escape_html(item.type)}`);
        }
        $$renderer2.push(`<!--]--></span></div>`);
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> <div class="section svelte-3osbwr">`);
    SectionHeader($$renderer2, { title: "Projects" });
    $$renderer2.push(`<!----> `);
    if (data.projectsError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.projectsError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.projects.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        message: "No projects yet. Add one by git URL.",
        actionHref: `/workspaces/${data.workspace.id}/projects/new`,
        actionLabel: "Add project"
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array_6 = ensure_array_like(data.projects);
      for (let $$index_6 = 0, $$length = each_array_6.length; $$index_6 < $$length; $$index_6++) {
        let p = each_array_6[$$index_6];
        $$renderer2.push(`<a${attr("href", `/projects/${p.id}`)} class="project-row svelte-3osbwr">`);
        StatusPill($$renderer2, { status: p.status, size: "sm" });
        $$renderer2.push(`<!----> <div><div class="project-name svelte-3osbwr">${escape_html(p.name)}</div> <div class="project-sub svelte-3osbwr"><span class="mono svelte-3osbwr">${escape_html(p.git_url.replace(/^https?:\/\//, "").replace(/\.git$/, ""))}</span> `);
        if (p.last_commit_sha) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="sep svelte-3osbwr">·</span> `);
          Tooltip($$renderer2, {
            text: p.last_commit_sha,
            children: ($$renderer3) => {
              $$renderer3.push(`<span class="mono svelte-3osbwr">${escape_html(shortSha(p.last_commit_sha))}</span>`);
            }
          });
          $$renderer2.push(`<!---->`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> `);
        if (p.last_synced_at) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="sep svelte-3osbwr">·</span> <span>synced `);
          TimeAgo($$renderer2, { value: p.last_synced_at });
          $$renderer2.push(`<!----></span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></div></div></a>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div> <div class="section svelte-3osbwr">`);
    SectionHeader($$renderer2, { title: "Members" });
    $$renderer2.push(`<!----> <div class="list-card svelte-3osbwr"><!--[-->`);
    const each_array_7 = ensure_array_like(data.members);
    for (let $$index_7 = 0, $$length = each_array_7.length; $$index_7 < $$length; $$index_7++) {
      let m = each_array_7[$$index_7];
      $$renderer2.push(`<div class="member-row svelte-3osbwr"><span class="member-name svelte-3osbwr">${escape_html(m.display_name)}</span> <span class="member-email svelte-3osbwr">${escape_html(m.email)}</span> <span${attr_class(`role-label ${stringify(m.role === "owner" ? "role-owner" : "role-member")}`, "svelte-3osbwr")}>${escape_html(m.role)}</span> `);
      if (data.workspace.my_role === "owner" || m.user_id === data.user?.id) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<form method="POST" action="?/removeMember" class="inline-form"><input type="hidden" name="user_id"${attr("value", m.user_id)}/> <button type="submit" class="ghost">${escape_html(m.user_id === data.user?.id ? "Leave" : "Remove")}</button></form>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<span></span>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.workspace.my_role === "owner") {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="add-member-form svelte-3osbwr"><form method="POST" action="?/addMember" class="add-member-grid svelte-3osbwr"><div class="field field-tight svelte-3osbwr"><label for="email">Add member by email</label> <input id="email" name="email" type="email" required="" placeholder="alice@example.com"${attr("value", form?.email || "")}/></div> <div class="field field-tight svelte-3osbwr"><label for="role">Role</label> <select id="role" name="role" class="role-select svelte-3osbwr">`);
      $$renderer2.option({ value: "member" }, ($$renderer3) => {
        $$renderer3.push(`member`);
      });
      $$renderer2.option({ value: "owner" }, ($$renderer3) => {
        $$renderer3.push(`owner`);
      });
      $$renderer2.push(`</select></div> <button type="submit">Add</button></form> `);
      if (form?.memberError) {
        $$renderer2.push("<!--[0-->");
        FlashMessage($$renderer2, {
          type: "error",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->${escape_html(form.memberError)}`);
          }
        });
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    if (data.recentIssues.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="section svelte-3osbwr">`);
      SectionHeader($$renderer2, { title: "Recent issues" });
      $$renderer2.push(`<!----> <div class="list-card svelte-3osbwr"><!--[-->`);
      const each_array_8 = ensure_array_like(data.recentIssues);
      for (let $$index_9 = 0, $$length = each_array_8.length; $$index_9 < $$length; $$index_9++) {
        let issue = each_array_8[$$index_9];
        $$renderer2.push(`<a${attr("href", issue.html_url)} target="_blank" rel="noopener noreferrer" class="issue-row svelte-3osbwr"><span class="dot svelte-3osbwr"${attr_style(`background: ${stringify(issue.state === "open" ? "var(--success)" : "var(--text-dim)")}`)}></span> <div class="svelte-3osbwr"><span class="issue-num svelte-3osbwr">#${escape_html(issue.provider_number)}</span> <span>${escape_html(issue.title)}</span> `);
        if (issue.labels.length > 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="issue-labels svelte-3osbwr"><!--[-->`);
          const each_array_9 = ensure_array_like(issue.labels);
          for (let $$index_8 = 0, $$length2 = each_array_9.length; $$index_8 < $$length2; $$index_8++) {
            let label = each_array_9[$$index_8];
            Badge($$renderer2, {
              variant: "info",
              size: "sm",
              children: ($$renderer3) => {
                $$renderer3.push(`<!---->${escape_html(label)}`);
              }
            });
          }
          $$renderer2.push(`<!--]--></span>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></div> <span class="issue-project svelte-3osbwr">${escape_html(issue.project_name)}</span> <span class="issue-time svelte-3osbwr">${escape_html(issue.upstream_updated_at ? formatRelative(issue.upstream_updated_at) : "")}</span></a>`);
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.activity.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="section svelte-3osbwr">`);
      {
        let actions = function($$renderer3) {
          $$renderer3.push(`<a${attr("href", `/workspaces/${data.workspace.id}/activity`)} class="widget-link svelte-3osbwr">View all</a>`);
        };
        SectionHeader($$renderer2, { title: "Activity", actions });
      }
      $$renderer2.push(`<!----> <div class="list-card svelte-3osbwr"><!--[-->`);
      const each_array_10 = ensure_array_like(data.activity);
      for (let $$index_10 = 0, $$length = each_array_10.length; $$index_10 < $$length; $$index_10++) {
        let e = each_array_10[$$index_10];
        const href = eventHref(e);
        $$renderer2.push(`<div class="activity-row svelte-3osbwr"><span class="activity-icon svelte-3osbwr">${escape_html(actionIcon(e.action))}</span> <span><strong class="activity-actor svelte-3osbwr">${escape_html(e.actor_name || "someone")}</strong> <span class="activity-action svelte-3osbwr">${escape_html(actionLabel(e.action))}</span> `);
        if (href) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<a${attr("href", href)} class="activity-target svelte-3osbwr">${escape_html(e.target_id.slice(0, 8))}</a>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></span> <span class="activity-time svelte-3osbwr">${escape_html(formatRelative(e.occurred_at))}</span></div>`);
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
