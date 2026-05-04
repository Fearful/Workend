import { ad as attr_style, ae as stringify } from './renderer-mjPKoiGx.js';
import { a as statusColor } from './utils2-B5RqTmai.js';

function StatusDot($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { status, size = 8 } = $$props;
    $$renderer2.push(`<span class="dot"${attr_style(`background: ${stringify(statusColor(status))}; width: ${stringify(size)}px; height: ${stringify(size)}px;`)}></span>`);
  });
}

export { StatusDot as S };
//# sourceMappingURL=StatusDot-Cf5xUSPv.js.map
