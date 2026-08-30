<script lang="ts">
  import { onMount } from 'svelte';
  import WebApp from '@twa-dev/sdk';

  let theme = 'dark';

  onMount(() => {
    WebApp.ready();
    WebApp.expand();
    theme = WebApp.colorScheme;
    document.documentElement.classList.toggle('dark', theme === 'dark');
  });
</script>

<main class="min-h-screen bg-black text-white flex flex-col">
  <header class="p-4 border-b border-zinc-800">
    <h1 class="text-xl font-bold tracking-tight">Music</h1>
    <p class="text-sm text-zinc-400">Spotify inside Telegram</p>
  </header>

  <section class="flex-1 flex items-center justify-center p-6">
    <div class="text-center space-y-4">
      <div class="w-48 h-48 mx-auto rounded-2xl bg-zinc-900 flex items-center justify-center">
        <span class="text-zinc-600 text-sm">Album Art</span>
      </div>
      <h2 class="text-lg font-semibold">Ready for blazing fast playback</h2>
      <p class="text-zinc-400 text-sm">UI scaffold • Backend connected soon</p>
    </div>
  </section>

  <footer class="p-4 border-t border-zinc-800 text-center text-xs text-zinc-500">
    Hybrid TMA + Go Streaming Architecture
  </footer>
</main>

<style>
  :global(body) {
    margin: 0;
    font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, sans-serif;
  }
</style>
