import WebApp from '@twa-dev/sdk';

// Empty VITE_API_BASE means same-origin. This lets the unified Railway
// deployment serve both the Mini App and Go API from one domain.
const API_BASE = (import.meta.env.VITE_API_BASE || '').replace(/\/$/, '');

function getInitData(): string {
  try { return WebApp.initData || ''; } catch { return ''; }
}

async function request(path: string, options: RequestInit = {}) {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    'X-Telegram-Init-Data': getInitData(),
    ...(options.headers as Record<string, string> || {}),
  };
  const res = await fetch(`${API_BASE}${path}`, {...options, headers});
  if (!res.ok) {
    const err = await res.json().catch(() => ({}));
    throw new Error(err.error || `HTTP ${res.status}`);
  }
  return res.json();
}

export async function searchTracks(q: string) { return request(`/api/search?q=${encodeURIComponent(q)}`); }
export async function getTrack(id: string, source?: string) {
  const params = source ? `?source=${encodeURIComponent(source)}` : '';
  return request(`/api/track/${encodeURIComponent(id)}${params}`);
}

export async function getCastStatus(): Promise<{ cast_enabled: boolean; reason?: string }> {
  try { const res = await fetch(`${API_BASE}/api/cast/status`); if (!res.ok) return {cast_enabled:false}; return res.json(); }
  catch { return {cast_enabled:false,reason:'unreachable'}; }
}

export function getWsUrl(roomId: string, userId: string) {
  const base = API_BASE ? API_BASE.replace(/^http/, 'ws') : `${location.protocol === 'https:' ? 'wss:' : 'ws:'}//${location.host}`;
  return `${base}/ws?room=${encodeURIComponent(roomId)}&user=${encodeURIComponent(userId)}`;
}

export { API_BASE };
