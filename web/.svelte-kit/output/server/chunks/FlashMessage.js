import { b as attr_class, a7 as clsx, a as attr } from "./renderer.js";
function FlashMessage($$renderer, $$props) {
  let { type = "error", children } = $$props;
  $$renderer.push(`<div${attr_class(clsx(type === "error" ? "error-banner" : "success-banner"))}${attr("role", type === "error" ? "alert" : "status")}>`);
  children($$renderer);
  $$renderer.push(`<!----></div>`);
}
export {
  FlashMessage as F
};
