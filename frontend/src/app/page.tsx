'use client';

import { useCallback, useEffect, useMemo, useState } from 'react';
import type { LucideIcon } from 'lucide-react';
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Bot as BotIcon,
  Check,
  CheckCircle2,
  ChevronLeft,
  ChevronRight,
  Clock,
  Gauge,
  Globe,
  KeyRound,
  Link2,
  Loader2,
  Megaphone,
  MessageSquare,
  Plus,
  RefreshCw,
  Rocket,
  Send,
  Server,
  Shield,
  Sparkles,
  Trash2,
  User,
  Webhook,
  X,
  Zap,
} from 'lucide-react';

/* ============================================================
   TURLAR VA DOIMIY MA'LUMOTLAR
============================================================ */

type TemplateId = 'AniTez' | 'AnixUltra';

interface BotItem {
  id: string;
  name: string;
  username?: string;
  template: TemplateId;
  token?: string;
  adminId?: string;
  webhookUrl?: string;
  running: boolean;
  webhookActive: boolean;
  latency?: number;
  lastActivity?: string;
}

interface Stats {
  totalBots: number;
  activeWebhooks: number;
  processedRequests: number;
}

interface ToastMsg {
  id: number;
  type: 'success' | 'error' | 'info';
  message: string;
}

const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? 'http://localhost:8085';
const WEBHOOK_HOST = 'nokori-uz.duckdns.org';

const TEMPLATES: {
  id: TemplateId;
  name: string;
  tagline: string;
  gradient: string;
  chip: string;
  glyph: 'zap' | 'rocket';
  features: string[];
}[] = [
  {
    id: 'AniTez',
    name: 'AniTez',
    tagline: 'Tezkor javob beruvchi media va obuna boshqaruv boti',
    gradient: 'from-amber-400 via-orange-500 to-rose-500',
    chip: 'border-amber-400/20 bg-amber-500/10 text-amber-300',
    glyph: 'zap',
    features: [
      '~2ms webhook javob vaqti',
      'Avtomatik media qayta ishlash',
      'Obuna (subscription) tizimi',
      'Inline rejim va tezkor qidiruv',
    ],
  },
  {
    id: 'AnixUltra',
    name: 'AnixUltra',
    tagline: 'Kengaytirilgan qidiruv, statistika va ko‘p tarmoqli bot',
    gradient: 'from-violet-500 via-fuchsia-500 to-cyan-400',
    chip: 'border-fuchsia-400/20 bg-fuchsia-500/10 text-fuchsia-300',
    glyph: 'rocket',
    features: [
      'Kuchli aniq qidiruv (fuzzy match)',
      'Real-time statistika paneli',
      'Ko‘p tilli UI (uz / ru / en)',
      'Batch broadcast va segmentatsiya',
    ],
  },
];

const DEMO_BOTS: BotItem[] = [
  {
    id: 'b1',
    name: 'AniTez Poster',
    username: 'anitez_poster_bot',
    template: 'AniTez',
    running: true,
    webhookActive: true,
    latency: 12,
    lastActivity: '2 daqiqa oldin',
  },
  {
    id: 'b2',
    name: 'AnixUltra Finder',
    username: 'anixultra_finder_bot',
    template: 'AnixUltra',
    running: true,
    webhookActive: true,
    latency: 4,
    lastActivity: '1 soat oldin',
  },
  {
    id: 'b3',
    name: 'Anime Weekend',
    username: 'anime_weekend_bot',
    template: 'AniTez',
    running: false,
    webhookActive: false,
    lastActivity: '3 kun oldin',
  },
];

const DEMO_STATS: Stats = { totalBots: 3, activeWebhooks: 2, processedRequests: 12458 };

/* ============================================================
   BACKEND API INTEGRATSIYASI
   Avval Next.js proxisi (/api), bo‘lmasa to‘g‘ridan-to‘g‘ri
   http://localhost:8085/api ga urinadi.
============================================================ */

let apiBase: string | null = null;

async function ensureApiBase(): Promise<string> {
  if (apiBase) return apiBase;
  const candidates = ['/api', `${API_BASE}/api`];
  for (const base of candidates) {
    for (const probe of ['/health', '/stats', '/bots', '']) {
      try {
        const res = await fetch(`${base}${probe}`, {
          signal: AbortSignal.timeout(2500),
        });
        // Server javob bergan bo‘lsa (200 yoki 404) — u yetib olinadi
        if (res.ok || res.status === 404) {
          apiBase = base;
          return base;
        }
      } catch {
        /* keyingi manzilni sinab ko‘ramiz */
      }
    }
  }
  throw new Error('Backend serverga ulanib bo‘lmadi');
}

async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const base = await ensureApiBase();
  const res = await fetch(`${base}${path}`, {
    headers: { 'Content-Type': 'application/json' },
    ...init,
  });
  if (!res.ok) throw new Error(`HTTP ${res.status}`);
  if (res.status === 204) return undefined as T;
  return (await res.json()) as T;
}

/* ---- Normalizatsiya: backend javoblarini yagona shaklga keltirish ---- */

