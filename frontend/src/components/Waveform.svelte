<script lang="ts">
  import { onMount, onDestroy } from 'svelte';
  import { isPlaying } from '../stores/player';

  export let audioElement: HTMLAudioElement | null = null;

  let canvas: HTMLCanvasElement;
  let ctx: CanvasRenderingContext2D | null = null;
  let analyser: AnalyserNode | null = null;
  let dataArray: Uint8Array | null = null;
  let raf: number;
  let audioCtx: AudioContext | null = null;
  let source: MediaElementAudioSourceNode | null = null;

  onMount(() => {
    ctx = canvas.getContext('2d');
    resize();
    window.addEventListener('resize', resize);
  });

  onDestroy(() => {
    cancelAnimationFrame(raf);
    window.removeEventListener('resize', resize);
    if (audioCtx) audioCtx.close();
  });

  function resize() {
    if (!canvas) return;
    const dpr = window.devicePixelRatio || 1;
    canvas.width = canvas.clientWidth * dpr;
    canvas.height = canvas.clientHeight * dpr;
    if (ctx) ctx.scale(dpr, dpr);
  }

  $: if ($isPlaying && audioElement && !analyser) {
    setupAnalyser();
  }

  function setupAnalyser() {
    if (!audioElement || analyser) return;
    try {
      audioCtx = new (window.AudioContext || (window as any).webkitAudioContext)();
      source = audioCtx.createMediaElementSource(audioElement);
      analyser = audioCtx.createAnalyser();
      analyser.fftSize = 256;
      source.connect(analyser);
      analyser.connect(audioCtx.destination);
      dataArray = new Uint8Array(analyser.frequencyBinCount);
      draw();
    } catch (e) {
      console.warn('Waveform setup failed', e);
    }
  }

  function draw() {
    if (!ctx || !analyser || !dataArray || !canvas) return;
    raf = requestAnimationFrame(draw);

    analyser.getByteFrequencyData(dataArray);

    const w = canvas.clientWidth;
    const h = canvas.clientHeight;
    ctx.clearRect(0, 0, w, h);

    const bars = 48;
    const gap = 2;
    const barWidth = (w - gap * bars) / bars;

    for (let i = 0; i < bars; i++) {
      const value = dataArray[i % dataArray.length] / 255;
      const barHeight = Math.max(4, value * h * 0.85);
      const x = i * (barWidth + gap);
      const y = (h - barHeight) / 2;

      const gradient = ctx.createLinearGradient(0, y, 0, y + barHeight);
      gradient.addColorStop(0, '#ffffff');
      gradient.addColorStop(1, '#a1a1aa');
      ctx.fillStyle = gradient;
      ctx.beginPath();
      ctx.roundRect(x, y, barWidth, barHeight, 2);
      ctx.fill();
    }
  }
</script>

<canvas
  bind:this={canvas}
  class="w-full h-16 opacity-80"
  style="display: block;"
></canvas>
