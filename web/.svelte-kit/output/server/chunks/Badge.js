import { b as attr_class, s as stringify } from "./renderer.js";
/* empty css                                    */
function Badge($$renderer, $$props) {
  let { variant = "muted", size = "md", children } = $$props;
  $$renderer.push(`<span${attr_class(`badge badge-${stringify(variant)} size-${stringify(size)}`, "svelte-dtbgkf")}>`);
  children($$renderer);
  $$renderer.push(`<!----></span>`);
}
export {
  Badge as B
};
