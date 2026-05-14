import { a9 as attr_class, af as clsx, a7 as attr } from './renderer-D0X3o35U.js';

function FlashMessage($$renderer, $$props) {
  let { type = "error", children } = $$props;
  $$renderer.push(`<div${attr_class(clsx(type === "error" ? "error-banner" : "success-banner"))}${attr("role", type === "error" ? "alert" : "status")}>`);
  children($$renderer);
  $$renderer.push(`<!----></div>`);
}

export { FlashMessage as F };
//# sourceMappingURL=FlashMessage-CReJ7kuH.js.map
