/* BotMaker AI — backend API klienti.
   Avval real backend endpointlari chaqiriladi; ular mavjud bo'lmasa
   (demo rejim) localStorage'ga asoslangan lokal API ishlaydi. */

import { API_BASE, DEFAULT_CARD, DEMO_CHECKS, DEMO_USERS, num } from './data';
import type {
  AdminTemplate,
  BotItem,
  BotSettings,
  CardSettings,
  PaymentCheck,
  PlatformUser,
  Stats,
  UserProfile,
} from './types';

let apiBase: string | null = null;

export async function ensureApiBase(): Promise<string> {
  if (apiBase) return apiBase;
  const candidates = ['/botmaker/api', '/api', `${API_BASE}/api`];
  for (const base of candidates) {
    for (const probe of ['/health', '/stats', '/bots', '']) {
      try {
        const res = await fetch(`${base}${probe}`, { signal: AbortSignal.timeout(2500) });
        // Server javob bergan bo'lsa (200 yoki 404) — u yetib olinadi
        if (res.ok || res.status === 404) {
          apiBase = base;
          return base;
        }
      } catch {
        /* keyingi manzilni sinab ko'ramiz */
      }
    }
  }
  throw new Error('Backend serverga ulanib bo‘lmadi');
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const base = await ensureApiBase();
  const res = await fetch(`${base}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  });
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    try {
      const errJson = (await res.json()) as Record<string, string>;
      if (errJson.error) msg = errJson.error;
      else if (errJson.message) msg = errJson.message;
      else if (errJson.warning) msg = errJson.warning;
    } catch {
      /* ignore */
    }
    throw new Error(msg);
  }
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

/* ---- localStorage yordamchilari ---- */

const LS_PREFIX = 'botmaker.';

function readLS<T>(key: string, fallback: T): T {
  if (typeof window === 'undefined') return fallback;
  try {
    const raw = window.localStorage.getItem(LS_PREFIX + key);
    if (!raw) return fallback;
    return JSON.parse(raw) as T;
  } catch {
    return fallback;
  }
}

function writeLS<T>(key: string, value: T): void {
  if (typeof window === 'undefined') return;
  try {
    window.localStorage.setItem(LS_PREFIX + key, JSON.stringify(value));
  } catch {
    /* xotira cheklovi — demo ma'lumotlar saqlanmaydi */
  }
}

/* ============================================================
   TO'LOV CHEKLARI
============================================================ */

export async function fetchPaymentChecks(): Promise<PaymentCheck[]> {
  try {
    const raw = await apiFetch<unknown>('/payments/checks');
    const list = Array.isArray(raw) ? raw : [];
    return list.map((c, i) => {
      const r = (c ?? {}) as Record<string, unknown>;
      return {
        id: String(r.id ?? r.check_id ?? `c-${i}`),
        userId: r.user_id != null ? String(r.user_id) : undefined,
        userName: String(r.user_name ?? r.name ?? 'Foydalanuvchi'),
        amount: num(r.amount),
        currency: String(r.currency ?? 'UZS'),
        screenshotUrl: r.screenshot_url ?? r.screenshotUrl ? String(r.screenshot_url ?? r.screenshotUrl) : undefined,
        status: (r.status as PaymentCheck['status']) ?? 'pending',
        createdAt: String(r.created_at ?? r.createdAt ?? new Date().toISOString()),
        note: r.note ? String(r.note) : undefined,
      };
    });
  } catch {
    return readLS<PaymentCheck[]>('checks', DEMO_CHECKS);
  }
}

export async function submitPaymentCheck(input: {
  amount: number;
  currency?: string;
  screenshotUrl?: string;
  userId?: string;
  userName: string;
  note?: string;
}): Promise<PaymentCheck> {
  try {
    return await apiFetch<PaymentCheck>('/payments/checks', {
      method: 'POST',
      body: JSON.stringify(input),
    });
  } catch {
    const check: PaymentCheck = {
      id: `c${Date.now()}`,
      userId: input.userId,
      userName: input.userName,
      amount: input.amount,
      currency: input.currency ?? 'UZS',
      screenshotUrl: input.screenshotUrl,
      status: 'pending',
      createdAt: new Date().toISOString(),
      note: input.note,
    };
    const all = readLS<PaymentCheck[]>('checks', DEMO_CHECKS);
    writeLS('checks', [check, ...all]);
    return check;
  }
}

export async function resolvePaymentCheck(
  id: string,
  status: 'approved' | 'rejected',
): Promise<PaymentCheck | null> {
  try {
    return await apiFetch<PaymentCheck>(`/payments/checks/${id}`, {
      method: 'PATCH',
      body: JSON.stringify({ status }),
    });
  } catch {
    const all = readLS<PaymentCheck[]>('checks', DEMO_CHECKS);
    const idx = all.findIndex((c) => c.id === id);
    if (idx === -1) return null;
    const updated = { ...all[idx], status };
    all[idx] = updated;
    writeLS('checks', all);
    return updated;
  }
}

/* ============================================================
   KARTA SOZLAMALARI
============================================================ */

export async function fetchCardSettings(): Promise<CardSettings> {
  try {
    return await apiFetch<CardSettings>('/settings/card');
  } catch {
    return readLS<CardSettings>('card', DEFAULT_CARD);
  }
}

export async function saveCardSettings(card: CardSettings): Promise<CardSettings> {
  try {
    return await apiFetch<CardSettings>('/settings/card', {
      method: 'PUT',
      body: JSON.stringify(card),
    });
  } catch {
    writeLS('card', card);
    return card;
  }
}

/* ============================================================
   ADMIN: GIT REPO ASOSIDAGI SHABLONLAR
============================================================ */

export async function fetchAdminTemplates(): Promise<AdminTemplate[]> {
  try {
    const raw = await apiFetch<unknown>('/templates/custom');
    const list = Array.isArray(raw) ? raw : [];
    return list.map((t, i) => {
      const r = (t ?? {}) as Record<string, unknown>;
      return {
        id: String(r.id ?? `t-${i}`),
        name: String(r.name ?? 'Nomsiz shablon'),
        repoUrl: String(r.repo_url ?? r.repoUrl ?? ''),
        price: num(r.price),
        category: String(r.category ?? 'Boshqa'),
        description: r.description ? String(r.description) : undefined,
        createdAt: String(r.created_at ?? r.createdAt ?? new Date().toISOString()),
      };
    });
  } catch {
    return readLS<AdminTemplate[]>('adminTemplates', []);
  }
}

export async function addAdminTemplate(
  tpl: Omit<AdminTemplate, 'id' | 'createdAt'>,
): Promise<AdminTemplate> {
  try {
    return await apiFetch<AdminTemplate>('/templates/custom', {
      method: 'POST',
      body: JSON.stringify(tpl),
    });
  } catch {
    const full: AdminTemplate = { ...tpl, id: `t${Date.now()}`, createdAt: new Date().toISOString() };
    const all = readLS<AdminTemplate[]>('adminTemplates', []);
    writeLS('adminTemplates', [full, ...all]);
    return full;
  }
}

export async function deleteAdminTemplate(id: string): Promise<void> {
  try {
    await apiFetch(`/templates/custom/${id}`, { method: 'DELETE' });
  } catch {
    const all = readLS<AdminTemplate[]>('adminTemplates', []);
    writeLS(
      'adminTemplates',
      all.filter((t) => t.id !== id),
    );
  }
}

/* ============================================================
   ADMIN: FOYDALANUVCHILAR
============================================================ */

export async function fetchPlatformUsers(bots: BotItem[]): Promise<PlatformUser[]> {
  try {
    const raw = await apiFetch<unknown>('/users');
    const list = Array.isArray(raw) ? raw : [];
    return list.map((u, i) => {
      const r = (u ?? {}) as Record<string, unknown>;
      return {
        id: String(r.id ?? r.user_id ?? `u-${i}`),
        name: String(r.name ?? r.first_name ?? 'Foydalanuvchi'),
        username: r.username ? String(r.username) : undefined,
        avatar: r.avatar ?? r.photo_url ? String(r.avatar ?? r.photo_url) : undefined,
        balance: num(r.balance),
        botCount: num(r.bot_count ?? r.bots_count),
        role: r.role === 'admin' || r.is_admin ? 'admin' : 'user',
        joinedAt: r.joined_at ? String(r.joined_at) : undefined,
      };
    });
  } catch {
    const users = readLS<PlatformUser[]>('users', DEMO_USERS);
    const session = readLS<UserProfile | null>('session', null);
    const merged = users.map((u) =>
      session && u.id === session.id
        ? {
            ...u,
            name: session.name,
            username: session.username ?? u.username,
            avatar: session.avatar ?? u.avatar,
            balance: session.balance,
            role: session.role,
          }
        : u,
    );
    if (session && !merged.some((u) => u.id === session.id)) {
      merged.push({
        id: session.id,
        name: session.name,
        username: session.username,
        avatar: session.avatar,
        balance: session.balance,
        botCount: 0,
        role: session.role,
        joinedAt: session.joinedAt,
      });
    }
    return merged.map((u) => ({
      ...u,
      botCount: bots.filter((b) => b.adminId === u.id).length || u.botCount,
    }));
  }
}

/* Balans qo'shish/ayirish (lokal demo rejim) */
export function applyLocalUserBalance(userId: string, amount: number): void {
  const users = readLS<PlatformUser[]>('users', DEMO_USERS);
  writeLS(
    'users',
    users.map((u) => (u.id === userId ? { ...u, balance: u.balance + amount } : u)),
  );
  const session = readLS<UserProfile | null>('session', null);
  if (session && session.id === userId) {
    writeLS('session', { ...session, balance: session.balance + amount });
  }
}

/* ============================================================
   BOT SOZLAMALARI  (GET/PUT /api/bots/:id/settings)
============================================================ */

export async function fetchBotSettings(
  id: string,
  fallback: Partial<BotSettings>,
): Promise<BotSettings> {
  try {
    const raw = await apiFetch<Record<string, unknown>>(`/bots/${id}/settings`);
    return {
      adminId: String(raw.admin_id ?? raw.adminId ?? raw.admin ?? ''),
      apiKey: String(raw.api_key ?? raw.apiKey ?? ''),
      channelId: String(raw.channel_id ?? raw.channelId ?? ''),
      currency: String(raw.currency ?? 'UZS'),
      webhookUrl: String(raw.webhook_url ?? raw.webhookUrl ?? ''),
    };
  } catch {
    const local = readLS<BotSettings>(`botsettings.${id}`, {
      adminId: '',
      apiKey: '',
      channelId: '',
      currency: 'UZS',
      webhookUrl: '',
    });
    return {
      adminId: local.adminId || fallback.adminId || '',
      apiKey: local.apiKey || fallback.apiKey || '',
      channelId: local.channelId || fallback.channelId || '',
      currency: local.currency || fallback.currency || 'UZS',
      webhookUrl: local.webhookUrl || fallback.webhookUrl || '',
    };
  }
}

export async function saveBotSettings(
  id: string,
  settings: BotSettings,
): Promise<'api' | 'legacy' | 'local'> {
  const body = {
    admin_id: settings.adminId.trim() || undefined,
    api_key: settings.apiKey.trim() || undefined,
    channel_id: settings.channelId.trim() || undefined,
    currency: settings.currency.trim() || undefined,
    webhook_url: settings.webhookUrl.trim() || undefined,
  };
  const cache = () => writeLS(`botsettings.${id}`, settings);
  try {
    await apiFetch(`/bots/${id}/settings`, { method: 'PUT', body: JSON.stringify(body) });
    cache();
    return 'api';
  } catch {
    try {
      await apiFetch(`/bots/${id}`, { method: 'PUT', body: JSON.stringify(body) });
      cache();
      return 'legacy';
    } catch {
      cache();
      return 'local';
    }
  }
}

/* ============================================================
   CHEK SKRINSHOTINI YUKLASH (Backblaze B2 / lokal)
============================================================ */

export async function uploadScreenshot(
  file: File,
): Promise<{ url: string; storage: 'b2' | 'local' }> {
  const endpoint = process.env.NEXT_PUBLIC_B2_UPLOAD_URL;
  if (endpoint) {
    const form = new FormData();
    form.append('file', file);
    const res = await fetch(endpoint, { method: 'POST', body: form });
    if (!res.ok) throw new Error(`Yuklash xatosi: HTTP ${res.status}`);
    const data = (await res.json()) as Record<string, unknown>;
    const url = String(data.url ?? data.fileUrl ?? data.file_url ?? '');
    if (url) return { url, storage: 'b2' };
  }
  const url = await compressImage(file);
  return { url, storage: 'local' };
}

/* Rasmni siqib data-URL ga aylantirish (lokal saqlash uchun) */
export function compressImage(file: File, maxDim = 900, quality = 0.82): Promise<string> {
  return new Promise((resolve, reject) => {
    const reader = new FileReader();
    reader.onerror = () => reject(new Error('Faylni o‘qib bo‘lmadi'));
    reader.onload = () => {
      const img = new Image();
      img.onerror = () => reject(new Error('Rasm formati noto‘g‘ri'));
      img.onload = () => {
        const scale = Math.min(1, maxDim / Math.max(img.width, img.height));
        const w = Math.max(1, Math.round(img.width * scale));
        const h = Math.max(1, Math.round(img.height * scale));
        const canvas = document.createElement('canvas');
        canvas.width = w;
        canvas.height = h;
        const ctx = canvas.getContext('2d');
        if (!ctx) {
          resolve(reader.result as string);
          return;
        }
        ctx.fillStyle = '#ffffff';
        ctx.fillRect(0, 0, w, h);
        ctx.drawImage(img, 0, 0, w, h);
        resolve(canvas.toDataURL('image/jpeg', quality));
      };
      img.src = reader.result as string;
    };
    reader.readAsDataURL(file);
  });
}

/* ---- Klipbordga nusxalash ---- */

export async function copyText(text: string): Promise<boolean> {
  try {
    await navigator.clipboard.writeText(text);
    return true;
  } catch {
    try {
      const ta = document.createElement('textarea');
      ta.value = text;
      ta.style.position = 'fixed';
      ta.style.opacity = '0';
      document.body.appendChild(ta);
      ta.select();
      const ok = document.execCommand('copy');
      document.body.removeChild(ta);
      return ok;
    } catch {
      return false;
    }
  }
}

