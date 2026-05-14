import { o as onDestroy } from './index-server-BFLhAcPs.js';

function Tooltip($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let {
      text,
      content,
      children,
      placement = "top",
      delay = 200,
      maxWidth = 280
    } = $$props;
    onDestroy(() => {
    });
    $$renderer2.push(`<span class="tooltip-wrap svelte-11extwn">`);
    children($$renderer2);
    $$renderer2.push(`<!----></span> `);
    {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { Tooltip as T };
//# sourceMappingURL=Tooltip-Bg85MZ7d.js.map
