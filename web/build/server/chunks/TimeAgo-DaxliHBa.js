import { a7 as attr, a8 as escape_html, a2 as derived } from './renderer-D0X3o35U.js';
import { T as Tooltip } from './Tooltip-Bg85MZ7d.js';

function TimeAgo($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { value, style = "relative", placeholder = "—" } = $$props;
    let date = derived(() => {
      if (!value) return null;
      const d = value instanceof Date ? value : new Date(value);
      return Number.isFinite(d.getTime()) ? d : null;
    });
    function formatRelative(d) {
      const diff = (Date.now() - d.getTime()) / 1e3;
      const past = diff >= 0;
      const abs = Math.abs(diff);
      if (abs < 5) return past ? "just now" : "imminent";
      const suffix = past ? " ago" : "";
      const prefix = past ? "" : "in ";
      if (abs < 60) return `${prefix}${Math.round(abs)}s${suffix}`;
      if (abs < 3600) return `${prefix}${Math.round(abs / 60)}m${suffix}`;
      if (abs < 86400) return `${prefix}${Math.round(abs / 3600)}h${suffix}`;
      if (abs < 30 * 86400) return `${prefix}${Math.round(abs / 86400)}d${suffix}`;
      return d.toLocaleDateString(void 0, { month: "short", day: "numeric" });
    }
    function formatShort(d) {
      const now = /* @__PURE__ */ new Date();
      const sameDay = d.toDateString() === now.toDateString();
      const yest = new Date(now);
      yest.setDate(yest.getDate() - 1);
      const isYest = d.toDateString() === yest.toDateString();
      if (sameDay) return d.toLocaleTimeString(void 0, { hour: "numeric", minute: "2-digit" });
      if (isYest) return `Yesterday ${d.toLocaleTimeString(void 0, { hour: "numeric", minute: "2-digit" })}`;
      if (d.getFullYear() === now.getFullYear()) {
        return d.toLocaleDateString(void 0, { month: "short", day: "numeric" });
      }
      return d.toLocaleDateString(void 0, { month: "short", day: "numeric", year: "numeric" });
    }
    let display = derived(() => {
      if (!date()) return placeholder;
      if (style === "absolute") return date().toLocaleString();
      if (style === "short") return formatShort(date());
      return formatRelative(date());
    });
    let full = derived(() => date() ? date().toLocaleString() : "");
    if (date()) {
      $$renderer2.push("<!--[0-->");
      Tooltip($$renderer2, {
        text: full(),
        children: ($$renderer3) => {
          $$renderer3.push(`<time${attr("datetime", date().toISOString())} class="ta svelte-c1otde">${escape_html(display())}</time>`);
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      $$renderer2.push(`<span class="ta dim svelte-c1otde">${escape_html(placeholder)}</span>`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { TimeAgo as T };
//# sourceMappingURL=TimeAgo-DaxliHBa.js.map
