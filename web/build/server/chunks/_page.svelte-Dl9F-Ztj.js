import { a7 as attr, a9 as attr_class, ae as stringify, ac as ensure_array_like, a8 as escape_html, a2 as derived } from './renderer-mjPKoiGx.js';
import './root-gJ1T4-40.js';
import './state.svelte-CeeZin_s.js';
import { f as formatRelative } from './utils2-B5RqTmai.js';
import { B as Breadcrumb } from './Breadcrumb-BbENYDvT.js';
import { P as PageHeader } from './PageHeader-CaLG0rbe.js';
import { F as FlashMessage } from './FlashMessage-6DSDTh4l.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    function providerLabel(kind) {
      switch (kind) {
        case "github":
          return "GitHub";
        case "gitlab":
          return "GitLab";
        case "gitea":
          return "Gitea";
        default:
          return kind;
      }
    }
    function pageURL(p) {
      const u = new URL(window.location.href);
      u.searchParams.set("page", String(p));
      return u.pathname + u.search;
    }
    let filter = "";
    let visibleRepos = derived(() => {
      if (!filter.trim()) return data.repos;
      const q = filter.toLowerCase();
      return data.repos.filter((r) => r.full_name.toLowerCase().includes(q) || r.description.toLowerCase().includes(q));
    });
    Breadcrumb($$renderer2, {
      segments: [
        { label: "workspaces", href: "/" },
        {
          label: data.workspace.name,
          href: `/workspaces/${data.workspace.id}`
        },
        { label: "new project" }
      ]
    });
    $$renderer2.push(`<!----> `);
    PageHeader($$renderer2, { title: "Add project" });
    $$renderer2.push(`<!----> <div class="layout svelte-1lviau7"><form method="POST" class="svelte-1lviau7"><div class="field"><label for="name">Name</label> <input id="name" name="name" type="text" required="" maxlength="100"${attr("value", form?.name || "")}/> <p class="hint svelte-1lviau7">Display name for this project. Must be unique within the workspace.</p></div> <div class="field"><label for="git_url">Git URL</label> <input id="git_url" name="git_url" type="url" required="" placeholder="https://github.com/user/repo.git"${attr("value", form?.git_url || "")}/> <p class="hint svelte-1lviau7">Public HTTPS URL, or click a repo from a connected provider on the right →</p></div> <div class="field"><label for="branch">Branch (optional)</label> <input id="branch" name="branch" type="text" placeholder="main"${attr("value", form?.branch || "")}/> <p class="hint svelte-1lviau7">Leave blank to use the repo's default branch.</p></div> `);
    if (form?.error) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(form.error)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <div class="actions svelte-1lviau7"><button type="submit">Add</button> <a${attr("href", `/workspaces/${data.workspace.id}`)}><button type="button" class="ghost">Cancel</button></a></div></form> <div><h2 class="svelte-1lviau7">Pick from a connected provider</h2> `);
    if (data.connectedProviders.length === 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="panel panel-empty svelte-1lviau7">No providers connected. <a href="/settings">Connect one in Settings →</a></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="picker-tabs svelte-1lviau7"><button${attr_class(`pill ${stringify(!data.activeProvider ? "active" : "")}`, "svelte-1lviau7")}>—</button> <!--[-->`);
      const each_array = ensure_array_like(data.connectedProviders);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let c = each_array[$$index];
        $$renderer2.push(`<button${attr_class(`pill ${stringify(data.activeProvider?.provider_id === c.provider_id ? "active" : "")}`, "svelte-1lviau7")}>${escape_html(providerLabel(c.provider))}: ${escape_html(c.instance_host)}</button>`);
      }
      $$renderer2.push(`<!--]--></div> `);
      if (!data.activeProvider) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="panel panel-empty svelte-1lviau7">Select a provider to browse repos.</div>`);
      } else if (data.reposError) {
        $$renderer2.push("<!--[1-->");
        $$renderer2.push(`<div class="panel panel-empty error svelte-1lviau7">${escape_html(data.reposError)}</div>`);
      } else if (data.repos.length === 0) {
        $$renderer2.push("<!--[2-->");
        $$renderer2.push(`<div class="panel panel-empty svelte-1lviau7">No repos returned for page ${escape_html(data.page)}.</div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<input type="text" placeholder="Filter this page…"${attr("value", filter)} class="filter-input svelte-1lviau7"/> <div class="panel panel-tight svelte-1lviau7">`);
        if (visibleRepos().length === 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<div class="panel-empty svelte-1lviau7">No repos match <code>${escape_html(filter)}</code>.</div>`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<!--[-->`);
          const each_array_1 = ensure_array_like(visibleRepos());
          for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
            let r = each_array_1[$$index_1];
            $$renderer2.push(`<button type="button" class="repo-row svelte-1lviau7"><div class="repo-name svelte-1lviau7">${escape_html(r.full_name)} `);
            if (r.private) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="lock svelte-1lviau7">🔒</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            if (r.description) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<div class="desc svelte-1lviau7">${escape_html(r.description)}</div>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--></div> <div class="repo-meta svelte-1lviau7">${escape_html(r.default_branch)}</div> <div class="repo-meta svelte-1lviau7">${escape_html(formatRelative(r.updated_at))}</div></button>`);
          }
          $$renderer2.push(`<!--]-->`);
        }
        $$renderer2.push(`<!--]--></div> <div class="pager svelte-1lviau7">`);
        if (data.page > 1) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<a${attr("href", pageURL(data.page - 1))} class="svelte-1lviau7">← prev</a>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--> <span class="pager-current svelte-1lviau7">page ${escape_html(data.page)}</span> `);
        if (data.repos.length >= 50) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<a${attr("href", pageURL(data.page + 1))} class="svelte-1lviau7">next →</a>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></div>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div></div>`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-Dl9F-Ztj.js.map
