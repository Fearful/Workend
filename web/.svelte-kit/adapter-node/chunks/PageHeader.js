import { e as escape_html } from "./renderer.js";
import "clsx";
function PageHeader($$renderer, $$props) {
  let { title, actions, children } = $$props;
  $$renderer.push(`<div class="header-row svelte-162svzm"><h1 class="svelte-162svzm">`);
  if (children) {
    $$renderer.push("<!--[0-->");
    children($$renderer);
    $$renderer.push(`<!---->`);
  } else {
    $$renderer.push("<!--[-1-->");
    $$renderer.push(`${escape_html(title)}`);
  }
  $$renderer.push(`<!--]--></h1> `);
  if (actions) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<div class="actions svelte-162svzm">`);
    actions($$renderer);
    $$renderer.push(`<!----></div>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--></div>`);
}
export {
  PageHeader as P
};
