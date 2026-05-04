import { c as ensure_array_like, a as attr, e as escape_html } from "./renderer.js";
function Breadcrumb($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { segments } = $$props;
    $$renderer2.push(`<div class="breadcrumb svelte-mhuuw7"><!--[-->`);
    const each_array = ensure_array_like(segments);
    for (let i = 0, $$length = each_array.length; i < $$length; i++) {
      let seg = each_array[i];
      if (seg.href) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<a${attr("href", seg.href)}>${escape_html(seg.label)}</a>`);
      } else {
        $$renderer2.push("<!--[-1-->");
        $$renderer2.push(`<span class="current svelte-mhuuw7">${escape_html(seg.label)}</span>`);
      }
      $$renderer2.push(`<!--]--> `);
      if (i < segments.length - 1) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<span class="sep svelte-mhuuw7">/</span>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]-->`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
export {
  Breadcrumb as B
};