function num(v: unknown): number {
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

function normalizeStats(raw: unknown): Stats {
  const r = (raw ?? {}) as Record<string, unknown>;
  return {
    totalBots: num(r.total_bots ?? r.totalBots ?? r.bots),
    activeWebhooks: num(r.active_webhooks ?? r.activeWebhooks ?? r.webhooks),
    processedRequests: num(r.processed_requests ?? r.processedRequests ?? r.requests),
  };
}

function normalizeBots(raw: unknown): BotItem[] {
  const list = Array.isArray(raw) ? raw : ((raw as Record<string, unknown>)?.bots as unknown[] | undefined) ?? [];
  return list.map((b, i) => {
    const row = (b ?? {}) as Record<string, unknown>;
    const webhook = row.webhook as Record<string, unknown> | undefined;
    const templateRaw = String(row.template ?? 'AniTez').toLowerCase();
    const adminRaw = row.admin_id ?? row.adminId;
    const urlRaw = row.webhook_url ?? row.webhookUrl;
    const lastRaw = row.last_activity ?? row.updated_at;
    const latencyRaw = row.latency ?? webhook?.latency_ms;
    return {
      id: String(row.id ?? row.bot_id ?? row.botId ?? `bot-${i}`),
      name: String(row.name ?? row.bot_name ?? 'Nomsiz bot'),
      username: row.username ? String(row.username) : undefined,
      template: templateRaw.includes('anix') ? 'AnixUltra' : 'AniTez',
      token: row.token ? String(row.token) : undefined,
      adminId: adminRaw ? String(adminRaw) : undefined,
      webhookUrl: urlRaw ? String(urlRaw) : undefined,
      running: Boolean(row.running ?? row.is_running ?? row.active),
      webhookActive: Boolean(row.webhook_active ?? row.webhookActive ?? webhook?.active ?? true),
      latency: latencyRaw != null ? num(latencyRaw) : undefined,
      lastActivity: lastRaw ? String(lastRaw) : undefined,
    };
  });
}

/* ---- Yordamchi funksiyalar ---- */

function slugify(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'bot';
}

function tplOf(id: TemplateId) {
  return TEMPLATES.find((t) => t.id === id) ?? TEMPLATES[0];
}

function fmtNum(n: number): string {
  return n.toLocaleString('en-US');
}

const INITIAL_FORM = {
  template: 'AniTez' as TemplateId,
  name: '',
  token: '',
  adminId: '',
  webhookUrl: '',
  useWebhook: true,
};

/* ============================================================
   KICHIK UI KOMPONENTLAR
============================================================ */

const inputCls =
  'w-full rounded-xl border border-white/10 bg-white/[0.04] px-3.5 py-2.5 text-sm text-white placeholder:text-slate-500 outline-none transition focus:border-cyan-400/50 focus:bg-white/[0.06] focus:ring-2 focus:ring-cyan-400/20';

function Field({
  label,
  icon: Icon,
  hint,
  children,
}: {
  label: string;
  icon?: LucideIcon;
  hint?: string;
  children: React.ReactNode;
}) {
  return (
    <label className="block">
      <span className="mb-1.5 flex items-center gap-1.5 text-xs font-medium text-slate-400">
        {Icon && <Icon className="h-3.5 w-3.5 text-cyan-400/70" />}
        {label}
      </span>
      {children}
      {hint && <span className="mt-1.5 block text-[11px] leading-relaxed text-slate-500">{hint}</span>}
    </label>
  );
}

function Modal({
  open,
  onClose,
  title,
  subtitle,
  children,
  wide = false,
}: {
  open: boolean;
  onClose: () => void;
  title: string;
  subtitle?: string;
  children: React.ReactNode;
  wide?: boolean;
}) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={onClose} />
      <div
        className={`animate-modal-in relative w-full ${wide ? 'max-w-2xl' : 'max-w-lg'} rounded-2xl border border-white/10 bg-[#0b0b1a]/95 p-6 shadow-2xl shadow-violet-950/50 backdrop-blur-2xl`}
      >
        <div className="mb-5 flex items-start justify-between gap-4">
          <div>
            <h3 className="text-lg font-bold text-white">{title}</h3>
            {subtitle && <p className="mt-0.5 text-sm text-slate-400">{subtitle}</p>}
          </div>
          <button
            onClick={onClose}
            className="flex h-8 w-8 shrink-0 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-400 transition hover:bg-white/10 hover:text-white"
            aria-label="Yopish"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
        {children}
      </div>
    </div>
  );
}

