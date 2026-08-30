<script lang="ts">
  import { searchTracks, getTrack } from '../lib/api';
  import { playTrack, queue } from '../stores/player';
  import WebApp from '@twa-dev/sdk';

  let query = '';
  let results: any[] = [];
  let loading = false;
  let error = '';

  async function doSearch() {
    if (!query.trim()) return;
    loading = true;
    error = '';
    try {
      const data = await searchTracks(query.trim());
      results = data.results || [];
    } catch (e: any) {
      error = e.message || 'Search failed';
      results = [];
    } finally {
      loading = false;
    }
  }

  async function play(item: any) {
    try {
      WebApp.HapticFeedback.impactOccurred('medium');
      const meta = await getTrack(item.id);
      const track = {
        id: item.id,
        title: item.title,
        artist: item.artist,
        duration: item.duration,
        thumbnail: item.thumbnail,
        streamUrl: meta.streamUrl,
      };
      queue.update(q => {
        if (!q.find(t => t.id === track.id)) return [...q, track];
        return q;
      });
      await playTrack(track);
    } catch (e: any) {
      error = e.message;
    }
  }

  function onKey(e: KeyboardEvent) {
    if (e.key === 'Enter') doSearch();
  }
</script>

<div class="flex flex-col h-full">
  <div class="p-4">
    <div class="relative">
      <input
        bind:value={query}
        on:keydown={onKey}
        placeholder="Search songs, artists..."
        class="w-full bg-zinc-900 border border-zinc-700 rounded-xl px-4 py-3 text-sm outline-none focus:border-zinc-500"
      />
      <button
        on:click={doSearch}
        class="absolute right-2 top-1/2 -translate-y-1/2 px-3 py-1.5 text-xs bg-white text-black rounded-lg font-medium"
      >
        Search
      </button>
    </div>
  </div>

  {#if loading}
    <div class="flex-1 flex items-center justify-center text-zinc-500 text-sm">Searching...</div>
  {:else if error}
    <div class="p-4 text-red-400 text-sm">{error}</div>
  {:else if results.length === 0 && query}
    <div class="flex-1 flex items-center justify-center text-zinc-500 text-sm">No results</div>
  {:else}
    <div class="flex-1 overflow-y-auto px-2 pb-24">
      {#each results as item}
        <button
          class="w-full flex items-center gap-3 p-3 rounded-xl active:bg-zinc-900 transition text-left"
          on:click={() => play(item)}
        >
          <div class="w-12 h-12 rounded-lg overflow-hidden bg-zinc-800 flex-shrink-0">
            {#if item.thumbnail}
              <img src={item.thumbnail} alt="" class="w-full h-full object-cover" />
            {/if}
          </div>
          <div class="min-w-0 flex-1">
            <div class="text-sm font-medium truncate">{item.title}</div>
            <div class="text-xs text-zinc-400 truncate">{item.artist}</div>
          </div>
          <div class="text-xs text-zinc-600">{Math.floor(item.duration / 60)}:{String(item.duration % 60).padStart(2, '0')}</div>
        </button>
      {/each}
    </div>
  {/if}
</div>
