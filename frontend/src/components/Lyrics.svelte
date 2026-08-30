<script lang="ts">
  import { position } from '../stores/player';
  import { onMount } from 'svelte';

  export let lrc = ''; // raw LRC string

  interface Line {
    time: number; // seconds
    text: string;
  }

  let lines: Line[] = [];
  let activeIndex = 0;
  let container: HTMLDivElement;

  $: lines = parseLRC(lrc);
  $: activeIndex = findActive($position, lines);

  $: if (container && activeIndex >= 0) {
    const el = container.children[activeIndex] as HTMLElement;
    if (el) {
      el.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }

  function parseLRC(raw: string): Line[] {
    if (!raw) return [];
    const result: Line[] = [];
    const regex = /\[(\d{2}):(\d{2})(?:\.(\d{2,3}))?\](.*)/g;
    let match;
    while ((match = regex.exec(raw)) !== null) {
      const min = parseInt(match[1], 10);
      const sec = parseInt(match[2], 10);
      const ms = match[3] ? parseInt(match[3].padEnd(3, '0').slice(0, 3), 10) : 0;
      const time = min * 60 + sec + ms / 1000;
      const text = match[4].trim();
      if (text) result.push({ time, text });
    }
    return result.sort((a, b) => a.time - b.time);
  }

  function findActive(pos: number, lines: Line[]): number {
    if (!lines.length) return -1;
    let idx = 0;
    for (let i = 0; i < lines.length; i++) {
      if (lines[i].time <= pos) idx = i;
      else break;
    }
    return idx;
  }
</script>

{#if lines.length}
  <div
    bind:this={container}
    class="h-40 overflow-y-auto scroll-smooth px-4 text-center space-y-3"
  >
    {#each lines as line, i}
      <p
        class="transition-all duration-300 text-sm leading-relaxed
               {i === activeIndex ? 'text-white text-base font-semibold scale-105' : 'text-zinc-500'}"
      >
        {line.text}
      </p>
    {/each}
  </div>
{:else}
  <div class="h-20 flex items-center justify-center text-zinc-600 text-xs">
    No lyrics available
  </div>
{/if}
