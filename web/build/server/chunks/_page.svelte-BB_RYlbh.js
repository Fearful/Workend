import { ac as ensure_array_like, a7 as attr, a9 as attr_class, ad as stringify, a8 as escape_html } from './renderer-D0X3o35U.js';
import { b as formatRelative } from './utils2-BUPlP7zG.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';
import { E as EmptyState } from './EmptyState-5A_wG7lI.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    PageHeader($$renderer2, { title: "Mentions" });
    $$renderer2.push(`<!----> <div class="toolbar svelte-9oz8a8">`);
    if (data.unreadOnly) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<a href="/mentions">All mentions</a>`);
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<a href="/mentions?unread=true">Unread only</a>`);
    }
    $$renderer2.push(`<!--]--> <span class="toolbar-spacer svelte-9oz8a8"></span> <form method="POST" action="?/markRead" class="inline-form"><button type="submit" class="ghost">Mark all read</button></form></div> `);
    if (data.mentions.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, { icon: "@", message: "No mentions yet." });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(data.mentions);
      for (let $$index = 0, $$length = each_array.length; $$index < $$length; $$index++) {
        let m = each_array[$$index];
        $$renderer2.push(`<a${attr("href", `/runs/${m.run_id}`)}${attr_class(`row ${stringify(m.read_at === null ? "unread" : "")}`, "svelte-9oz8a8")}><div class="head svelte-9oz8a8"><span class="actor svelte-9oz8a8">${escape_html(m.actor_name || "someone")}</span> <span>mentioned you in</span> <span class="task-name svelte-9oz8a8">${escape_html(m.task_name)}</span> <span>·</span> <span>${escape_html(m.project_name)}</span> <span class="timestamp svelte-9oz8a8">${escape_html(formatRelative(m.created_at))}</span></div> <div class="preview svelte-9oz8a8">${escape_html(m.body)}</div></a>`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BB_RYlbh.js.map
