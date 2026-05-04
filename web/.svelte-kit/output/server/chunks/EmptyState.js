import { e as escape_html, a as attr } from "./renderer.js";
/* empty css                                         */
function EmptyState($$renderer, $$props) {
  let { icon, message, actionHref, actionLabel, children } = $$props;
  $$renderer.push(`<div class="empty svelte-13862ru">`);
  if (icon) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<div class="icon svelte-13862ru">${escape_html(icon)}</div>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--> `);
  if (children) {
    $$renderer.push("<!--[0-->");
    children($$renderer);
    $$renderer.push(`<!---->`);
  } else {
    $$renderer.push("<!--[-1-->");
    if (message) {
      $$renderer.push("<!--[0-->");
      $$renderer.push(`<p class="message svelte-13862ru">${escape_html(message)}</p>`);
    } else {
      $$renderer.push("<!--[-1-->");
    }
    $$renderer.push(`<!--]--> `);
    if (actionHref) {
      $$renderer.push("<!--[0-->");
      $$renderer.push(`<a${attr("href", actionHref)} class="action svelte-13862ru">${escape_html(actionLabel ?? "Get started")}</a>`);
    } else {
      $$renderer.push("<!--[-1-->");
    }
    $$renderer.push(`<!--]-->`);
  }
  $$renderer.push(`<!--]--></div>`);
}
export {
  EmptyState as E
};
