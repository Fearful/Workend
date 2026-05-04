<script lang="ts">
  import Tooltip from './Tooltip.svelte';

  let {
    value,
    style = 'relative',
    placeholder = '—'
  }: {
    value: string | Date | null | undefined;
    style?: 'relative' | 'short' | 'absolute';
    placeholder?: string;
  } = $props();

  let date = $derived.by(() => {
    if (!value) return null;
    const d = value instanceof Date ? value : new Date(value);
    return Number.isFinite(d.getTime()) ? d : null;
  });

  function formatRelative(d: Date): string {
    const diff = (Date.now() - d.getTime()) / 1000;
    const past = diff >= 0;
    const abs = Math.abs(diff);
    if (abs < 5) return past ? 'just now' : 'imminent';
    const suffix = past ? ' ago' : '';
    const prefix = past ? '' : 'in ';
    if (abs < 60)         return `${prefix}${Math.round(abs)}s${suffix}`;
    if (abs < 3600)       return `${prefix}${Math.round(abs / 60)}m${suffix}`;
    if (abs < 86400)      return `${prefix}${Math.round(abs / 3600)}h${suffix}`;
    if (abs < 30 * 86400) return `${prefix}${Math.round(abs / 86400)}d${suffix}`;
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
  }

  function formatShort(d: Date): string {
    const now = new Date();
    const sameDay = d.toDateString() === now.toDateString();
    const yest = new Date(now); yest.setDate(yest.getDate() - 1);
    const isYest = d.toDateString() === yest.toDateString();
    if (sameDay) return d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' });
    if (isYest) return `Yesterday ${d.toLocaleTimeString(undefined, { hour: 'numeric', minute: '2-digit' })}`;
    if (d.getFullYear() === now.getFullYear()) {
      return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric' });
    }
    return d.toLocaleDateString(undefined, { month: 'short', day: 'numeric', year: 'numeric' });
  }

  let tick = $state(0);

  let display = $derived.by(() => {
    void tick;
    if (!date) return placeholder;
    if (style === 'absolute') return date.toLocaleString();
    if (style === 'short') return formatShort(date);
    return formatRelative(date);
  });

  let full = $derived(date ? date.toLocaleString() : '');

  $effect(() => {
    if (style !== 'relative' || !date) return;
    const int = setInterval(() => { tick++; }, 30_000);
    return () => clearInterval(int);
  });
</script>

{#if date}
  <Tooltip text={full}>
    <time datetime={date.toISOString()} class="ta">{display}</time>
  </Tooltip>
{:else}
  <span class="ta dim">{placeholder}</span>
{/if}

<style>
  .ta { font-variant-numeric: tabular-nums; }
  .dim { color: var(--text-dim); }
</style>
