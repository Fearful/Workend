import { a8 as escape_html, a7 as attr, ac as ensure_array_like, a9 as attr_class, a2 as derived } from './renderer-mjPKoiGx.js';
import './root-gJ1T4-40.js';
import './state.svelte-CeeZin_s.js';
import { s as shortSha } from './utils2-B5RqTmai.js';
import { B as Breadcrumb } from './Breadcrumb-BbENYDvT.js';
import { p as page } from './index2-E0jCU3nx.js';
import { S as StatusDot } from './StatusDot-Cf5xUSPv.js';
import { M as Modal } from './Modal-BJw4bo_R.js';
import './client-CqnwEXXv.js';
import './index-B8uUmbHp.js';

function ProjectTabs($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { projectID } = $$props;
    let tabs = derived(() => [
      {
        label: "Overview",
        path: `/projects/${projectID}`,
        match: "exact"
      },
      {
        label: "Runs",
        path: `/projects/${projectID}/runs`,
        match: "prefix"
      },
      {
        label: "Board",
        path: `/projects/${projectID}/board`,
        match: "prefix"
      },
      {
        label: "Branches",
        path: `/projects/${projectID}/branches`,
        match: "prefix"
      },
      {
        label: "Trends",
        path: `/projects/${projectID}/trends`,
        match: "prefix"
      },
      {
        label: "Images",
        path: `/projects/${projectID}/images`,
        match: "prefix"
      },
      {
        label: "Schedules",
        path: `/projects/${projectID}/schedules`,
        match: "prefix"
      }
    ]);
    function isActive(t) {
      const p = page.url.pathname;
      return t.match === "exact" ? p === t.path : p.startsWith(t.path);
    }
    $$renderer2.push(`<nav class="project-tabs svelte-1p8fbsl" aria-label="Project sections"><div class="tabs-inner svelte-1p8fbsl"><!--[-->`);
    const each_array = ensure_array_like(tabs());
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let t = each_array[$$index];
      $$renderer2.push(`<a${attr("href", t.path)}${attr_class("tab svelte-1p8fbsl", void 0, { "active": isActive(t) })}${attr("aria-current", isActive(t) ? "page" : void 0)}>${escape_html(t.label)}</a>`);
    }
    $$renderer2.push(`<!--]--></div></nav>`);
  });
}
function _layout($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, children } = $$props;
    function shortSha$1(sha) {
      return shortSha(sha, 12);
    }
    let deleteModal = false;
    let deletePending = false;
    let syncPending = false;
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        ...data.workspace ? [
          {
            label: data.workspace.name,
            href: `/workspaces/${data.workspace.id}`
          }
        ] : [],
        { label: data.project.name }
      ]
    });
    $$renderer2.push(`<!----> <div class="header-row svelte-j0mxji"><h1 class="svelte-j0mxji">`);
    StatusDot($$renderer2, { status: data.project.status, size: 10 });
    $$renderer2.push(`<!----> ${escape_html(data.project.name)}</h1> <div class="actions-bar svelte-j0mxji"><button type="button"${attr("disabled", syncPending, true)}>${escape_html("Sync")}</button> <button type="button" class="ghost">Delete</button></div></div> <div class="status-strip svelte-j0mxji"><span><strong class="svelte-j0mxji">${escape_html(data.project.status)}</strong></span> <span class="sep svelte-j0mxji">·</span> <span>branch <strong class="svelte-j0mxji">${escape_html(data.project.default_branch || "—")}</strong></span> `);
    if (data.project.last_commit_sha) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="sep svelte-j0mxji">·</span> <span>commit <strong class="svelte-j0mxji">${escape_html(shortSha$1(data.project.last_commit_sha))}</strong></span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.project.last_synced_at) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="sep svelte-j0mxji">·</span> <span>synced ${escape_html(new Date(data.project.last_synced_at).toLocaleString())}</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> `);
    ProjectTabs($$renderer2, { projectID: data.project.id });
    $$renderer2.push(`<!----> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    children($$renderer2);
    $$renderer2.push(`<!----> `);
    Modal($$renderer2, {
      open: deleteModal,
      title: "Delete project?",
      width: 420,
      onClose: () => deleteModal = false,
      children: ($$renderer3) => {
        $$renderer3.push(`<p style="margin: 0 0 var(--space-3) 0;">This will permanently delete <strong>${escape_html(data.project.name)}</strong>, its tasks, run history, and any configured schedules.
    The remote git repository is unaffected.</p> <div class="delete-actions svelte-j0mxji"><button type="button" class="ghost"${attr("disabled", deletePending, true)}>Cancel</button> <button type="button" class="danger"${attr("disabled", deletePending, true)}>${escape_html("Delete project")}</button></div>`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}

export { _layout as default };
//# sourceMappingURL=_layout.svelte-CUy9vLBM.js.map
