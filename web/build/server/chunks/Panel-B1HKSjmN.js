import { a9 as attr_class, ad as stringify, a8 as escape_html } from './renderer-D0X3o35U.js';

/* empty css                                    */
function Panel($$renderer, $$props) {
  let { title, actions, padding = "normal", children } = $$props;
  $$renderer.push(`<section${attr_class(`panel padding-${stringify(padding)}`, "svelte-hxsa5u")}>`);
  if (title || actions) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<div class="panel-head svelte-hxsa5u">`);
    if (title) {
      $$renderer.push("<!--[0-->");
      $$renderer.push(`<h2 class="panel-title svelte-hxsa5u">${escape_html(title)}</h2>`);
    } else {
      $$renderer.push("<!--[-1-->");
    }
    $$renderer.push(`<!--]--> `);
    if (actions) {
      $$renderer.push("<!--[0-->");
      $$renderer.push(`<div class="panel-actions svelte-hxsa5u">`);
      actions($$renderer);
      $$renderer.push(`<!----></div>`);
    } else {
      $$renderer.push("<!--[-1-->");
    }
    $$renderer.push(`<!--]--></div>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--> `);
  children($$renderer);
  $$renderer.push(`<!----></section>`);
}

export { Panel as P };
//# sourceMappingURL=Panel-B1HKSjmN.js.map
