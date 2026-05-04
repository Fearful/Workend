import { a7 as attr, ad as attr_style, a8 as escape_html, ae as stringify } from './renderer-mjPKoiGx.js';

function Modal($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { open, title, width = 420, onClose, children, footer } = $$props;
    if (open) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<div role="dialog" aria-modal="true"${attr("aria-label", title)} tabindex="-1" class="modal-backdrop svelte-ta60gp"><div class="modal-content svelte-ta60gp"${attr_style(`width: min(${stringify(width)}px, 90vw);`)}>`);
      if (title) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<h2 class="modal-title svelte-ta60gp">${escape_html(title)}</h2>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> `);
      children($$renderer2);
      $$renderer2.push(`<!----> `);
      if (footer) {
        $$renderer2.push("<!--[0-->");
        $$renderer2.push(`<div class="modal-footer svelte-ta60gp">`);
        footer($$renderer2);
        $$renderer2.push(`<!----></div>`);
      } else {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></div></div>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { Modal as M };
//# sourceMappingURL=Modal-BJw4bo_R.js.map
