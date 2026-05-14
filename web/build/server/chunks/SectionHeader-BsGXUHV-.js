import { a8 as escape_html } from './renderer-D0X3o35U.js';

function SectionHeader($$renderer, $$props) {
  let { title, subtitle, actions, level = "h2", children } = $$props;
  $$renderer.push(`<div class="sh svelte-ls99y2"><div class="sh-text svelte-ls99y2">`);
  if (level === "h1") {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<h1 class="sh-title sh-title-h1 svelte-ls99y2">`);
    if (children) {
      $$renderer.push("<!--[0-->");
      children($$renderer);
      $$renderer.push(`<!---->`);
    } else {
      $$renderer.push("<!--[-1-->");
      $$renderer.push(`${escape_html(title)}`);
    }
    $$renderer.push(`<!--]--></h1>`);
  } else if (level === "h3") {
    $$renderer.push("<!--[1-->");
    $$renderer.push(`<h3 class="sh-title sh-title-h3 svelte-ls99y2">`);
    if (children) {
      $$renderer.push("<!--[0-->");
      children($$renderer);
      $$renderer.push(`<!---->`);
    } else {
      $$renderer.push("<!--[-1-->");
      $$renderer.push(`${escape_html(title)}`);
    }
    $$renderer.push(`<!--]--></h3>`);
  } else {
    $$renderer.push("<!--[-1-->");
    $$renderer.push(`<h2 class="sh-title sh-title-h2 svelte-ls99y2">`);
    if (children) {
      $$renderer.push("<!--[0-->");
      children($$renderer);
      $$renderer.push(`<!---->`);
    } else {
      $$renderer.push("<!--[-1-->");
      $$renderer.push(`${escape_html(title)}`);
    }
    $$renderer.push(`<!--]--></h2>`);
  }
  $$renderer.push(`<!--]--> `);
  if (subtitle) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<span class="sh-sub svelte-ls99y2">${escape_html(subtitle)}</span>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--></div> `);
  if (actions) {
    $$renderer.push("<!--[0-->");
    $$renderer.push(`<div class="sh-actions svelte-ls99y2">`);
    actions($$renderer);
    $$renderer.push(`<!----></div>`);
  } else {
    $$renderer.push("<!--[-1-->");
  }
  $$renderer.push(`<!--]--></div>`);
}

export { SectionHeader as S };
//# sourceMappingURL=SectionHeader-BsGXUHV-.js.map
