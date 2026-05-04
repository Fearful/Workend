import { a9 as attr_class, ae as stringify } from './renderer-mjPKoiGx.js';

function Badge($$renderer, $$props) {
  let { variant = "muted", size = "md", children } = $$props;
  $$renderer.push(`<span${attr_class(`badge badge-${stringify(variant)} size-${stringify(size)}`, "svelte-dtbgkf")}>`);
  children($$renderer);
  $$renderer.push(`<!----></span>`);
}

export { Badge as B };
//# sourceMappingURL=Badge-BAVenLdx.js.map
