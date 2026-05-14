import { a8 as escape_html, a7 as attr, ac as ensure_array_like, a9 as attr_class, a2 as derived } from './renderer-D0X3o35U.js';
import './root-BPP4VMgG.js';
import './state.svelte-DaEZ9E9Z.js';
import { s as shortSha } from './utils2-BUPlP7zG.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { p as page } from './index2-DKElt7m-.js';
import { S as StatusPill } from './StatusPill-j6_a2per.js';
import { T as TimeAgo } from './TimeAgo-DaxliHBa.js';
import { T as Tooltip } from './Tooltip-Bg85MZ7d.js';
import { M as Modal } from './Modal-BnjRYgDm.js';
import './client-Bp0-LT_C.js';
import './index-rPPP9l7g.js';
import './index-server-BFLhAcPs.js';

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
        label: "Pipelines",
        path: `/projects/${projectID}/pipelines`,
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
      return shortSha(sha, 7);
    }
    let deleteModal = false;
    let deletePending = false;
    let syncPending = false;
    let actionsMenuOpen = false;
    $$renderer2.push(`<div class="project-shell svelte-j0mxji">`);
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
    StatusPill($$renderer2, { status: data.project.status });
    $$renderer2.push(`<!----> <span class="name svelte-j0mxji">${escape_html(data.project.name)}</span></h1> <div class="actions-bar svelte-j0mxji"><button type="button"${attr("disabled", syncPending, true)} class="desktop-only svelte-j0mxji">${escape_html("Sync")}</button> <div class="more-menu-wrap svelte-j0mxji"><button type="button" class="more-btn svelte-j0mxji" aria-haspopup="menu"${attr("aria-expanded", actionsMenuOpen)} title="Project actions">⋯</button> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div></div></div> <div class="ribbon svelte-j0mxji"><span class="ribbon-item svelte-j0mxji"><span class="label svelte-j0mxji">Branch</span> `);
    Tooltip($$renderer2, {
      text: data.project.default_branch || "No default branch detected",
      children: ($$renderer3) => {
        $$renderer3.push(`<span class="val mono svelte-j0mxji">${escape_html(data.project.default_branch || "—")}</span>`);
      }
    });
    $$renderer2.push(`<!----></span> <span class="ribbon-sep svelte-j0mxji">·</span> `);
    if (data.project.last_commit_sha) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="ribbon-item svelte-j0mxji"><span class="label svelte-j0mxji">Commit</span> `);
      {
        let content = function($$renderer3) {
          $$renderer3.push(`<div class="commit-tip svelte-j0mxji"><div class="mono svelte-j0mxji">${escape_html(data.project.last_commit_sha)}</div> `);
          if (data.project.last_commit_author) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div class="author svelte-j0mxji">${escape_html(data.project.last_commit_author)}</div>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--> `);
          if (data.project.last_commit_message) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<div class="msg svelte-j0mxji">${escape_html(data.project.last_commit_message)}</div>`);
          } else {
            $$renderer3.push("<!--[-1-->");
          }
          $$renderer3.push(`<!--]--></div>`);
        };
        Tooltip($$renderer2, {
          placement: "bottom",
          content,
          children: ($$renderer3) => {
            $$renderer3.push(`<span class="val mono svelte-j0mxji">${escape_html(shortSha$1(data.project.last_commit_sha))}</span>`);
          }
        });
      }
      $$renderer2.push(`<!----></span> <span class="ribbon-sep svelte-j0mxji">·</span>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (data.project.last_synced_at) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<span class="ribbon-item svelte-j0mxji"><span class="label svelte-j0mxji">Synced</span> <span class="val svelte-j0mxji">`);
      TimeAgo($$renderer2, { value: data.project.last_synced_at });
      $$renderer2.push(`<!----></span></span>`);
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
    $$renderer2.push(`<!----></div> `);
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
//# sourceMappingURL=_layout.svelte-CaDVYT2z.js.map
