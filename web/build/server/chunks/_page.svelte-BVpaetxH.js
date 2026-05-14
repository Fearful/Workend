import { a8 as escape_html, ac as ensure_array_like, a9 as attr_class, ad as stringify, ae as attr_style, a2 as derived } from './renderer-D0X3o35U.js';
import { s as shortSha, a as statusColor } from './utils2-BUPlP7zG.js';
import { B as Breadcrumb } from './Breadcrumb-Bl_x71bg.js';
import { P as PageHeader } from './PageHeader-C8D5NSCp.js';

function StatusDot($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { status, size = 8 } = $$props;
    $$renderer2.push(`<span class="dot"${attr_style(`background: ${stringify(statusColor(status))}; width: ${stringify(size)}px; height: ${stringify(size)}px;`)}></span>`);
  });
}
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    let stats = derived(() => {
      let added = 0, removed = 0, equal = 0;
      for (const h of data.diff) {
        if (h.op === 0) equal++;
        else if (h.op === 1) added++;
        else removed++;
      }
      return { added, removed, equal };
    });
    Breadcrumb($$renderer2, {
      segments: [{ label: "← back to run", href: `/runs/${data.left.id}` }]
    });
    $$renderer2.push(`<!----> `);
    PageHeader($$renderer2, { title: "Run comparison" });
    $$renderer2.push(`<!----> <div class="meta-row svelte-1yyjyuw"><div class="panel svelte-1yyjyuw"><h2 class="svelte-1yyjyuw">Left (older)</h2> <div class="meta-line svelte-1yyjyuw">`);
    StatusDot($$renderer2, { status: data.left.status });
    $$renderer2.push(`<!----> <span class="label svelte-1yyjyuw">status</span> ${escape_html(data.left.status)} · exit ${escape_html(data.left.exit_code ?? "—")}</div> <div class="meta-line svelte-1yyjyuw"><span class="label svelte-1yyjyuw">commit</span> ${escape_html(shortSha(data.left.commit_sha))}</div> <div class="meta-line svelte-1yyjyuw"><span class="label svelte-1yyjyuw">started</span> ${escape_html(data.left.started_at ? new Date(data.left.started_at).toLocaleString() : "—")}</div></div> <div class="panel svelte-1yyjyuw"><h2 class="svelte-1yyjyuw">Right (newer)</h2> <div class="meta-line svelte-1yyjyuw">`);
    StatusDot($$renderer2, { status: data.right.status });
    $$renderer2.push(`<!----> <span class="label svelte-1yyjyuw">status</span> ${escape_html(data.right.status)} · exit ${escape_html(data.right.exit_code ?? "—")}</div> <div class="meta-line svelte-1yyjyuw"><span class="label svelte-1yyjyuw">commit</span> ${escape_html(shortSha(data.right.commit_sha))}</div> <div class="meta-line svelte-1yyjyuw"><span class="label svelte-1yyjyuw">started</span> ${escape_html(data.right.started_at ? new Date(data.right.started_at).toLocaleString() : "—")}</div></div></div> <div class="summary svelte-1yyjyuw"><span class="added svelte-1yyjyuw">+${escape_html(stats().added)}</span> <span class="removed svelte-1yyjyuw">-${escape_html(stats().removed)}</span> <span class="equal svelte-1yyjyuw">${escape_html(stats().equal)} unchanged</span></div> <pre class="diff svelte-1yyjyuw"><!--[-->`);
    const each_array = ensure_array_like(data.diff);
    for (let i = 0, $$length = each_array.length; i < $$length; i++) {
      let h = each_array[i];
      $$renderer2.push(`<span${attr_class(`line ${stringify(h.op === 0 ? "eq" : h.op === 1 ? "add" : "del")}`, "svelte-1yyjyuw")}><span class="marker svelte-1yyjyuw">${escape_html(h.op === 0 ? " " : h.op === 1 ? "+" : "-")}</span>${escape_html(h.text.replace(/\n$/, ""))}
</span>`);
    }
    $$renderer2.push(`<!--]--></pre>`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BVpaetxH.js.map
