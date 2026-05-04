import { h as head } from "../../../chunks/renderer.js";
function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    head("1s1slfy", $$renderer2, ($$renderer3) => {
      $$renderer3.title(($$renderer4) => {
        $$renderer4.push(`<title>Dockscope | Workend</title>`);
      });
    });
    $$renderer2.push(`<div class="iframe-panel svelte-1s1slfy">`);
    {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<div class="empty">Loading Dockscope...</div>`);
    }
    $$renderer2.push(`<!--]--></div>`);
  });
}
export {
  _page as default
};
