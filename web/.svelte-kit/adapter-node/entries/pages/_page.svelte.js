import { a as attr, f as attr_style, s as stringify, e as escape_html, c as ensure_array_like } from "../../chunks/renderer.js";
import { P as PageHeader } from "../../chunks/PageHeader.js";
import { E as EmptyState } from "../../chunks/EmptyState.js";
import { F as FlashMessage } from "../../chunks/FlashMessage.js";
import { a as statusColor, f as formatRelative } from "../../chunks/utils2.js";
function RunRow($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { run, showLocation = true } = $$props;
    $$renderer2.push(`<a${attr("href", `/runs/${run.id}`)} class="run-row svelte-1n4sxwt"><span class="dot"${attr_style(`background: ${stringify(statusColor(run.status))}`)}></span> <div class="run-info"><div class="run-task svelte-1n4sxwt">${escape_html(run.task_name)} <span class="dim svelte-1n4sxwt">(${escape_html(run.task_source)})</span></div> `);
    if (showLocation && run.workspace_name && run.project_name) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div class="run-loc svelte-1n4sxwt">${escape_html(run.workspace_name)} / ${escape_html(run.project_name)}</div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--></div> <span class="run-meta svelte-1n4sxwt">${escape_html(run.status)}${escape_html(run.exit_code !== null ? ` · exit ${run.exit_code}` : "")}</span> <span class="run-meta svelte-1n4sxwt">${escape_html(formatRelative(run.created_at))}</span></a>`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
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
      const each_array = ensure_array_like(data.workspaces);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let ws = each_array[$$index];
        $$renderer2.push(`<a${attr("href", `/workspaces/${ws.id}`)} class="card svelte-1uha8ag"><h3 class="svelte-1uha8ag">${escape_html(ws.name)}</h3> <p class="svelte-1uha8ag">${escape_html(ws.description || "—")}</p> <div class="meta svelte-1uha8ag">created ${escape_html(new Date(ws.created_at).toLocaleDateString())}</div></a>`);
      }
      $$renderer2.push(`<!--]--></div>`);
    }
    $$renderer2.push(`<!--]--> `);
    if (data.recentRuns.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<h2 class="svelte-1uha8ag">Recent runs</h2> <!--[-->`);
      const each_array_1 = ensure_array_like(data.recentRuns);
      for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
        let r = each_array_1[$$index_1];
        RunRow($$renderer2, { run: r });
      }
      $$renderer2.push(`<!--]-->`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