function Toggle({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      onClick={() => onChange(!checked)}
      className={`relative h-6 w-11 shrink-0 rounded-full transition-colors duration-200 ${
        checked ? 'bg-gradient-to-r from-emerald-400 to-teal-500 shadow-glow' : 'bg-white/10'
      }`}
    >
      <span
        className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all duration-200 ${
          checked ? 'left-[22px]' : 'left-0.5'
        }`}
      />
    </button>
  );
}

function TemplateGlyph({ tpl, className }: { tpl: TemplateId; className?: string }) {
  const t = tplOf(tpl);
  return t.glyph === 'zap' ? (
    <Zap className={className} />
  ) : (
    <Rocket className={className} />
  );
}

/* ============================================================
   ASOSIY DASHBOARD
============================================================ */

export default function DashboardPage() {
  const [bots, setBots] = useState<BotItem[]>([]);
  const [stats, setStats] = useState<Stats>({ totalBots: 0, activeWebhooks: 0, processedRequests: 0 });
  const [loading, setLoading] = useState(true);
  const [demoMode, setDemoMode] = useState(false);
  const [busy, setBusy] = useState(false);
  const [busyId, setBusyId] = useState<string | null>(null);
  const [toasts, setToasts] = useState<ToastMsg[]>([]);

  // Yangi bot wizard
  const [createOpen, setCreateOpen] = useState(false);
  const [wizardStep, setWizardStep] = useState(0);
  const [form, setForm] = useState(INITIAL_FORM);

  // Broadcast
  const [broadcastOpen, setBroadcastOpen] = useState(false);
  const [broadcastTarget, setBroadcastTarget] = useState<string>('all');
  const [broadcastText, setBroadcastText] = useState('');

  // O‘chirish tasdiqlash
  const [deleteTarget, setDeleteTarget] = useState<BotItem | null>(null);

  const stepTitles = ['Shablon tanlang', 'Bot konfiguratsiyasi', 'Webhook ulash'];

  /* ---- Toast ---- */

  const pushToast = useCallback((type: ToastMsg['type'], message: string) => {
    const id = Date.now() + Math.random();
    setToasts((prev) => [...prev, { id, type, message }]);
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4200);
  }, []);

  /* ---- Ma'lumot yuklash ---- */

  const loadData = useCallback(async (): Promise<boolean> => {
    setLoading(true);
    try {
      const [statsRes, botsRes] = await Promise.all([
        apiFetch<unknown>('/stats'),
        apiFetch<unknown>('/bots'),
      ]);
      setStats(normalizeStats(statsRes));
      setBots(normalizeBots(botsRes));
      setDemoMode(false);
      return true;
    } catch {
      // Backend mavjud emas — demo rejim
      setBots(DEMO_BOTS);
      setStats(DEMO_STATS);
      setDemoMode(true);
      return false;
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  const webhookLatency = useMemo(() => {
    const actives = bots.filter((b) => b.webhookActive && b.running);
    if (actives.length === 0) return 2;
    return Math.min(...actives.map((b) => b.latency ?? 2));
  }, [bots]);

  const statCards: { label: string; value: string; icon: LucideIcon; gradient: string; sub: string }[] = [
    {
      label: 'Jami botlar',
      value: fmtNum(stats.totalBots),
      icon: BotIcon,
      gradient: 'from-cyan-400 to-blue-500',
      sub: 'platformada ro‘yxatdan o‘tgan',
    },
    {
      label: 'Faol Webhooklar',
      value: fmtNum(stats.activeWebhooks),
      icon: Webhook,
      gradient: 'from-violet-400 to-fuchsia-500',
      sub: 'so‘rov qabul qilmoqda',
    },
    {
      label: 'Qayta ishlangan so‘rovlar',
      value: fmtNum(stats.processedRequests),
      icon: Activity,
      gradient: 'from-emerald-400 to-teal-500',
      sub: '24 soat ichida',
    },
  ];

  /* ---- Yangi bot wizard ---- */

  const openCreate = (template?: TemplateId) => {
    setForm({ ...INITIAL_FORM, template: template ?? 'AniTez' });
    setWizardStep(0);
    setCreateOpen(true);
  };

  const closeCreate = () => {
    setCreateOpen(false);
    setWizardStep(0);
    setForm(INITIAL_FORM);
  };

  const autoWebhookUrl = useMemo(
    () => `https://${WEBHOOK_HOST}/webhook/${slugify(form.name || 'bot')}`,
    [form.name],
  );

  const handleCreate = async () => {
    if (!form.name.trim() || !form.token.trim()) {
      pushToast('error', 'Bot nomi va token kiritilishi shart');
      return;
    }
    setBusy(true);
    try {
      await apiFetch('/bots', {
        method: 'POST',
        body: JSON.stringify({
          name: form.name.trim(),
          template: form.template,
          token: form.token.trim(),
          admin_id: form.adminId.trim() || undefined,
          webhook_url: form.useWebhook ? form.webhookUrl.trim() || autoWebhookUrl : undefined,
        }),
      });
      pushToast('success', `«${form.name}» boti muvaffaqiyatli yaratildi`);
      closeCreate();
      await loadData();
    } catch (e) {
      pushToast('error', `Bot yaratilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  /* ---- Bot boshqaruvi ---- */

  const toggleBot = async (bot: BotItem) => {
    const next = !bot.running;
    setBusyId(bot.id);
    // Optimistik yangilash
    setBots((prev) => prev.map((b) => (b.id === bot.id ? { ...b, running: next } : b)));
    try {
      await apiFetch(`/bots/${bot.id}/${next ? 'start' : 'stop'}`, { method: 'POST' });
      pushToast('success', `«${bot.name}» ${next ? 'ishga tushirildi' : 'to‘xtatildi'}`);
    } catch (e) {
      setBots((prev) => prev.map((b) => (b.id === bot.id ? { ...b, running: !next } : b)));
      pushToast('error', `Amal bajarilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusyId(null);
    }
  };

  const confirmDelete = async () => {
    if (!deleteTarget) return;
    setBusy(true);
    try {
      await apiFetch(`/bots/${deleteTarget.id}`, { method: 'DELETE' });
      setBots((prev) => prev.filter((b) => b.id !== deleteTarget.id));
      setStats((s) => ({ ...s, totalBots: Math.max(0, s.totalBots - 1) }));
      pushToast('success', `«${deleteTarget.name}» o‘chirildi`);
      setDeleteTarget(null);
    } catch (e) {
      pushToast('error', `O‘chirishda xatolik: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  const sendBroadcast = async () => {
    if (!broadcastText.trim()) return;
    setBusy(true);
    try {
      await apiFetch('/broadcast', {
        method: 'POST',
        body: JSON.stringify({
          bot_id: broadcastTarget === 'all' ? undefined : broadcastTarget,
          target: broadcastTarget,
          message: broadcastText,
        }),
      });
      pushToast(
        'success',
        broadcastTarget === 'all' ? 'Broadcast barcha botlarga yuborildi' : 'Broadcast muvaffaqiyatli yuborildi',
      );
      setBroadcastOpen(false);
      setBroadcastText('');
    } catch (e) {
      pushToast('error', `Broadcast yuborilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  /* ============================================================
     JSX
  ============================================================ */

  return (
    <div className="relative min-h-screen overflow-x-hidden text-slate-200">
      {/* Fon effektlari */}
      <div aria-hidden className="pointer-events-none fixed inset-0 z-0">
        <div className="bg-grid absolute inset-0 opacity-40" />
        <div className="absolute -left-40 -top-40 h-[500px] w-[500px] rounded-full bg-violet-600/20 blur-[140px]" />
        <div className="absolute -right-40 top-1/3 h-[450px] w-[450px] rounded-full bg-cyan-500/15 blur-[140px]" />
        <div className="absolute -bottom-40 left-1/3 h-[500px] w-[500px] rounded-full bg-fuchsia-600/15 blur-[160px]" />
      </div>

      <div className="relative z-10 mx-auto flex min-h-screen max-w-7xl flex-col px-4 sm:px-6 lg:px-8">
        {/* ===== HEADER ===== */}
        <header className="sticky top-0 z-40 -mx-4 border-b border-white/5 bg-[#05050f]/70 px-4 backdrop-blur-xl sm:-mx-6 sm:px-6 lg:-mx-8 lg:px-8">
          <div className="flex h-16 items-center justify-between gap-3">
            <div className="flex items-center gap-3">
              <div className="relative flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br from-cyan-400 via-violet-500 to-fuchsia-500 shadow-lg shadow-violet-500/30">
                <BotIcon className="h-5 w-5 text-white" />
                <span className="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full bg-emerald-400 ring-2 ring-[#05050f]" />
              </div>
              <div>
                <h1 className="text-lg font-bold leading-tight tracking-tight text-white">
                  BotMaker <span className="text-gradient">AI</span>
                </h1>
                <p className="text-[11px] leading-tight text-slate-500">Telegram bot platformasi</p>
              </div>
            </div>

            <div className="flex items-center gap-2 sm:gap-3">
              {/* VPS Server holati */}
              <div className="hidden items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5 backdrop-blur md:flex">
                <span className="relative flex h-2 w-2">
                  <span
                    className={`absolute inline-flex h-full w-full animate-ping rounded-full opacity-60 ${
                      demoMode ? 'bg-amber-400' : 'bg-emerald-400'
                    }`}
                  />
                  <span
                    className={`relative inline-flex h-2 w-2 rounded-full ${demoMode ? 'bg-amber-400' : 'bg-emerald-400'}`}
                  />
                </span>
                <Server className="h-3.5 w-3.5 text-slate-400" />
                <span className="font-mono text-xs text-slate-300">nokori-uz.duckdns.org</span>
                <span
                  className={`rounded-md px-1.5 py-0.5 text-[10px] font-semibold ${
                    demoMode
                      ? 'border border-amber-400/20 bg-amber-500/10 text-amber-300'
                      : 'border border-emerald-400/20 bg-emerald-500/10 text-emerald-300'
                  }`}
                >
                  {demoMode ? 'OFFLINE' : 'ONLINE'}
                </span>
              </div>

              {/* Webhook status */}
              <div className="flex items-center gap-1.5 rounded-full border border-emerald-400/20 bg-emerald-500/10 px-3 py-1.5">
                <Gauge className="h-3.5 w-3.5 text-emerald-300" />
                <span className="text-xs font-medium text-emerald-300">
                  Webhook: <span className="font-semibold">Active {webhookLatency}ms</span>
                </span>
              </div>

              <button
                onClick={() => void loadData()}
                disabled={loading}
                className="flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 bg-white/5 text-slate-300 transition hover:bg-white/10 hover:text-white disabled:opacity-50"
                title="Ma'lumotlarni yangilash"
              >
                <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
              </button>

              <button
                onClick={() => openCreate()}
                className="hidden items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-3.5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 active:scale-[0.98] sm:flex"
              >
                <Plus className="h-4 w-4" />
                Yangi Bot
              </button>
            </div>
          </div>
        </header>

        <main className="flex-1 space-y-8 pb-16 pt-8">
          {/* Demo rejim banneri */}
          {demoMode && (
            <div className="flex flex-wrap items-center gap-3 rounded-2xl border border-amber-400/20 bg-amber-400/5 px-4 py-3 text-sm text-amber-200/90">
              <AlertTriangle className="h-4 w-4 shrink-0 text-amber-400" />
              <span>
                Backend API topilmadi ({API_BASE}). Hozircha <b className="text-amber-100">demo ma’lumotlar</b>{' '}
                ko‘rsatilmoqda — server ishga tushganda «Yangilash» tugmasini bosing.
              </span>
              <button
                onClick={() => void loadData()}
                className="ml-auto rounded-lg border border-amber-400/30 bg-amber-400/10 px-3 py-1.5 text-xs font-semibold text-amber-200 transition hover:bg-amber-400/20"
              >
                Qayta urinish
              </button>
            </div>
          )}

          {/* ===== STATISTIKA KARTALARI ===== */}
          <section className="grid gap-4 sm:grid-cols-3">
            {statCards.map((c) => (
              <div
                key={c.label}
                className="group relative overflow-hidden rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-xl transition-all hover:border-white/20 hover:bg-white/[0.06]"
              >
                <div
                  className={`absolute -right-8 -top-8 h-28 w-28 rounded-full bg-gradient-to-br ${c.gradient} opacity-20 blur-2xl transition-opacity group-hover:opacity-35`}
                />
                <div className="flex items-center justify-between">
                  <p className="text-sm text-slate-400">{c.label}</p>
                  <div
                    className={`flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${c.gradient} shadow-lg`}
                  >
                    <c.icon className="h-5 w-5 text-white" />
                  </div>
                </div>
                <p className="mt-3 text-3xl font-bold tabular-nums tracking-tight text-white">
                  {loading ? (
                    <span className="inline-block h-8 w-16 animate-pulse rounded-lg bg-white/10 align-middle" />
                  ) : (
                    c.value
                  )}
                </p>
                <p className="mt-1 text-xs text-slate-500">{c.sub}</p>
              </div>
            ))}
          </section>

          {/* ===== MENING BOTLARIM ===== */}
          <section className="space-y-4">
            <div className="flex flex-wrap items-center justify-between gap-3">
              <div className="flex items-center gap-2.5">
                <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-violet-400/20 bg-violet-500/10 text-violet-300">
                  <BotIcon className="h-4 w-4" />
                </div>
                <h2 className="text-xl font-bold text-white">Mening Botlarim</h2>
                <span className="rounded-full border border-white/10 bg-white/5 px-2.5 py-0.5 text-xs text-slate-400">
                  {bots.length} ta
                </span>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => {
                    setBroadcastTarget('all');
                    setBroadcastOpen(true);
                  }}
                  className="flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/5 px-3.5 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10 hover:text-white"
                >
                  <Megaphone className="h-4 w-4 text-cyan-300" />
                  Global Broadcast
                </button>
                <button
                  onClick={() => openCreate()}
                  className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-3.5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 active:scale-[0.98] sm:hidden"
                >
                  <Plus className="h-4 w-4" />
                  Yangi Bot
                </button>
              </div>
            </div>

            {loading ? (
              <div className="space-y-3">
                {[0, 1, 2].map((i) => (
                  <div
                    key={i}
                    className="flex animate-pulse items-center gap-4 rounded-2xl border border-white/5 bg-white/[0.03] p-4"
                  >
                    <div className="h-12 w-12 rounded-xl bg-white/10" />
                    <div className="flex-1 space-y-2">
                      <div className="h-4 w-40 rounded bg-white/10" />
                      <div className="h-3 w-56 rounded bg-white/5" />
                    </div>
                    <div className="h-8 w-24 rounded-full bg-white/5" />
                  </div>
                ))}
              </div>
            ) : bots.length === 0 ? (
              <div className="flex flex-col items-center gap-3 rounded-2xl border border-dashed border-white/15 bg-white/[0.02] px-6 py-14 text-center">
                <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-white/10 bg-white/5 text-slate-400">
                  <BotIcon className="h-7 w-7" />
                </div>
                <p className="font-semibold text-white">Hali botlar yo‘q</p>
                <p className="max-w-sm text-sm text-slate-500">
                  Birinchi Telegram botingizni yarating — «Yangi Bot» tugmasini bosing va wizardni yakunlang.
                </p>
                <button
                  onClick={() => openCreate()}
                  className="mt-2 flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110"
                >
                  <Plus className="h-4 w-4" />
                  Bot yaratish
                </button>
              </div>
            ) : (
              <div className="space-y-3">
                {bots.map((bot) => {
                  const tpl = tplOf(bot.template);
                  const isToggling = busyId === bot.id;
                  return (
                    <div
                      key={bot.id}
                      className="group flex flex-col gap-4 rounded-2xl border border-white/10 bg-white/[0.04] p-4 backdrop-blur-xl transition-all hover:border-white/20 hover:bg-white/[0.06] sm:flex-row sm:items-center"
                    >
                      {/* Bot ma'lumotlari */}
                      <div className="flex min-w-0 flex-1 items-center gap-4">
                        <div
                          className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${tpl.gradient} shadow-lg`}
                        >
                          <TemplateGlyph tpl={bot.template} className="h-6 w-6 text-white" />
                        </div>
                        <div className="min-w-0">
                          <div className="flex items-center gap-2">
                            <p className="truncate font-semibold text-white">{bot.name}</p>
                            {bot.running ? (
                              <span className="rounded-md border border-emerald-400/20 bg-emerald-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-emerald-300">
                                RUN
                              </span>
                            ) : (
                              <span className="rounded-md border border-slate-500/20 bg-slate-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-slate-400">
                                STOP
                              </span>
                            )}
                          </div>
                          <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-500">
                            <span className="font-mono">@{bot.username ?? '—'}</span>
                            <span className={`rounded-full border px-2 py-0.5 font-medium ${tpl.chip}`}>
                              {bot.template}
                            </span>
                            <span className="flex items-center gap-1">
                              <Clock className="h-3 w-3" />
                              {bot.lastActivity ?? 'hozir'}
                            </span>
                          </div>
                        </div>
                      </div>

                      {/* Amallar */}
                      <div className="flex items-center gap-3 pl-16 sm:pl-0">
                        <span
                          className={`flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium ${
                            bot.webhookActive
                              ? 'border-emerald-400/20 bg-emerald-500/10 text-emerald-300'
                              : 'border-slate-500/20 bg-slate-500/10 text-slate-400'
                          }`}
                        >
                          <span
                            className={`h-1.5 w-1.5 rounded-full ${
                              bot.webhookActive ? 'animate-pulse-soft bg-emerald-400' : 'bg-slate-500'
                            }`}
                          />
                          {bot.webhookActive ? `Active ${bot.latency ?? 2}ms` : 'Inactive'}
                        </span>

                        <div
                          className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-300 transition hover:border-white/20 hover:bg-white/10 hover:text-white"
                          title={bot.running ? 'To‘xtatish' : 'Ishga tushirish'}
                        >
                          {isToggling ? (
                            <Loader2 className="h-4 w-4 animate-spin" />
                          ) : (
                            <Toggle checked={bot.running} onChange={() => void toggleBot(bot)} />
                          )}
                        </div>

                        <button
                          onClick={() => {
                            setBroadcastTarget(bot.id);
                            setBroadcastOpen(true);
                          }}
                          className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-300 transition hover:border-cyan-400/30 hover:bg-cyan-400/10 hover:text-cyan-300"
                          title="Broadcast yuborish"
                        >
                          <Megaphone className="h-4 w-4" />
                        </button>

                        <button
                          onClick={() => setDeleteTarget(bot)}
                          className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-300 transition hover:border-rose-400/30 hover:bg-rose-400/10 hover:text-rose-400"
                          title="Botni o‘chirish"
                        >
                          <Trash2 className="h-4 w-4" />
                        </button>
                      </div>
                    </div>
                  );
                })}
              </div>
            )}
          </section>

          {/* ===== SHABLONLAR KATALOGI ===== */}
          <section className="space-y-4">
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-cyan-400/20 bg-cyan-500/10 text-cyan-300">
                <Sparkles className="h-4 w-4" />
              </div>
              <h2 className="text-xl font-bold text-white">Shablonlar Katalogi</h2>
            </div>

            <div className="grid gap-5 md:grid-cols-2">
              {TEMPLATES.map((t) => (
                <div
                  key={t.id}
                  className="group relative overflow-hidden rounded-2xl border border-white/10 bg-white/[0.04] p-6 backdrop-blur-xl transition-all hover:border-white/20"
                >
                  <div className={`absolute inset-x-0 top-0 h-px bg-gradient-to-r ${t.gradient} opacity-70`} />
                  <div
                    className={`absolute -right-16 -top-16 h-48 w-48 rounded-full bg-gradient-to-br ${t.gradient} opacity-10 blur-3xl transition-opacity group-hover:opacity-20`}
                  />
                  <div className="relative flex items-center gap-3">
                    <div
                      className={`flex h-11 w-11 items-center justify-center rounded-xl bg-gradient-to-br ${t.gradient} shadow-lg`}
                    >
                      {t.glyph === 'zap' ? (
                        <Zap className="h-5 w-5 text-white" />
                      ) : (
                        <Rocket className="h-5 w-5 text-white" />
                      )}
                    </div>
                    <div>
                      <h3 className="text-lg font-bold text-white">{t.name}</h3>
                      <p className="text-xs text-slate-400">{t.tagline}</p>
                    </div>
                  </div>
                  <ul className="relative mt-4 space-y-2">
                    {t.features.map((f) => (
                      <li key={f} className="flex items-center gap-2 text-sm text-slate-300">
                        <Check className="h-4 w-4 shrink-0 text-emerald-400" />
                        {f}
                      </li>
                    ))}
                  </ul>
                  <button
                    onClick={() => openCreate(t.id)}
                    className="relative mt-5 flex w-full items-center justify-center gap-1.5 rounded-xl border border-white/10 bg-white/5 px-4 py-2.5 text-sm font-semibold text-white transition hover:bg-white/10 active:scale-[0.99]"
                  >
                    {t.name} bilan bot yaratish
                    <ArrowRight className="h-4 w-4 text-cyan-300" />
                  </button>
                </div>
              ))}
            </div>
          </section>
        </main>

        {/* ===== FOOTER ===== */}
        <footer className="border-t border-white/5 py-6 text-center text-xs text-slate-500">
          BotMaker AI v1.0 · Telegram bot platformasi ·{' '}
          <span className="font-mono text-slate-400">{WEBHOOK_HOST}</span>
        </footer>
      </div>

      {/* ===== YANGI BOT WIZARD ===== */}
      <Modal
        open={createOpen}
        onClose={closeCreate}
        title="Yangi Bot Yaratish"
        subtitle={`Qadam ${wizardStep + 1}/3 — ${stepTitles[wizardStep]}`}
      >
        {/* Progress */}
        <div className="mb-6 flex items-center gap-2">
          {[0, 1, 2].map((i) => (
            <div
              key={i}
              className={`h-1.5 flex-1 rounded-full transition-all duration-300 ${
                i <= wizardStep ? 'bg-gradient-to-r from-cyan-400 to-violet-500' : 'bg-white/10'
              }`}
            />
          ))}
        </div>

        {/* Qadam 1 — Shablon tanlash */}
        {wizardStep === 0 && (
          <div className="space-y-3">
            {TEMPLATES.map((t) => {
              const selected = form.template === t.id;
              return (
                <button
                  key={t.id}
                  onClick={() => setForm((f) => ({ ...f, template: t.id }))}
                  className={`w-full rounded-xl border p-4 text-left transition-all ${
                    selected
                      ? 'border-cyan-400/50 bg-cyan-400/5 ring-2 ring-cyan-400/20'
                      : 'border-white/10 bg-white/[0.03] hover:border-white/25'
                  }`}
                >
                  <div className="flex items-center justify-between gap-3">
                    <div className="flex items-center gap-3">
                      <div
                        className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br ${t.gradient}`}
                      >
                        {t.glyph === 'zap' ? (
                          <Zap className="h-5 w-5 text-white" />
                        ) : (
                          <Rocket className="h-5 w-5 text-white" />
                        )}
                      </div>
                      <div>
                        <p className="font-semibold text-white">{t.name}</p>
                        <p className="text-xs text-slate-400">{t.tagline}</p>
                      </div>
                    </div>
                    {selected && <CheckCircle2 className="h-5 w-5 shrink-0 text-cyan-400" />}
                  </div>
                </button>
              );
            })}
          </div>
        )}

        {/* Qadam 2 — Konfiguratsiya */}
        {wizardStep === 1 && (
          <div className="space-y-4">
            <Field label="Bot nomi" icon={BotIcon}>
              <input
                value={form.name}
                onChange={(e) => setForm((f) => ({ ...f, name: e.target.value }))}
                placeholder="Masalan: AniTez Poster"
                className={inputCls}
              />
            </Field>
            <Field
              label="Bot Token"
              icon={KeyRound}
              hint="@BotFather dan olingan tokenni kiriting — server webhook'ni avtomatik sozlaydi."
            >
              <input
                value={form.token}
                onChange={(e) => setForm((f) => ({ ...f, token: e.target.value }))}
                placeholder="1234567890:AAH..."
                className={`${inputCls} font-mono`}
              />
            </Field>
            <Field label="Admin ID" icon={User} hint="Faqat shu foydalanuvchi botni boshqarishi mumkin.">
              <input
                value={form.adminId}
                onChange={(e) => setForm((f) => ({ ...f, adminId: e.target.value }))}
                placeholder="Masalan: 521348907"
                inputMode="numeric"
                className={inputCls}
              />
            </Field>
          </div>
        )}

        {/* Qadam 3 — Webhook ulash */}
        {wizardStep === 2 && (
          <div className="space-y-4">
            <div className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/[0.03] p-4">
              <div className="flex items-center gap-3">
                <Link2 className="h-4 w-4 shrink-0 text-cyan-400" />
                <div>
                  <p className="text-sm font-medium text-white">Webhookni ulash</p>
                  <p className="text-xs text-slate-400">Bot so‘rovlari shu manzilga yo‘naltiriladi</p>
                </div>
              </div>
              <Toggle
                checked={form.useWebhook}
                onChange={(v) => setForm((f) => ({ ...f, useWebhook: v }))}
              />
            </div>

            {form.useWebhook && (
              <Field label="Webhook URL" icon={Globe} hint={`Avtomatik: ${autoWebhookUrl}`}>
                <input
                  value={form.webhookUrl}
                  onChange={(e) => setForm((f) => ({ ...f, webhookUrl: e.target.value }))}
                  placeholder={autoWebhookUrl}
                  className={`${inputCls} font-mono`}
                />
              </Field>
            )}

            <div className="flex gap-2.5 rounded-xl border border-cyan-400/20 bg-cyan-400/5 p-3.5 text-xs leading-relaxed text-cyan-100/80">
              <Shield className="mt-0.5 h-4 w-4 shrink-0 text-cyan-300" />
              <span>
                Token Telegram API orqali <code className="font-mono text-cyan-200">setWebhook</code> bilan ulanadi.
                Server: <span className="font-mono text-cyan-200">{WEBHOOK_HOST}</span> (VPS).
              </span>
            </div>
          </div>
        )}

        {/* Navigatsiya */}
        <div className="mt-6 flex items-center justify-between">
          <button
            onClick={() => setWizardStep((s) => Math.max(0, s - 1))}
            disabled={wizardStep === 0}
            className="flex items-center gap-1 rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10 disabled:cursor-not-allowed disabled:opacity-40"
          >
            <ChevronLeft className="h-4 w-4" />
            Orqaga
          </button>

          {wizardStep < 2 ? (
            <button
              onClick={() => setWizardStep((s) => Math.min(2, s + 1))}
              className="flex items-center gap-1 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110"
            >
              Keyingi
              <ChevronRight className="h-4 w-4" />
            </button>
          ) : (
            <button
              onClick={() => void handleCreate()}
              disabled={busy}
              className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-emerald-500/25 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
            >
              {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Sparkles className="h-4 w-4" />}
              Botni yaratish
            </button>
          )}
        </div>
      </Modal>

      {/* ===== BROADCAST MODAL ===== */}
      <Modal
        open={broadcastOpen}
        onClose={() => setBroadcastOpen(false)}
        title="Broadcast — Xabar yuborish"
        subtitle="Bot obunachilariga ommaviy xabar jo‘natiladi"
      >
        <div className="space-y-4">
          <Field label="Qaysi bot orqali" icon={Megaphone}>
            <select
              value={broadcastTarget}
              onChange={(e) => setBroadcastTarget(e.target.value)}
              className={`${inputCls} appearance-none`}
            >
              <option value="all">Barcha botlar</option>
              {bots.map((b) => (
                <option key={b.id} value={b.id}>
                  {b.name} (@{b.username ?? '—'})
                </option>
              ))}
            </select>
          </Field>

          <Field label="Xabar matni" icon={MessageSquare} hint="Markdown va HTML formatlash qo‘llab-quvvatlanadi.">
            <textarea
              value={broadcastText}
              onChange={(e) => setBroadcastText(e.target.value)}
              rows={6}
              placeholder="Obunachilarga yuboriladigan matn..."
              className={`${inputCls} resize-none leading-relaxed`}
            />
          </Field>

          <div className="flex gap-2.5 rounded-xl border border-amber-400/20 bg-amber-400/5 p-3.5 text-xs leading-relaxed text-amber-100/80">
            <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-amber-400" />
            <span>
              Broadcast barcha obunachilarga <b>darhol</b> yetkaziladi — matnni diqqat bilan tekshiring.
            </span>
          </div>
        </div>

        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={() => setBroadcastOpen(false)}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10"
          >
            Bekor qilish
          </button>
          <button
            onClick={() => void sendBroadcast()}
            disabled={busy || !broadcastText.trim()}
            className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
            Yuborish
          </button>
        </div>
      </Modal>

      {/* ===== O‘CHIRISH TASDIQLASH ===== */}
      <Modal
        open={!!deleteTarget}
        onClose={() => setDeleteTarget(null)}
        title="Botni o‘chirish"
        subtitle="Bu amalni ortga qaytarib bo‘lmaydi"
      >
        <p className="text-sm leading-relaxed text-slate-300">
          <b className="text-white">{deleteTarget?.name}</b> boti va uning barcha ma’lumotlari (obunachilar,
          statistika) o‘chiriladi. Davom etasizmi?
        </p>
        <div className="mt-6 flex justify-end gap-3">
          <button
            onClick={() => setDeleteTarget(null)}
            className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10"
          >
            Bekor qilish
          </button>
          <button
            onClick={() => void confirmDelete()}
            disabled={busy}
            className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-rose-500 to-red-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-rose-500/25 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
          >
            {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Trash2 className="h-4 w-4" />}
            O‘chirish
          </button>
        </div>
      </Modal>

      {/* ===== TOASTLAR ===== */}
      <div className="fixed bottom-4 right-4 z-[60] flex w-full max-w-sm flex-col gap-2 px-4 sm:px-0">
        {toasts.map((t) => (
          <div
            key={t.id}
            className={`animate-modal-in flex items-start gap-3 rounded-xl border p-3.5 backdrop-blur-xl ${
              t.type === 'success'
                ? 'border-emerald-400/20 bg-emerald-500/10 text-emerald-100'
                : t.type === 'error'
                  ? 'border-rose-400/20 bg-rose-500/10 text-rose-100'
                  : 'border-cyan-400/20 bg-cyan-500/10 text-cyan-100'
            }`}
          >
            {t.type === 'success' ? (
              <CheckCircle2 className="mt-0.5 h-4 w-4 shrink-0 text-emerald-400" />
            ) : t.type === 'error' ? (
              <AlertTriangle className="mt-0.5 h-4 w-4 shrink-0 text-rose-400" />
            ) : (
              <Sparkles className="mt-0.5 h-4 w-4 shrink-0 text-cyan-400" />
            )}
            <p className="flex-1 text-sm leading-relaxed">{t.message}</p>
            <button
              onClick={() => setToasts((prev) => prev.filter((x) => x.id !== t.id))}
              className="text-current opacity-60 transition hover:opacity-100"
              aria-label="Yopish"
            >
              <X className="h-4 w-4" />
            </button>
          </div>
        ))}
      </div>
    </div>
  );
}
