import { a9 as attr_class, ad as stringify, a8 as escape_html, a2 as derived } from './renderer-D0X3o35U.js';

function StatusPill($$renderer, $$props) {
  let { status, size = "md", label } = $$props;
  function tone(s) {
    const lc = s.toLowerCase();
    if (lc === "ready" || lc === "succeeded" || lc === "success" || lc === "ok" || lc === "running-ok") return "success";
    if (lc === "running" || lc === "starting") return "info";
    if (lc === "cloning" || lc === "pending" || lc === "queued" || lc === "pause" || lc === "paused") return "warning";
    if (lc === "failed" || lc === "failure" || lc === "error") return "danger";
    return "neutral";
  }
  let t = derived(() => tone(status));
  let display = derived(() => label ?? status);
  $$renderer.push(`<span${attr_class(`pill pill-${stringify(t())} pill-${stringify(size)}`, "svelte-1swmi23")}><span class="dot svelte-1swmi23"></span> <span class="text svelte-1swmi23">${escape_html(display())}</span></span>`);
}

export { StatusPill as S };
//# sourceMappingURL=StatusPill-j6_a2per.js.map
