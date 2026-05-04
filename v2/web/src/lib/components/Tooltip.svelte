<script lang="ts">
  import type { Snippet } from 'svelte';
  import { onMount, onDestroy } from 'svelte';

  let {
    text,
    content,
    children,
    placement = 'top',
    delay = 200,
    maxWidth = 280
  }: {
    text?: string;
    content?: Snippet;
    children: Snippet;
    placement?: 'top' | 'bottom' | 'left' | 'right';
    delay?: number;
    maxWidth?: number;
  } = $props();

  let wrapper: HTMLSpanElement | undefined = $state();
  let visible = $state(false);
  let pos = $state<{ x: number; y: number; place: 'top' | 'bottom' | 'left' | 'right' }>({ x: 0, y: 0, place: 'top' });
  let timer: ReturnType<typeof setTimeout> | null = null;

  function compute(target: HTMLElement) {
    const rect = target.getBoundingClientRect();
    const margin = 8;
    let x = 0, y = 0;
    let place = placement;
    switch (placement) {
      case 'top': x = rect.left + rect.width / 2; y = rect.top - margin; break;
      case 'bottom': x = rect.left + rect.width / 2; y = rect.bottom + margin; break;
      case 'left': x = rect.left - margin; y = rect.top + rect.height / 2; break;
      case 'right': x = rect.right + margin; y = rect.top + rect.height / 2; break;
    }
    if (place === 'top' && y < 50) { place = 'bottom'; y = rect.bottom + margin; }
    pos = { x, y, place };
  }

  function show(e: MouseEvent | FocusEvent) {
    const target = e.currentTarget as HTMLElement;
    if (timer) clearTimeout(timer);
    timer = setTimeout(() => { compute(target); visible = true; }, delay);
  }

  function hide() {
    if (timer) { clearTimeout(timer); timer = null; }
    visible = false;
  }

  onMount(() => {
    if (!wrapper) return;
    const child = wrapper.firstElementChild as HTMLElement | null;
    if (!child) return;
    child.addEventListener('mouseenter', show);
    child.addEventListener('mouseleave', hide);
    child.addEventListener('focusin', show);
    child.addEventListener('focusout', hide);
  });

  onDestroy(() => {
    if (timer) clearTimeout(timer);
  });
</script>

<span bind:this={wrapper} class="tooltip-wrap">{@render children()}</span>

{#if visible}
  <div class="tooltip tooltip-{pos.place}"
       style="left:{pos.x}px; top:{pos.y}px; max-width:{maxWidth}px;"
       role="tooltip">
    {#if content}{@render content()}
    {:else if text}{text}
    {/if}
  </div>
{/if}

<style>
  .tooltip-wrap { display: inline-flex; }
  .tooltip {
    position: fixed;
    z-index: 1000;
    background: var(--bg-panel);
    color: var(--text);
    font-size: var(--fs-xs);
    line-height: var(--lh-normal);
    padding: 0.4rem 0.625rem;
    border: 1px solid var(--border-strong);
    border-radius: var(--radius-md);
    box-shadow: var(--shadow-popover);
    pointer-events: none;
    white-space: normal;
  }
  .tooltip-top { transform: translate(-50%, -100%); }
  .tooltip-bottom { transform: translate(-50%, 0); }
  .tooltip-left { transform: translate(-100%, -50%); }
  .tooltip-right { transform: translate(0, -50%); }
  @media (max-width: 640px) {
    .tooltip { display: none; }
  }
</style>
