<script lang="ts">
  import {
    currentTrack,
    isPlaying,
    position,
    duration,
    progress,
    volume,
    togglePlay,
    seek,
    setVolume,
    playNext,
    playPrev,
  } from '../stores/player';
  import WebApp from '@twa-dev/sdk';
  import Waveform from './Waveform.svelte';
  import Lyrics from './Lyrics.svelte';
  import { API_BASE } from '../lib/api';

  let showLyrics = false;
  let casting = false;
  let lrcText = ''; // can be loaded from API later

  function formatTime(sec: number) {
    if (!sec || isNaN(sec)) return '0:00';
    const m = Math.floor(sec / 60);
    const s = Math.floor(sec % 60);
    return `${m}:${s.toString().padStart(2, '0')}`;
  }

  function onSeek(e: Event) {
    const target = e.target as HTMLInputElement;
    const pct = Number(target.value);
    const dur = $duration || 0;
    seek((pct / 100) * dur);
  }

  function haptic(type: 'light' | 'medium' = 'light') {
    try {
      WebApp.HapticFeedback.impactOccurred(type);
    } catch {}
  }

  async function castToVC() {
    if (!$currentTrack) return;
    haptic('medium');
    casting = true;
    try {
      // In real usage, chat_id comes from the group the Mini App was opened from
      // WebApp.initDataUnsafe.chat?.id
      const chatId = (WebApp as any).initDataUnsafe?.chat?.id;
      if (!chatId) {
        alert('Open this Mini App from a group to cast to its Voice Chat');
        return;
      }

      const res = await fetch(`${API_BASE}/api/cast`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'X-Telegram-Init-Data': WebApp.initData || '',
        },
        body: JSON.stringify({
          chat_id: chatId,
          track_id: $currentTrack.id,
          title: $currentTrack.title,
          artist: $currentTrack.artist,
        }),
      });

      if (!res.ok) {
        const err = await res.json().catch(() => ({}));
        throw new Error(err.error || 'Cast failed');
      }

      WebApp.showAlert('Now playing in Voice Chat 🎵');
    } catch (e: any) {
      WebApp.showAlert(e.message || 'Cast failed');
    } finally {
      casting = false;
    }
  }
</script>

{#if $currentTrack}
  <div class="fixed inset-0 z-50 flex flex-col bg-black text-white">
    <!-- Background blur -->
    <div
      class="absolute inset-0 opacity-25 blur-3xl scale-110"
      style="background-image: url({$currentTrack.thumbnail || ''}); background-size: cover; background-position: center;"
    ></div>

    <div class="relative z-10 flex flex-col h-full p-5">
      <!-- Top bar -->
      <div class="flex items-center justify-between mb-4">
        <button class="text-zinc-400 text-lg px-2" on:click={() => history.back()}>↓</button>
        <span class="text-xs tracking-widest uppercase text-zinc-400">Now Playing</span>
        <button
          class="text-zinc-400 text-xs px-2"
          on:click={() => (showLyrics = !showLyrics)}
        >
          {showLyrics ? 'Cover' : 'Lyrics'}
        </button>
      </div>

      {#if showLyrics}
        <div class="flex-1 flex flex-col justify-center">
          <Lyrics lrc={lrcText} />
        </div>
      {:else}
        <!-- Album Art -->
        <div class="flex-1 flex items-center justify-center">
          <div
            class="w-64 h-64 sm:w-72 sm:h-72 rounded-2xl shadow-2xl overflow-hidden bg-zinc-900 transition-transform duration-300"
            class:scale-95={!$isPlaying}
          >
            {#if $currentTrack.thumbnail}
              <img src={$currentTrack.thumbnail} alt="cover" class="w-full h-full object-cover" />
            {:else}
              <div class="w-full h-full flex items-center justify-center text-zinc-600">No Art</div>
            {/if}
          </div>
        </div>

        <!-- Waveform -->
        <div class="my-3">
          <Waveform />
        </div>
      {/if}

      <!-- Meta -->
      <div class="text-center mt-2">
        <h2 class="text-xl font-bold truncate">{$currentTrack.title}</h2>
        <p class="text-zinc-400 text-sm mt-1 truncate">{$currentTrack.artist}</p>
      </div>

      <!-- Progress -->
      <div class="mt-5">
        <input
          type="range"
          min="0"
          max="100"
          step="0.1"
          value={$progress}
          on:input={onSeek}
          class="w-full h-1 appearance-none bg-zinc-700 rounded-full outline-none
                 [&::-webkit-slider-thumb]:appearance-none
                 [&::-webkit-slider-thumb]:w-3.5
                 [&::-webkit-slider-thumb]:h-3.5
                 [&::-webkit-slider-thumb]:rounded-full
                 [&::-webkit-slider-thumb]:bg-white"
        />
        <div class="flex justify-between text-xs text-zinc-500 mt-1.5">
          <span>{formatTime($position)}</span>
          <span>{formatTime($duration)}</span>
        </div>
      </div>

      <!-- Controls -->
      <div class="flex items-center justify-center gap-10 mt-6">
        <button class="text-2xl text-zinc-300 active:scale-90 transition" on:click={() => { haptic(); playPrev(); }}>⏮</button>
        <button
          class="w-16 h-16 rounded-full bg-white text-black flex items-center justify-center text-2xl active:scale-95 transition shadow-lg"
          on:click={() => { haptic('medium'); togglePlay(); }}
        >
          {$isPlaying ? '⏸' : '▶'}
        </button>
        <button class="text-2xl text-zinc-300 active:scale-90 transition" on:click={() => { haptic(); playNext(); }}>⏭</button>
      </div>

      <!-- Volume -->
      <div class="flex items-center gap-3 mt-5">
        <span class="text-xs text-zinc-500">🔈</span>
        <input
          type="range" min="0" max="1" step="0.01" value={$volume}
          on:input={(e) => setVolume(Number(e.currentTarget.value))}
          class="flex-1 h-1 appearance-none bg-zinc-700 rounded-full"
        />
        <span class="text-xs text-zinc-500">🔊</span>
      </div>

      <!-- Cast -->
      <button
        class="mt-5 w-full py-3.5 rounded-xl bg-zinc-900 border border-zinc-700 text-sm font-medium active:scale-[0.98] transition disabled:opacity-50"
        disabled={casting}
        on:click={castToVC}
      >
        {casting ? 'Casting…' : 'Cast to Group Voice Chat'}
      </button>
    </div>
  </div>
{/if}
