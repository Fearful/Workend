import { ac as ensure_array_like, a7 as attr, ad as attr_style, ae as stringify, a8 as escape_html, a2 as derived } from './renderer-mjPKoiGx.js';
import { P as Panel } from './Panel-BruuJ27z.js';
import { E as EmptyState } from './EmptyState-Foiy8g97.js';

function _page($$renderer, $$props) {
  $$renderer.component(($$renderer2) => {
    let { data } = $$props;
    function taskColor(taskID) {
      let h = 0;
      for (let i = 0; i < taskID.length; i++) h = h * 31 + taskID.charCodeAt(i) | 0;
      return `hsl(${Math.abs(h) % 360}, 65%, 60%)`;
    }
    function formatDur(sec) {
      if (sec < 60) return `${sec}s`;
      if (sec < 3600) return `${Math.floor(sec / 60)}m ${sec % 60}s`;
      return `${Math.floor(sec / 3600)}h ${Math.floor(sec % 3600 / 60)}m`;
    }
    const W = 760;
    const H = 320;
    const pad = { top: 16, right: 16, bottom: 28, left: 48 };
    const innerW = W - pad.left - pad.right;
    const innerH = H - pad.top - pad.bottom;
    let domain = derived(() => {
      if (data.trends.points.length === 0) {
        return { minT: 0, maxT: 1, maxD: 1 };
      }
      let minT = Infinity, maxT = -Infinity, maxD = 0;
      for (const p of data.trends.points) {
        const t = new Date(p.finished_at).getTime();
        if (t < minT) minT = t;
        if (t > maxT) maxT = t;
        if (p.duration_sec > maxD) maxD = p.duration_sec;
      }
      if (minT === maxT) maxT = minT + 1;
      if (maxD === 0) maxD = 1;
      return { minT, maxT, maxD };
    });
    function xPos(tISO) {
      const t = new Date(tISO).getTime();
      return pad.left + (t - domain().minT) / (domain().maxT - domain().minT) * innerW;
    }
    function yPos(sec) {
      return pad.top + innerH - sec / domain().maxD * innerH;
    }
    let series = derived(() => {
      const byTask = /* @__PURE__ */ new Map();
      for (const p of data.trends.points) {
        const list = byTask.get(p.task_id) ?? [];
        list.push(p);
        byTask.set(p.task_id, list);
      }
      return Array.from(byTask.entries()).map(([taskID, pts]) => ({
        taskID,
        taskName: pts[0].task_name,
        taskSource: pts[0].task_source,
        color: taskColor(taskID),
        points: pts
      }));
    });
    function pathFor(pts) {
      return pts.map((p, i) => `${i === 0 ? "M" : "L"} ${xPos(p.finished_at).toFixed(1)} ${yPos(p.duration_sec).toFixed(1)}`).join(" ");
    }
    let yTicks = derived(() => {
      const out = [];
      for (let i = 0; i <= 4; i++) {
        const v = domain().maxD * i / 4;
        out.push({ v, y: yPos(v) });
      }
      return out;
    });
    function fmtDate(ms) {
      const d = new Date(ms);
      return `${d.getMonth() + 1}/${d.getDate()}`;
    }
    let heatmap = derived(() => {
      const counts = /* @__PURE__ */ new Map();
      for (const c of data.activity.cells) counts.set(c.date, c);
      const today = /* @__PURE__ */ new Date();
      today.setHours(0, 0, 0, 0);
      const end = new Date(today);
      end.setDate(end.getDate() + (6 - end.getDay()));
      const start = new Date(end);
      start.setDate(start.getDate() - 7 * 53 + 1);
      const cols = [];
      let cursor = new Date(start);
      let week = [];
      while (cursor <= end) {
        const iso = cursor.toISOString().slice(0, 10);
        week.push({ date: iso, cell: counts.get(iso) ?? null });
        if (week.length === 7) {
          cols.push(week);
          week = [];
        }
        cursor = new Date(cursor);
        cursor.setDate(cursor.getDate() + 1);
      }
      if (week.length > 0) cols.push(week);
      let max = 0;
      for (const c of data.activity.cells) if (c.total > max) max = c.total;
      return { cols, max };
    });
    function cellColor(c, max) {
      if (!c || c.total === 0) return "var(--bg-hover)";
      const intensity = Math.min(1, c.total / Math.max(1, max));
      if (c.failed > 0) {
        const a2 = 0.25 + 0.55 * intensity;
        return `rgba(239, 68, 68, ${a2})`;
      }
      const a = 0.25 + 0.55 * intensity;
      return `rgba(34, 197, 94, ${a})`;
    }
    function cellTooltip(d) {
      if (!d.cell) return `${d.date} — no runs`;
      const parts = [`${d.cell.total} runs`];
      if (d.cell.succeeded) parts.push(`${d.cell.succeeded} ok`);
      if (d.cell.failed) parts.push(`${d.cell.failed} failed`);
      if (d.cell.cancelled) parts.push(`${d.cell.cancelled} cancelled`);
      return `${d.date} — ${parts.join(", ")}`;
    }
    let xTicks = derived(() => {
      const out = [];
      for (let i = 0; i <= 4; i++) {
        const t = domain().minT + (domain().maxT - domain().minT) * i / 4;
        out.push({ label: fmtDate(t), x: pad.left + innerW * i / 4 });
      }
      return out;
    });
    function navWith(params) {
      const p = new URLSearchParams(window.location.search);
      for (const [k, v] of Object.entries(params)) {
        if (v === "") p.delete(k);
        else p.set(k, v);
      }
      window.location.search = p.toString();
    }
    $$renderer2.push(`<h2 class="section-title svelte-tdlzdu">Run duration trends</h2> `);
    if (heatmap().cols.length > 0) {
      $$renderer2.push("<!--[0-->");
      $$renderer2.push(`<section class="panel heatmap-section svelte-tdlzdu"><div class="heatmap-header svelte-tdlzdu">Activity (last 53 weeks)</div> <div class="heatmap-grid svelte-tdlzdu"><!--[-->`);
      const each_array = ensure_array_like(heatmap().cols);
      for (let wi = 0, $$length = each_array.length; wi < $$length; wi++) {
        let week = each_array[wi];
        $$renderer2.push(`<div class="heatmap-week svelte-tdlzdu"><!--[-->`);
        const each_array_1 = ensure_array_like(week);
        for (let di = 0, $$length2 = each_array_1.length; di < $$length2; di++) {
          let day = each_array_1[di];
          $$renderer2.push(`<div${attr("title", cellTooltip(day))} class="heatmap-cell svelte-tdlzdu"${attr_style(`background: ${stringify(cellColor(day.cell, heatmap().max))};`)}></div>`);
        }
        $$renderer2.push(`<!--]--></div>`);
      }
      $$renderer2.push(`<!--]--></div> <div class="heatmap-legend svelte-tdlzdu"><span>less</span> <span class="heatmap-legend-cell svelte-tdlzdu" style="background:var(--bg-hover);"></span> <span class="heatmap-legend-cell svelte-tdlzdu" style="background:rgba(34,197,94,0.4);"></span> <span class="heatmap-legend-cell svelte-tdlzdu" style="background:rgba(34,197,94,0.7);"></span> <span>more</span> <span class="heatmap-legend-divider svelte-tdlzdu">·</span> <span class="heatmap-legend-cell svelte-tdlzdu" style="background:rgba(239,68,68,0.5);"></span> <span>has failures</span></div></section>`);
    } else {
      $$renderer2.push("<!--[-1-->");
    }
    $$renderer2.push(`<!--]--> <div class="controls svelte-tdlzdu"><label for="task" class="svelte-tdlzdu">Task:</label> `);
    $$renderer2.select(
      {
        id: "task",
        value: data.filterTaskID,
        onchange: (e) => navWith({ task_id: e.currentTarget.value }),
        class: ""
      },
      ($$renderer3) => {
        $$renderer3.option({ value: "" }, ($$renderer4) => {
          $$renderer4.push(`All tasks`);
        });
        $$renderer3.push(`<!--[-->`);
        const each_array_2 = ensure_array_like(data.tasks);
        for (let $$index_2 = 0, $$length = each_array_2.length; $$index_2 < $$length; $$index_2++) {
          let t = each_array_2[$$index_2];
          $$renderer3.option({ value: t.id }, ($$renderer4) => {
            $$renderer4.push(`${escape_html(t.source)} · ${escape_html(t.name)}`);
          });
        }
        $$renderer3.push(`<!--]-->`);
      },
      "svelte-tdlzdu"
    );
    $$renderer2.push(` <label for="days" class="svelte-tdlzdu">Window:</label> `);
    $$renderer2.select(
      {
        id: "days",
        value: String(data.days),
        onchange: (e) => navWith({ days: e.currentTarget.value }),
        class: ""
      },
      ($$renderer3) => {
        $$renderer3.option({ value: "7" }, ($$renderer4) => {
          $$renderer4.push(`7 days`);
        });
        $$renderer3.option({ value: "30" }, ($$renderer4) => {
          $$renderer4.push(`30 days`);
        });
        $$renderer3.option({ value: "90" }, ($$renderer4) => {
          $$renderer4.push(`90 days`);
        });
        $$renderer3.option({ value: "180" }, ($$renderer4) => {
          $$renderer4.push(`180 days`);
        });
        $$renderer3.option({ value: "365" }, ($$renderer4) => {
          $$renderer4.push(`365 days`);
        });
      },
      "svelte-tdlzdu"
    );
    $$renderer2.push(`</div> `);
    if (data.trends.points.length === 0) {
      $$renderer2.push("<!--[0-->");
      Panel($$renderer2, {
        children: ($$renderer3) => {
          EmptyState($$renderer3, { message: `No completed runs in the last ${data.days} days.` });
        }
      });
    } else {
      $$renderer2.push("<!--[-1-->");
      const total = data.trends.points.length;
      const succ = data.trends.points.filter((p) => p.status === "succeeded").length;
      const fail = data.trends.points.filter((p) => p.status === "failed").length;
      const avg = Math.round(data.trends.points.reduce((s, p) => s + p.duration_sec, 0) / total);
      const max = Math.max(...data.trends.points.map((p) => p.duration_sec));
      $$renderer2.push(`<div class="summary svelte-tdlzdu"><div class="stat svelte-tdlzdu"><div class="stat-num svelte-tdlzdu">${escape_html(total)}</div><div class="stat-label svelte-tdlzdu">Runs</div></div> <div class="stat svelte-tdlzdu"><div class="stat-num svelte-tdlzdu">${escape_html(Math.round(succ / total * 100))}%</div><div class="stat-label svelte-tdlzdu">Success rate</div></div> <div class="stat svelte-tdlzdu"><div class="stat-num svelte-tdlzdu">${escape_html(fail)}</div><div class="stat-label svelte-tdlzdu">Failures</div></div> <div class="stat svelte-tdlzdu"><div class="stat-num svelte-tdlzdu">${escape_html(formatDur(avg))}</div><div class="stat-label svelte-tdlzdu">Average</div></div> <div class="stat svelte-tdlzdu"><div class="stat-num svelte-tdlzdu">${escape_html(formatDur(max))}</div><div class="stat-label svelte-tdlzdu">Slowest</div></div></div> <section class="panel chart-panel svelte-tdlzdu"><svg${attr("viewBox", `0 0 ${stringify(W)} ${stringify(H)}`)} role="img" aria-label="Duration trend chart" class="svelte-tdlzdu"><!--[-->`);
      const each_array_3 = ensure_array_like(yTicks());
      for (let $$index_3 = 0, $$length = each_array_3.length; $$index_3 < $$length; $$index_3++) {
        let t = each_array_3[$$index_3];
        $$renderer2.push(`<line class="grid-line svelte-tdlzdu"${attr("x1", pad.left)}${attr("x2", W - pad.right)}${attr("y1", t.y)}${attr("y2", t.y)}></line><text class="axis-label svelte-tdlzdu"${attr("x", pad.left - 6)}${attr("y", t.y + 3)} text-anchor="end">${escape_html(formatDur(Math.round(t.v)))}</text>`);
      }
      $$renderer2.push(`<!--]--><!--[-->`);
      const each_array_4 = ensure_array_like(xTicks());
      for (let i = 0, $$length = each_array_4.length; i < $$length; i++) {
        let t = each_array_4[i];
        $$renderer2.push(`<text class="axis-label svelte-tdlzdu"${attr("x", t.x)}${attr("y", H - 8)} text-anchor="middle">${escape_html(t.label)}</text>`);
      }
      $$renderer2.push(`<!--]--><!--[-->`);
      const each_array_5 = ensure_array_like(series());
      for (let $$index_6 = 0, $$length = each_array_5.length; $$index_6 < $$length; $$index_6++) {
        let s = each_array_5[$$index_6];
        $$renderer2.push(`<path${attr("d", pathFor(s.points))} fill="none"${attr("stroke", s.color)} stroke-width="1.5" stroke-linejoin="round"></path><!--[-->`);
        const each_array_6 = ensure_array_like(s.points);
        for (let $$index_5 = 0, $$length2 = each_array_6.length; $$index_5 < $$length2; $$index_5++) {
          let p = each_array_6[$$index_5];
          $$renderer2.push(`<circle${attr("cx", xPos(p.finished_at))}${attr("cy", yPos(p.duration_sec))} r="3"${attr("fill", p.status === "succeeded" ? s.color : "transparent")}${attr("stroke", s.color)} stroke-width="1.5" role="button" tabindex="0" class="point-circle svelte-tdlzdu"${attr("aria-label", `${p.task_name} run · ${formatDur(p.duration_sec)} · ${p.status}`)}></circle>`);
        }
        $$renderer2.push(`<!--]-->`);
      }
      $$renderer2.push(`<!--]-->`);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--></svg> `);
      {
        $$renderer2.push("<!--[-1-->");
      }
      $$renderer2.push(`<!--]--> <div class="legend svelte-tdlzdu"><!--[-->`);
      const each_array_7 = ensure_array_like(series());
      for (let $$index_7 = 0, $$length = each_array_7.length; $$index_7 < $$length; $$index_7++) {
        let s = each_array_7[$$index_7];
        $$renderer2.push(`<span class="legend-item svelte-tdlzdu"><span class="swatch svelte-tdlzdu"${attr_style(`background:${stringify(s.color)}`)}></span> ${escape_html(s.taskName)} <span style="color:var(--text-dim)">(${escape_html(s.taskSource)}, ${escape_html(s.points.length)})</span></span>`);
      }
      $$renderer2.push(`<!--]--></div></section>`);
    }
    $$renderer2.push(`<!--]-->`);
  });
}

export { _page as default };
//# sourceMappingURL=_page.svelte-BQpE39IM.js.map
