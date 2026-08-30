<script lang="ts">
  import { onMount } from 'svelte';
  import WebApp from '@twa-dev/sdk';
  import Search from './components/Search.svelte';
  import Player from './components/Player.svelte';
  import { currentTrack } from './stores/player';

  let tab: 'home' | 'search' | 'rooms' = 'search';

  onMount(() => {
    WebApp.ready();
    WebApp.expand();
    WebApp.setHeaderColor('#000000');
    WebApp.setBackgroundColor('#000000');
    document.documentElement.classList.add('dark');
  });
</script>

<main class="h-screen flex flex-col bg-black text-white overflow-hidden">
  <!-- Content -->
  <div class="flex-1 overflow-hidden">
    {#if tab === 'search'}
      <Search />
    {:else if tab === 'home'}
      <div class="p-6 flex flex-col items-center justify-center h-full text-center">
        <h1 class="text-2xl font-bold mb-2">Music</h1>
        <p class="text-zinc-400 text-sm max-w-xs">
          Hybrid Telegram Mini App + Go Streaming.<br />
          Zero spam. Sub-200ms buffering. Listen Together.
        </p>
      </div>
    {:else}
      <div class="p-6 flex flex-col items-center justify-center h-full text-center">
        <h2 class="text-lg font-semibold mb-2">Sync Rooms</h2>
        <p class="text-zinc-400 text-sm">Real-time Listen Together coming soon.</p>
      </div>
    {/if}
  </div>

  <!-- Mini player bar when track is loaded -->
  {#if $currentTrack}
    <button
      class="flex items-center gap-3 px-4 py-3 bg-zinc-950 border-t border-zinc-800 active:bg-zinc-900"
      on:click={() => { /* open full player via store or route */ }}
    >
      <div class="w-10 h-10 rounded overflow-hidden bg-zinc-800 flex-shrink-0">
        {#if $currentTrack.thumbnail}
          <img src={$currentTrack.thumbnail} class="w-full h-full object-cover" alt="" />
        {/if}
      </div>
      <div class="min-w-0 flex-1 text-left">
        <div class="text-sm font-medium truncate">{$currentTrack.title}</div>
        <div class="text-xs text-zinc-400 truncate">{$currentTrack.artist}</div>
      </div>
    </button>
  {/if}

  <!-- Bottom nav -->
  <nav class="flex border-t border-zinc-800 bg-black safe-bottom">
    <button
      class="flex-1 py-3 text-xs {tab === 'home' ? 'text-white' : 'text-zinc-500'}"
      on:click={() => (tab = 'home')}
    >Home</button>
    <button
      class="flex-1 py-3 text-xs {tab === 'search' ? 'text-white' : 'text-zinc-500'}"
      on:click={() => (tab = 'search')}
    >Search</button>
    <button
      class="flex-1 py-3 text-xs {tab === 'rooms' ? 'text-white' : 'text-zinc-500'}"
      on:click={() => (tab = 'rooms')}
    >Rooms</button>
  </nav>

  <!-- Full Player overlay -->
  <Player />
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
    background: #000;
  }
  .safe-bottom {
    padding-bottom: env(safe-area-inset-bottom);
  }
</style>
