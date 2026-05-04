import { e as escape_html, a as attr, c as ensure_array_like, b as attr_class, s as stringify, f as attr_style } from "../../../../chunks/renderer.js";
import { s as shortSha, f as formatRelative } from "../../../../chunks/utils2.js";
import { B as Breadcrumb } from "../../../../chunks/Breadcrumb.js";
import { S as StatusDot } from "../../../../chunks/StatusDot.js";
/* empty css                                                          */
import { F as FlashMessage } from "../../../../chunks/FlashMessage.js";
import { B as Badge } from "../../../../chunks/Badge.js";
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
      if (action.startsWith("workspace.")) return "◯";
      if (action.startsWith("user.")) return "☉";
      return "·";
    }
    function eventHref(e) {
      if (e.target_kind === "run" && e.target_id) return `/runs/${e.target_id}`;
      if (e.target_kind === "project" && e.target_id) return `/projects/${e.target_id}`;
      return null;
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
    $$renderer2.push(`<!--]--></div></div> <h2 class="svelte-3osbwr">Projects</h2> `);
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
      $$renderer2.push(`<div class="empty-bordered svelte-3osbwr"><p>No projects yet. Add one by git URL.</p></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(data.projects);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let p = each_array[$$index];
        $$renderer2.push(`<a${attr("href", `/projects/${p.id}`)} class="project-row svelte-3osbwr">`);
        StatusDot($$renderer2, { status: p.status });
        $$renderer2.push(`<!----> <div><div class="project-name svelte-3osbwr">${escape_html(p.name)}</div> <div class="project-meta svelte-3osbwr">${escape_html(p.git_url)}</div></div> <div class="project-meta svelte-3osbwr">${escape_html(shortSha(p.last_commit_sha))}</div> <div class="project-status svelte-3osbwr">${escape_html(p.status)}</div></a>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--> <h2 class="svelte-3osbwr">Members</h2> <div class="list-card svelte-3osbwr"><!--[-->`);
    const each_array_1 = ensure_array_like(data.members);
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let m = each_array_1[$$index_1];
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
    $$renderer2.push(`<!--]--> `);
    if (data.recentIssues.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-3osbwr">Recent Issues</h2> <div class="list-card svelte-3osbwr"><!--[-->`);
      const each_array_2 = ensure_array_like(data.recentIssues);
      for (let $$index_3 = 0, $$length = each_array_2.length; $$index_3 < $$length; $$index_3++) {
        let issue = each_array_2[$$index_3];
        $$renderer2.push(`<a${attr("href", issue.html_url)} target="_blank" rel="noopener noreferrer" class="issue-row svelte-3osbwr"><span class="dot svelte-3osbwr"${attr_style(`background: ${stringify(issue.state === "open" ? "var(--success)" : "var(--text-dim)")}`)}></span> <div class="svelte-3osbwr"><span class="issue-num svelte-3osbwr">#${escape_html(issue.provider_number)}</span> <span>${escape_html(issue.title)}</span> `);
        if (issue.labels.length > 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<span class="issue-labels svelte-3osbwr"><!--[-->`);
          const each_array_3 = ensure_array_like(issue.labels);
          for (let $$index_2 = 0, $$length2 = each_array_3.length; $$index_2 < $$length2; $$index_2++) {
            let label = each_array_3[$$index_2];
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
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.activity.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-3osbwr">Activity</h2> <div class="list-card svelte-3osbwr"><!--[-->`);
      const each_array_4 = ensure_array_like(data.activity);
      for (let $$index_4 = 0, $$length = each_array_4.length; $$index_4 < $$length; $$index_4++) {
        let e = each_array_4[$$index_4];
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
      $$renderer2.push(`<!--]--></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
