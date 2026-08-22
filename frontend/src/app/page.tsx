'use client';

/* BotMaker AI — SaaS Dashboard (User Panel + Admin Panel) */

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  Activity,
  AlertTriangle,
  ArrowRight,
  Bot as BotIcon,
  Brain,
  Check,
  Clock,
  Megaphone,
  Plus,
  RefreshCw,
  Server,
  Settings,
  Sparkles,
  Trash2,
  Webhook,
} from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import { apiFetch } from '../lib/api';
import { API_BASE, DEMO_BOTS, DEMO_STATS, TEMPLATES, WEBHOOK_HOST, fmtCompact, fmtNum, normalizeBots, normalizeStats, tplOf } from '../lib/data';
import { SessionProvider, useSession } from '../lib/store';
import { EmptyState, LiveDot, ToastProvider, useToast, Badge } from '../components/ui';
import Header from '../components/Header';
import AuthModal from '../components/AuthModal';
import TopUpModal from '../components/TopUpModal';
import AdminPanel from '../components/AdminPanel';
import BotWizardModal from '../components/BotWizardModal';
import BotSettingsModal from '../components/BotSettingsModal';
import BroadcastModal from '../components/BroadcastModal';
import DeleteConfirmModal from '../components/DeleteConfirmModal';
import type { BotItem, Stats, TemplateId } from '../lib/types';

export default function Page() {
  return (
    <ToastProvider>
      <SessionProvider>
        <Dashboard />
      </SessionProvider>
    </ToastProvider>
  );
}

