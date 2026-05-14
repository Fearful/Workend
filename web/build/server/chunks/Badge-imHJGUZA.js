import { a9 as attr_class, ad as stringify } from './renderer-D0X3o35U.js';

/* empty css                                    */
function Badge($$renderer, $$props) {
  let { variant = "muted", size = "md", children } = $$props;
  $$renderer.push(`<span${attr_class(`badge badge-${stringify(variant)} size-${stringify(size)}`, "svelte-dtbgkf")}>`);
  children($$renderer);
  $$renderer.push(`<!----></span>`);
}

export { Badge as B };
//# sourceMappingURL=Badge-imHJGUZA.js.map
