import { c as ensure_array_like, a as attr, e as escape_html, f as attr_style, s as stringify } from "../../../../../chunks/renderer.js";
import { s as shortSha } from "../../../../../chunks/utils2.js";
import { E as EmptyState } from "../../../../../chunks/EmptyState.js";
import { S as SectionHeader } from "../../../../../chunks/SectionHeader.js";
import { T as TimeAgo } from "../../../../../chunks/TimeAgo.js";
import "../../../../../chunks/Tooltip.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    function shortDigest(d) {
      return d ? d.slice(0, 24) + "…" : "—";
    }
    function formatBytes(b) {
      if (!b) return "—";
      const units = ["B", "KB", "MB", "GB"];
      let v = b;
      let i = 0;
      while (v >= 1024 && i < units.length - 1) {
        v /= 1024;
        i++;
      }
      return `${v.toFixed(1)} ${units[i]}`;
    }
    function sevColor(s) {
      return {
        critical: "#ef4444",
        high: "#f97316",
        medium: "#eab308",
        low: "#60a5fa",
        unknown: "var(--text-dim)"
      }[s];
    }
    SectionHeader($$renderer2, { title: "Images" });
    $$renderer2.push(`<!----> `);
    if (data.images.length === 0) {
      $$renderer2.push("<!--[0-->");
      EmptyState($$renderer2, {
        icon: "◇",
        children: ($$renderer3) => {
          $$renderer3.push(`<!---->No images built yet. Add a <code>Dockerfile</code> to the repo and run the <code>dockerfile: build</code> task.`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<!--[-->`);
      const each_array = ensure_array_like(data.images);
      for (let $$index_1 = 0, $$length = each_array.length; $$index_1 < $$length; $$index_1++) {
        let img = each_array[$$index_1];
        $$renderer2.push(`<div class="row svelte-1lzg3hg">`);
        if (img.run_id) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<a class="mono image-link svelte-1lzg3hg"${attr("href", `/runs/${img.run_id}`)} title="View the build run">${escape_html(shortDigest(img.digest))}</a>`);
        } else {
          $$renderer2.push("<!--[-1-->");
          $$renderer2.push(`<span class="mono svelte-1lzg3hg">${escape_html(shortDigest(img.digest))}</span>`);
        }
        $$renderer2.push(`<!--]--> <span class="muted svelte-1lzg3hg">${escape_html(img.dockerfile_path)}</span> <span class="muted svelte-1lzg3hg">commit ${escape_html(shortSha(img.commit_sha))}</span> <span class="muted svelte-1lzg3hg">${escape_html(formatBytes(img.size_bytes))}</span> <span class="muted svelte-1lzg3hg">`);
        TimeAgo($$renderer2, { value: img.built_at });
        $$renderer2.push(`<!----></span></div> `);
        if (img.scan_status) {
          $$renderer2.push("<!--[0-->");
          $$renderer2.push(`<div class="scan-row svelte-1lzg3hg">`);
          if (img.scan_status === "pending") {
            $$renderer2.push("<!--[0-->");
            $$renderer2.push(`<span class="scan-pending svelte-1lzg3hg">scan in progress…</span>`);
          } else if (img.scan_status === "error") {
            $$renderer2.push("<!--[1-->");
            $$renderer2.push(`<span class="scan-error svelte-1lzg3hg">scan failed</span>`);
          } else if (img.scan_status === "ok" && img.vuln_summary) {
            $$renderer2.push("<!--[2-->");
            const s = img.vuln_summary;
            if (s.critical > 0) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="sev-pill critical svelte-1lzg3hg">${escape_html(s.critical)} critical</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            if (s.high > 0) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="sev-pill high svelte-1lzg3hg">${escape_html(s.high)} high</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            if (s.medium > 0) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="sev-pill medium svelte-1lzg3hg">${escape_html(s.medium)} medium</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            if (s.low > 0) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="sev-pill low svelte-1lzg3hg">${escape_html(s.low)} low</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            if (s.critical === 0 && s.high === 0 && s.medium === 0 && s.low === 0) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<span class="scan-ok svelte-1lzg3hg">no vulnerabilities</span>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]--> `);
            if (s.top.length > 0) {
              $$renderer2.push("<!--[0-->");
              $$renderer2.push(`<details><summary class="findings-summary svelte-1lzg3hg">top findings</summary> <ul class="findings-list svelte-1lzg3hg"><!--[-->`);
              const each_array_1 = ensure_array_like(s.top);
              for (let $$index = 0, $$length2 = each_array_1.length; $$index < $$length2; $$index++) {
                let v = each_array_1[$$index];
                $$renderer2.push(`<li class="svelte-1lzg3hg"><span${attr_style(`color: ${stringify(sevColor(v.severity.toLowerCase()))};`)}>${escape_html(v.severity)}</span> · ${escape_html(v.id)} · ${escape_html(v.package)}${escape_html(v.fixed_in ? ` → fix ${v.fixed_in}` : "")}</li>`);
              }
              $$renderer2.push(`<!--]--></ul></details>`);
            } else {
              $$renderer2.push("<!--[-1-->");
            }
            $$renderer2.push(`<!--]-->`);
          } else {
            $$renderer2.push("<!--[-1-->");
          }
          $$renderer2.push(`<!--]--></div>`);
        } else {
          $$renderer2.push("<!--[-1-->");
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}
export {
  _page as default
};
