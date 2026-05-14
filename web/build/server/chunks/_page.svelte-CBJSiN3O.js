import { a8 as escape_html, ac as ensure_array_like, a7 as attr } from './renderer-D0X3o35U.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';
import { P as Panel } from './Panel-B1HKSjmN.js';
import { B as Badge } from './Badge-imHJGUZA.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data, form } = $$props;
    let exportFrom = "";
    let exportTo = "";
    PageHeader($$renderer2, { title: "Admin" });
    $$renderer2.push(`<!----> <h2 class="svelte-1jef3w8">Users (${escape_html(data.users.length)})</h2> <div class="table-wrapper svelte-1jef3w8"><table class="svelte-1jef3w8"><thead><tr><th class="svelte-1jef3w8">Display name</th><th class="svelte-1jef3w8">Email</th><th class="svelte-1jef3w8">Created</th><th class="svelte-1jef3w8">Workspaces</th><th class="svelte-1jef3w8">Projects</th><th class="svelte-1jef3w8">Runs</th></tr></thead><tbody><!--[-->`);
    const each_array = ensure_array_like(data.users);
    for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
      let u = each_array[$$index];
      $$renderer2.push(`<tr class="svelte-1jef3w8"><td class="svelte-1jef3w8"><span class="name-cell svelte-1jef3w8">${escape_html(u.display_name)} `);
      if (u.is_admin) {
        $$renderer2.push("<!--[0-->");
        Badge($$renderer2, {
          variant: "accent",
          size: "sm",
          children: ($$renderer3) => {
            $$renderer3.push(`<!---->ADMIN`);
          }
        });
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></span></td><td class="mono svelte-1jef3w8">${escape_html(u.email)}</td><td class="muted svelte-1jef3w8">${escape_html(new Date(u.created_at).toLocaleDateString())}</td><td class="mono svelte-1jef3w8">${escape_html(u.workspace_count)}</td><td class="mono svelte-1jef3w8">${escape_html(u.project_count)}</td><td class="mono svelte-1jef3w8">${escape_html(u.run_count)}</td></tr>`);
    }
    $$renderer2.push(`<!--]--></tbody></table></div> <h2 class="svelte-1jef3w8">Audit log (last ${escape_html(data.audit.length)})</h2> <div class="table-wrapper svelte-1jef3w8"><table class="svelte-1jef3w8"><thead><tr><th class="svelte-1jef3w8">When</th><th class="svelte-1jef3w8">Actor</th><th class="svelte-1jef3w8">Action</th><th class="svelte-1jef3w8">Target</th><th class="svelte-1jef3w8">IP</th></tr></thead><tbody><!--[-->`);
    const each_array_1 = ensure_array_like(data.audit);
    for (let $$index_1 = 0, $$length = each_array_1.length; $$index_1 < $$length; $$index_1++) {
      let e = each_array_1[$$index_1];
      $$renderer2.push(`<tr class="svelte-1jef3w8"><td class="muted svelte-1jef3w8">${escape_html(new Date(e.occurred_at).toLocaleString())}</td><td class="mono svelte-1jef3w8">${escape_html(e.actor_email || "—")}</td><td class="mono svelte-1jef3w8">${escape_html(e.action)}</td><td class="muted svelte-1jef3w8">${escape_html(e.target_kind ? `${e.target_kind}:${(e.target_id ?? "").slice(0, 8)}` : "—")}</td><td class="muted svelte-1jef3w8">${escape_html(e.ip || "—")}</td></tr>`);
    }
    $$renderer2.push(`<!--]--></tbody></table></div> `);
    Panel($$renderer2, {
      title: "Audit log management",
      children: ($$renderer3) => {
        $$renderer3.push(`<p class="hint svelte-1jef3w8">Export audit log entries as CSV or configure automatic retention cleanup.</p> <div class="export-row svelte-1jef3w8"><div class="field svelte-1jef3w8"><label for="export-from" class="svelte-1jef3w8">From</label> <input id="export-from" type="date"${attr("value", exportFrom)}/></div> <div class="field svelte-1jef3w8"><label for="export-to" class="svelte-1jef3w8">To</label> <input id="export-to" type="date"${attr("value", exportTo)}/></div> <button type="button"${attr("disabled", !exportFrom, true)} class="ghost">${escape_html("Export CSV")}</button></div> <div class="retention-row svelte-1jef3w8"><form method="POST" action="?/setRetention" style="display: flex; align-items: flex-end; gap: var(--space-3); flex-wrap: wrap;"><div class="field svelte-1jef3w8"><label for="retention-days">Retention period (days)</label> <input id="retention-days" name="days" type="number" min="1" max="3650" placeholder="e.g. 90, 180, 365" style="width: 180px;"/></div> <button type="submit" class="ghost">Set retention</button></form></div> `);
        if (form?.retentionError) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-error svelte-1jef3w8">${escape_html(form.retentionError)}</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]--> `);
        if (form?.retentionSet) {
          $$renderer3.push("<!--[0-->");
          $$renderer3.push(`<p class="form-success svelte-1jef3w8">Retention set to ${escape_html(form.retentionDays)} days.</p>`);
        } else {
          $$renderer3.push("<!--[-1-->");
        }
        $$renderer3.push(`<!--]-->`);
      }
    });
    $$renderer2.push(`<!---->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-CBJSiN3O.js.map