function Dashboard() {
  const { user, adminMode } = useSession();
  const { push } = useToast();

  const [bots, setBots] = useState<BotItem[]>([]);
  const [stats, setStats] = useState<Stats>({
    totalBots: 0,
    activeWebhooks: 0,
    processedRequests: 0,
    aiResponses: 0,
    serverLoad: null,
  });
  const [loading, setLoading] = useState(true);
  const [demoMode, setDemoMode] = useState(false);
  const [busyId, setBusyId] = useState<string | null>(null);

  /* Modallar */
  const [createOpen, setCreateOpen] = useState(false);
  const [createTemplate, setCreateTemplate] = useState<TemplateId | undefined>(undefined);
  const [settingsTarget, setSettingsTarget] = useState<BotItem | null>(null);
  const [broadcastOpen, setBroadcastOpen] = useState(false);
  const [broadcastTarget, setBroadcastTarget] = useState<string>('all');
  const [deleteTarget, setDeleteTarget] = useState<BotItem | null>(null);
  const [loginOpen, setLoginOpen] = useState(false);
  const [topUpOpen, setTopUpOpen] = useState(false);

  /* ---- Ma'lumot yuklash ---- */

  const loadData = useCallback(async (silent = false): Promise<boolean> => {
    if (!silent) setLoading(true);
    try {
      const [statsRes, botsRes] = await Promise.all([apiFetch<unknown>('/stats'), apiFetch<unknown>('/bots')]);
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
      if (!silent) setLoading(false);
    }
  }, []);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  /* ---- Jonli statistika: demo rejimda simulyatsiya, realda 15s polling ---- */

  useEffect(() => {
    const demo = window.setInterval(() => {
      if (demoMode) {
        setStats((s) => ({
          ...s,
          processedRequests: s.processedRequests + Math.floor(Math.random() * 8) + 2,
          aiResponses: s.aiResponses + Math.floor(Math.random() * 3),
        }));
      }
    }, 4000);

    const live = window.setInterval(() => {
      if (!demoMode) void loadData(true);
    }, 15000);

    return () => {
      window.clearInterval(demo);
      window.clearInterval(live);
    };
  }, [demoMode, loadData]);

  const webhookLatency = useMemo(() => {
    const actives = bots.filter((b) => b.webhookActive && b.running);
    if (actives.length === 0) return 2;
    return Math.min(...actives.map((b) => b.latency ?? 2));
  }, [bots]);

  const statCards: { label: string; value: string; icon: LucideIcon; gradient: string; sub: string; bar: number | null }[] = [
    {
      label: 'Faol botlar',
      value: fmtNum(stats.totalBots),
      icon: BotIcon,
      gradient: 'from-cyan-400 to-blue-500',
      sub: 'platformada ishlamoqda',
      bar: null,
    },
    {
      label: 'Webhook so‘rovlari',
      value: fmtNum(stats.processedRequests),
      icon: Webhook,
      gradient: 'from-violet-400 to-fuchsia-500',
      sub: 'jami qayta ishlangan',
      bar: null,
    },
    {
      label: 'AI javoblari',
      value: fmtNum(stats.aiResponses),
      icon: Brain,
      gradient: 'from-emerald-400 to-teal-500',
      sub: 'AI tomonidan yaratilgan',
      bar: null,
    },
    {
      label: 'Server quvvati',
      value: stats.serverLoad != null ? `${Math.round(stats.serverLoad)}%` : '—',
      icon: Server,
      gradient: 'from-amber-400 to-orange-500',
      sub: 'CPU yuklanishi',
      bar: stats.serverLoad,
    },
  ];

  /* ---- Bot boshqaruvi ---- */

  const toggleBot = async (bot: BotItem) => {
    const next = !bot.running;
    setBusyId(bot.id);
    // Optimistik yangilash
    setBots((prev) => prev.map((b) => (b.id === bot.id ? { ...b, running: next } : b)));
    try {
      await apiFetch(`/bots/${bot.id}/${next ? 'start' : 'stop'}`, { method: 'POST' });
      push('success', `«${bot.name}» ${next ? 'ishga tushirildi' : 'to‘xtatildi'}`);
    } catch (e) {
      setBots((prev) => prev.map((b) => (b.id === bot.id ? { ...b, running: !next } : b)));
      push('error', `Amal bajarilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusyId(null);
    }
  };

  const openCreate = (template?: TemplateId) => {
    if (!user) {
      push('info', 'Yangi bot yaratish uchun avval tizimga kiring!');
      setLoginOpen(true);
      return;
    }
    setCreateTemplate(template);
    setCreateOpen(true);
  };

  const openBroadcast = (botId: string) => {
    setBroadcastTarget(botId);
    setBroadcastOpen(true);
  };

  const isAdminView = adminMode && user?.role === 'admin';

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
        <Header
          demoMode={demoMode}
          loading={loading}
          webhookLatency={webhookLatency}
          onRefresh={() => void loadData()}
          onOpenCreate={() => openCreate()}
          onOpenTopUp={() => setTopUpOpen(true)}
          onOpenLogin={() => setLoginOpen(true)}
        />

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

          {!user && (
            <div className="flex flex-wrap items-center gap-3 rounded-2xl border border-cyan-400/20 bg-cyan-400/5 px-4 py-3 text-sm text-cyan-100/90">
              <Sparkles className="h-4 w-4 shrink-0 text-cyan-300" />
              <span>
                <b className="text-white">Tizimga kiring</b> — balansingizni to‘ldiring va barcha imkoniyatlardan
                foydalaning.
              </span>
              <button
                onClick={() => setLoginOpen(true)}
                className="ml-auto rounded-lg bg-gradient-to-r from-cyan-500 to-violet-500 px-3.5 py-1.5 text-xs font-bold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110"
              >
                Kirish
              </button>
            </div>
          )}

          {isAdminView ? (
            /* ===== 👑 ADMIN PANEL ===== */
            <AdminPanel bots={bots} />
          ) : (
            <>
              {/* ===== TEZKOR STATISTIKA KARTALARI ===== */}
              <section className="grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
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
                      <div className={`flex h-10 w-10 items-center justify-center rounded-xl bg-gradient-to-br ${c.gradient} shadow-lg`}>
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
                    {c.bar != null && (
                      <div className="mt-3 h-1.5 w-full overflow-hidden rounded-full bg-white/10">
                        <div
                          className="h-full rounded-full bg-gradient-to-r from-amber-400 to-orange-500 transition-all duration-700"
                          style={{ width: `${c.bar}%` }}
                        />
                      </div>
                    )}
                  </div>
                ))}
              </section>

              {/* ===== BOTLAR BOSHQARUV PANELI ===== */}
              <section className="space-y-4">
                <div className="flex flex-wrap items-center justify-between gap-3">
                  <div className="flex flex-wrap items-center gap-2.5">
                    <div className="flex h-8 w-8 items-center justify-center rounded-lg border border-violet-400/20 bg-violet-500/10 text-violet-300">
                      <BotIcon className="h-4 w-4" />
                    </div>
                    <h2 className="text-xl font-bold text-white">Botlar Boshqaruv Paneli</h2>
                    <span className="rounded-full border border-white/10 bg-white/5 px-2.5 py-0.5 text-xs text-slate-400">
                      {bots.length} ta
                    </span>
                    <span className="flex items-center gap-1.5 rounded-full border border-emerald-400/20 bg-emerald-500/10 px-2.5 py-0.5 text-[11px] font-medium text-emerald-300">
                      <LiveDot />
                      Jonli statistika
                    </span>
                  </div>
                  <div className="flex items-center gap-2">
                    <button
                      onClick={() => openBroadcast('all')}
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
                      <div key={i} className="flex animate-pulse items-center gap-4 rounded-2xl border border-white/5 bg-white/[0.03] p-4">
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
                  <EmptyState
                    icon={<BotIcon className="h-7 w-7" />}
                    title="Hali botlar yo‘q"
                    subtitle="Birinchi Telegram botingizni yarating — «Yangi Bot» tugmasini bosing va 4 bosqichli wizardni yakunlang."
                    action={
                      <button
                        onClick={() => openCreate()}
                        className="mt-2 flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-4 py-2.5 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110"
                      >
                        <Plus className="h-4 w-4" />
                        Bot yaratish
                      </button>
                    }
                  />
                ) : (
                  <div className="space-y-3">
                    {bots.map((bot) => {
                      const tpl = tplOf(bot.template);
                      const isToggling = busyId === bot.id;
                      const reqs = bot.requests ?? 0;
                      const ai = bot.aiResponses ?? 0;
                      return (
                        <div
                          key={bot.id}
                          className="group flex flex-col gap-4 rounded-2xl border border-white/10 bg-white/[0.04] p-4 backdrop-blur-xl transition-all hover:border-white/20 hover:bg-white/[0.06] sm:flex-row sm:items-center"
                        >
                          {/* Bot ma'lumotlari */}
                          <div className="flex min-w-0 flex-1 items-center gap-4">
                            <div className={`flex h-12 w-12 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br ${tpl.gradient} text-xl shadow-lg`}>
                              {tpl.emoji}
                            </div>
                            <div className="min-w-0">
                              <div className="flex items-center gap-2">
                                <p className="truncate font-semibold text-white">{bot.name}</p>
                                {bot.running ? (
                                  <Badge tone="green">RUN</Badge>
                                ) : (
                                  <Badge tone="slate">STOP</Badge>
                                )}
                              </div>
                              <div className="mt-1 flex flex-wrap items-center gap-x-3 gap-y-1 text-xs text-slate-500">
                                <span className="font-mono">@{bot.username ?? '—'}</span>
                                <span className={`rounded-full border px-2 py-0.5 font-medium ${tpl.chip}`}>
                                  {tpl.emoji} {tpl.short}
                                </span>
                                <span className="flex items-center gap-1">
                                  <Clock className="h-3 w-3" />
                                  {bot.lastActivity ?? 'hozir'}
                                </span>
                                {reqs > 0 && (
                                  <span className="flex items-center gap-1 font-mono text-emerald-300/80">
                                    <Activity className="h-3 w-3" />
                                    {fmtCompact(reqs)} so‘rov
                                  </span>
                                )}
                                {ai > 0 && (
                                  <span className="flex items-center gap-1 font-mono text-cyan-300/80">
                                    <Brain className="h-3 w-3" />
                                    {fmtCompact(ai)} AI
                                  </span>
                                )}
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
                              <LiveDot on={bot.webhookActive} />
                              {bot.webhookActive ? `Active ${bot.latency ?? 2}ms` : 'Inactive'}
                            </span>

                            <div
                              className="flex h-9 items-center justify-center rounded-lg border border-white/10 bg-white/5"
                              title={bot.running ? 'To‘xtatish' : 'Ishga tushirish'}
                            >
                              {isToggling ? (
                                <span className="flex h-9 w-9 items-center justify-center">
                                  <RefreshCw className="h-4 w-4 animate-spin text-cyan-300" />
                                </span>
                              ) : (
                                <ToggleButton checked={bot.running} onClick={() => void toggleBot(bot)} />
                              )}
                            </div>

                            <button
                              onClick={() => setSettingsTarget(bot)}
                              className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-300 transition hover:border-violet-400/30 hover:bg-violet-400/10 hover:text-violet-300"
                              title="Bot sozlamalari"
                            >
                              <Settings className="h-4 w-4" />
                            </button>

                            <button
                              onClick={() => openBroadcast(bot.id)}
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
                  <span className="rounded-full border border-white/10 bg-white/5 px-2.5 py-0.5 text-xs text-slate-400">
                    6 toifa
                  </span>
                </div>

                <div className="grid gap-5 sm:grid-cols-2 lg:grid-cols-3">
                  {TEMPLATES.map((t) => (
                    <div
                      key={t.id}
                      className="group relative flex flex-col overflow-hidden rounded-2xl border border-white/10 bg-white/[0.04] p-6 backdrop-blur-xl transition-all hover:border-white/20"
                    >
                      <div className={`absolute inset-x-0 top-0 h-px bg-gradient-to-r ${t.gradient} opacity-70`} />
                      <div
                        className={`absolute -right-16 -top-16 h-48 w-48 rounded-full bg-gradient-to-br ${t.gradient} opacity-10 blur-3xl transition-opacity group-hover:opacity-20`}
                      />
                      <div className="relative flex items-start justify-between gap-3">
                        <div className={`flex h-12 w-12 items-center justify-center rounded-xl bg-gradient-to-br ${t.gradient} text-2xl shadow-lg`}>
                          {t.emoji}
                        </div>
                        <span className={`rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${t.chip}`}>
                          {t.category}
                        </span>
                      </div>
                      <div className="relative mt-4">
                        <h3 className="text-lg font-bold text-white">{t.name}</h3>
                        <p className="mt-0.5 text-sm text-cyan-300/80">{t.tagline}</p>
                        <p className="mt-2 text-xs leading-relaxed text-slate-400">{t.description}</p>
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
                        {t.emoji} {t.short} yaratish
                        <ArrowRight className="h-4 w-4 text-cyan-300" />
                      </button>
                    </div>
                  ))}
                </div>
              </section>
            </>
          )}
        </main>

        {/* ===== FOOTER ===== */}
        <footer className="border-t border-white/5 py-6 text-center text-xs text-slate-500">
          🤖 BotMaker AI v2.0 · Universal Telegram Botlar Konstruktori ·{' '}
          <span className="font-mono text-slate-400">{WEBHOOK_HOST}</span>
          {user && (
            <span className="mx-2 text-slate-600">·</span>
          )}
          {user && <span className="text-emerald-300/80">💰 {fmtNum(user.balance)} UZS</span>}
        </footer>
      </div>

      {/* ===== MODALLAR ===== */}
      <AuthModal open={loginOpen} onClose={() => setLoginOpen(false)} />
      <TopUpModal open={topUpOpen} onClose={() => setTopUpOpen(false)} />
      <BotWizardModal
        open={createOpen}
        initialTemplate={createTemplate}
        onClose={() => setCreateOpen(false)}
        onCreated={() => void loadData(true)}
      />
      <BotSettingsModal
        bot={settingsTarget}
        onClose={() => setSettingsTarget(null)}
        onSaved={(updated) => setBots((prev) => prev.map((b) => (b.id === updated.id ? updated : b)))}
      />
      <BroadcastModal
        open={broadcastOpen}
        onClose={() => setBroadcastOpen(false)}
        bots={bots}
        initialTarget={broadcastTarget}
      />
      <DeleteConfirmModal
        bot={deleteTarget}
        onClose={() => setDeleteTarget(null)}
        onDeleted={() => void loadData(true)}
      />
    </div>
  );
}

/* Botni ishga tushirish/stop qilish uchun kichik toggle tugmasi */
function ToggleButton({ checked, onClick }: { checked: boolean; onClick: () => void }) {
  return (
    <button
      role="switch"
      aria-checked={checked}
      onClick={onClick}
      className={`relative m-1.5 h-6 w-11 shrink-0 rounded-full transition-colors duration-200 ${
        checked ? 'bg-gradient-to-r from-emerald-400 to-teal-500 shadow-glow' : 'bg-white/10'
      }`}
      title={checked ? 'To‘xtatish' : 'Ishga tushirish'}
    >
      <span
        className={`absolute top-0.5 h-5 w-5 rounded-full bg-white shadow transition-all duration-200 ${
          checked ? 'left-[22px]' : 'left-0.5'
        }`}
      />
    </button>
  );
}
