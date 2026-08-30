import { writable, derived, get } from 'svelte/store';
import { Howl } from 'howler';

export interface Track {
  id: string;
  title: string;
  artist: string;
  duration: number;
  thumbnail?: string;
  streamUrl?: string;
}

export const currentTrack = writable<Track | null>(null);
export const isPlaying = writable(false);
export const position = writable(0);
export const duration = writable(0);
export const volume = writable(0.85);
export const queue = writable<Track[]>([]);

let sound: Howl | null = null;
let rafId: number | null = null;

function tick() {
  if (sound && sound.playing()) {
    position.set(sound.seek() as number);
    rafId = requestAnimationFrame(tick);
  }
}

export async function playTrack(track: Track) {
  if (sound) {
    sound.unload();
    sound = null;
  }

  currentTrack.set(track);
  position.set(0);
  duration.set(track.duration || 0);

  if (!track.streamUrl) {
    console.warn('No streamUrl');
    return;
  }

  sound = new Howl({
    src: [track.streamUrl],
    html5: true, // important for streaming
    volume: get(volume),
    onplay: () => {
      isPlaying.set(true);
      tick();
    },
    onpause: () => {
      isPlaying.set(false);
      if (rafId) cancelAnimationFrame(rafId);
    },
    onend: () => {
      isPlaying.set(false);
      playNext();
    },
    onload: () => {
      duration.set(sound?.duration() || track.duration || 0);
    },
  });

  sound.play();
}

export function togglePlay() {
  if (!sound) return;
  if (sound.playing()) {
    sound.pause();
  } else {
    sound.play();
  }
}

export function seek(sec: number) {
  if (sound) {
    sound.seek(sec);
    position.set(sec);
  }
}

export function setVolume(v: number) {
  volume.set(v);
  if (sound) sound.volume(v);
}

export function playNext() {
  const q = get(queue);
  const cur = get(currentTrack);
  if (!cur || q.length === 0) return;
  const idx = q.findIndex(t => t.id === cur.id);
  const next = q[idx + 1] || q[0];
  if (next) playTrack(next);
}

export function playPrev() {
  const q = get(queue);
  const cur = get(currentTrack);
  if (!cur || q.length === 0) return;
  const idx = q.findIndex(t => t.id === cur.id);
  const prev = q[idx - 1] || q[q.length - 1];
  if (prev) playTrack(prev);
}

export const progress = derived([position, duration], ([$pos, $dur]) =>
  $dur > 0 ? ($pos / $dur) * 100 : 0
);
