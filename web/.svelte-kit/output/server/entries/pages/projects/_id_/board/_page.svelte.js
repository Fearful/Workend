import { c as ensure_array_like, b as attr_class, s as stringify, a as attr, e as escape_html, d as derived, f as attr_style } from "../../../../../chunks/renderer.js";
import "@sveltejs/kit/internal";
import "../../../../../chunks/exports.js";
import "../../../../../chunks/utils.js";
import "@sveltejs/kit/internal/server";
import "../../../../../chunks/root.js";
import "../../../../../chunks/state.svelte.js";
import { P as Panel } from "../../../../../chunks/Panel.js";
import { F as FlashMessage } from "../../../../../chunks/FlashMessage.js";
import { S as SectionHeader } from "../../../../../chunks/SectionHeader.js";
import { T as TimeAgo } from "../../../../../chunks/TimeAgo.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let dragIssueID = null;
    let dropColumnName = null;
    let columnsView = derived(() => {
      if (!data.boardResp.configured || !data.boardResp.board) return [];
      const cols = data.boardResp.board.columns.slice().sort((a, b) => a.position - b.position);
      const colNames = new Set(cols.map((c) => c.name));
      const buckets = {};
      for (const c of cols) buckets[c.name] = [];
      const unassigned = [];
      for (const issue of data.boardResp.board.issues) {
        if (issue.state === "closed") continue;
        const matched = issue.labels.find((l) => colNames.has(l));
        if (matched) buckets[matched].push(issue);
        else unassigned.push(issue);
      }
      const out = cols.map((c) => ({ name: c.name, issues: buckets[c.name] }));
      if (unassigned.length > 0) out.push({ name: "__unassigned__", issues: unassigned });
      return out;
    });
    {
      let actions = function($$renderer3) {
        if (data.boardResp.configured && data.boardResp.board) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<span class="sync-time svelte-ue3r68">`);
          if (data.boardResp.board.last_synced_at) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`synced `);
            TimeAgo($$renderer3, { value: data.boardResp.board.last_synced_at });
            $$renderer3.push(`<!---->`);
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`never synced`);
          }
          $$renderer3.push(`<!--]--></span> <form method="POST" action="?/sync" class="inline-form"><button type="submit" class="ghost">Sync</button></form>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      };
      SectionHeader($$renderer2, { title: "Issue board", actions });
    }
    $$renderer2.push(`<!----> `);
    if (data.boardError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->${escape_html(data.boardError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
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
    $$renderer2.push(`<!--]--> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.syncError) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "error",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Setup saved, but the initial sync failed: ${escape_html(form.syncError)}`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (form?.synced) {
      $$renderer2.push("<!--[0-->");
      FlashMessage($$renderer2, {
        type: "success",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->Board synced.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> `);
    if (!data.boardResp.configured) {
      $$renderer2.push("<!--[0-->");
      Panel($$renderer2, {
        title: "Choose your columns",
        children: ($$renderer3) => {
          $$renderer3.push(`<p style="color: var(--text-muted); font-size: 0.875rem; margin: 0 0 var(--space-2) 0;">Pick the upstream labels that should become board columns. Workflow-style
      labels like <em>todo</em>, <em>in progress</em>, <em>done</em> work best.
      You can change this later by re-setting up.</p> `);
          if ((data.boardResp.available_labels ?? []).length === 0) {
            $$renderer3.push("<!--[0-->");
            $$renderer3.push(`<p style="color: var(--text-dim); font-size: 0.875rem;">No upstream labels found. Connect this project's git provider in <a href="/settings">Settings</a>, then come back.</p>`);
          } else {
            $$renderer3.push("<!--[-1-->");
            $$renderer3.push(`<form method="POST" action="?/setup"><div class="label-grid svelte-ue3r68"><!--[-->`);
            const each_array = ensure_array_like(data.boardResp.available_labels ?? []);
            for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
              let l = each_array[$$index];
              $$renderer3.push(`<label class="label-pick svelte-ue3r68"><input type="checkbox" name="columns"${attr("value", l.name)} class="svelte-ue3r68"/> <span class="swatch svelte-ue3r68"${attr_style(`background: #${stringify(l.color || "6b7280")};`)}></span> <span>${escape_html(l.name)}</span></label>`);
            }
            $$renderer3.push(`<!--]--></div> <button type="submit">Save &amp; sync</button></form>`);
          }
          $$renderer3.push(`<!--]-->`);
        }
      });
    } else if (data.boardResp.board) {
      $$renderer2.push("<!--[1-->");
      $$renderer2.push(`<div class="kanban svelte-ue3r68"><!--[-->`);
      const each_array_1 = ensure_array_like(columnsView());
      for (let $$index_3 = 0, $$length = each_array_1.length; $$index_3 < $$length; $$index_3++) {
        let col = each_array_1[$$index_3];
        $$renderer2.push(`<div${attr_class(`col ${stringify(dropColumnName === col.name ? "drop-target" : "")}`, "svelte-ue3r68")} role="region"${attr("aria-label", col.name === "__unassigned__" ? "Unassigned column" : `Column ${col.name}`)}><div class="col-head svelte-ue3r68"><span>${escape_html(col.name === "__unassigned__" ? "Unassigned" : col.name)}</span> <span class="col-count svelte-ue3r68">${escape_html(col.issues.length)}</span></div> <!--[-->`);
        const each_array_2 = ensure_array_like(col.issues);
        for (let $$index_2 = 0, $$length2 = each_array_2.length; $$index_2 < $$length2; $$index_2++) {
          let i = each_array_2[$$index_2];
          $$renderer2.push(`<a${attr("href", `/issues/${i.id}`)}${attr_class(`card ${stringify(dragIssueID === i.id ? "dragging" : "")}`, "svelte-ue3r68")}${attr("draggable", col.name !== "__unassigned__")}><span class="card-num svelte-ue3r68">#${escape_html(i.provider_number)}</span> <p class="card-title svelte-ue3r68">${escape_html(i.title)}</p> <div class="card-tags svelte-ue3r68"><!--[-->`);
          const each_array_3 = ensure_array_like(i.labels);
          for (let $$index_1 = 0, $$length3 = each_array_3.length; $$index_1 < $$length3; $$index_1++) {
            let l = each_array_3[$$index_1];
            if (l !== col.name) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="tag svelte-ue3r68">${escape_html(l)}</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]-->`);
          }
          $$renderer2.push(`<!--]--></div></a>`);
        }
        $$renderer2.push(`<!--]--> `);
        if (col.issues.length === 0) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<p class="col-empty svelte-ue3r68">${escape_html(col.name === "__unassigned__" ? "" : "Drop issues here")}</p>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]--></div>`);
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
