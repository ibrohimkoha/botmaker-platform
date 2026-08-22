'use client';

/* 👑 Admin Panel — 4 tab:
   1. To'lov cheklari (tasdiqlash / rad etish + kattalashtirib ko'rish)
   2. Karta sozlamalari
   3. Yangi shablon / Git repo qo'shish
   4. Foydalanuvchilar (balans va botlar soni) */

import { useCallback, useEffect, useMemo, useState } from 'react';
import {
  ArrowLeft,
  Bot as BotIcon,
  Check,
  CheckCircle2,
  Copy,
  CreditCard,
  Crown,
  ExternalLink,
  Eye,
  FolderGit2,
  Landmark,
  Loader2,
  Plus,
  Save,
  Trash2,
  User,
  Users as UsersIcon,
  Wallet,
  X,
  XCircle,
} from 'lucide-react';
import {
  addAdminTemplate,
  applyLocalUserBalance,
  copyText,
  deleteAdminTemplate,
  fetchAdminTemplates,
  fetchCardSettings,
  fetchPaymentChecks,
  fetchPlatformUsers,
  resolvePaymentCheck,
  saveCardSettings,
} from '../lib/api';
import { fmtNum, timeAgo } from '../lib/data';
import { useSession } from '../lib/store';
import { Avatar, Badge, EmptyState, Field, Modal, inputCls, useToast } from './ui';
import type { AdminTemplate, BotItem, CardSettings, CheckStatus, PaymentCheck, PlatformUser } from '../lib/types';

type TabId = 'checks' | 'card' | 'templates' | 'users';

const TABS: { id: TabId; label: string; icon: typeof CreditCard }[] = [
  { id: 'checks', label: 'To‘lov Cheklari', icon: CreditCard },
  { id: 'card', label: 'Karta Sozlamalari', icon: Landmark },
  { id: 'templates', label: 'Shablon / Git Repo', icon: FolderGit2 },
  { id: 'users', label: 'Foydalanuvchilar', icon: UsersIcon },
];

const CHECK_FILTERS: { id: 'all' | CheckStatus; label: string }[] = [
  { id: 'all', label: 'Barchasi' },
  { id: 'pending', label: 'Kutilmoqda' },
  { id: 'approved', label: 'Tasdiqlangan' },
  { id: 'rejected', label: 'Rad etilgan' },
];

const CHECK_STATUS_META: Record<CheckStatus, { label: string; tone: string }> = {
  pending: { label: 'Kutilmoqda', tone: 'amber' },
  approved: { label: 'Tasdiqlangan', tone: 'green' },
  rejected: { label: 'Rad etilgan', tone: 'red' },
};

const TEMPLATE_CATEGORIES = ['AI', 'Biznes', 'Mijozlar', 'Marketing', 'Konstruktor', 'Media', 'Boshqa'];

const EMPTY_TEMPLATE = { name: '', repoUrl: '', price: '', category: 'AI', description: '' };

export default function AdminPanel({ bots }: { bots: BotItem[] }) {
  const { user, addBalance, setAdminMode } = useSession();
  const { push } = useToast();
  const [tab, setTab] = useState<TabId>('checks');

  /* Tab 1 — cheklar */
  const [checks, setChecks] = useState<PaymentCheck[]>([]);
  const [checkFilter, setCheckFilter] = useState<'all' | CheckStatus>('all');
  const [checksLoading, setChecksLoading] = useState(true);
  const [resolvingId, setResolvingId] = useState<string | null>(null);
  const [lightbox, setLightbox] = useState<PaymentCheck | null>(null);

  /* Tab 2 — karta */
  const [card, setCard] = useState<CardSettings | null>(null);
  const [cardForm, setCardForm] = useState({ cardNumber: '', cardHolder: '', bank: 'Uzcard' });
  const [cardSaving, setCardSaving] = useState(false);

  /* Tab 3 — shablonlar */
  const [templates, setTemplates] = useState<AdminTemplate[]>([]);
  const [tplForm, setTplForm] = useState(EMPTY_TEMPLATE);
  const [tplSaving, setTplSaving] = useState(false);

  /* Tab 4 — foydalanuvchilar */
  const [users, setUsers] = useState<PlatformUser[]>([]);
  const [usersLoading, setUsersLoading] = useState(true);

  const refreshChecks = useCallback(async () => {
    const list = await fetchPaymentChecks();
    setChecks(list);
    setChecksLoading(false);
  }, []);

  const refreshUsers = useCallback(async () => {
    const list = await fetchPlatformUsers(bots);
    setUsers(list);
    setUsersLoading(false);
  }, [bots]);

  useEffect(() => {
    void refreshChecks();
    void fetchCardSettings().then((c) => {
      setCard(c);
      setCardForm({ cardNumber: c.cardNumber, cardHolder: c.cardHolder, bank: c.bank ?? 'Uzcard' });
    });
    void fetchAdminTemplates().then(setTemplates);
    void refreshUsers();
  }, [refreshChecks, refreshUsers]);

  const filteredChecks = useMemo(
    () => (checkFilter === 'all' ? checks : checks.filter((c) => c.status === checkFilter)),
    [checks, checkFilter],
  );

  const pendingCount = useMemo(() => checks.filter((c) => c.status === 'pending').length, [checks]);

  /* ---- Tab 1: chekni tasdiqlash / rad etish ---- */

  const handleResolve = async (check: PaymentCheck, status: 'approved' | 'rejected') => {
    setResolvingId(check.id);
    try {
      await resolvePaymentCheck(check.id, status);
      if (status === 'approved') {
        if (check.userId) applyLocalUserBalance(check.userId, check.amount);
        if (user?.id === check.userId) addBalance(check.amount);
        push('success', `${fmtNum(check.amount)} UZS «${check.userName}» balansiga qo‘shildi ✅`);
      } else {
        push('info', `«${check.userName}» cheki rad etildi`);
      }
      await refreshChecks();
      await refreshUsers();
    } catch (e) {
      push('error', `Amal bajarilmadi: ${e instanceof Error ? e.message : 'xatolik'}`);
    } finally {
      setResolvingId(null);
    }
  };

  /* ---- Tab 2: karta saqlash ---- */

  const handleSaveCard = async () => {
    const number = cardForm.cardNumber.trim();
    const holder = cardForm.cardHolder.trim();
    if (number.replace(/\D/g, '').length < 12) {
      push('error', 'Karta raqami noto‘g‘ri — kamida 12 ta raqam');
      return;
    }
    if (!holder) {
      push('error', 'Karta egasi ismini kiriting');
      return;
    }
    setCardSaving(true);
    try {
      const saved = await saveCardSettings({ cardNumber: number, cardHolder: holder, bank: cardForm.bank.trim() || 'Uzcard' });
      setCard(saved);
      push('success', 'Karta sozlamalari saqlandi ✅');
    } catch (e) {
      push('error', `Saqlanmadi: ${e instanceof Error ? e.message : 'xatolik'}`);
    } finally {
      setCardSaving(false);
    }
  };

  /* ---- Tab 3: shablon qo'shish ---- */

  const handleAddTemplate = async () => {
    const name = tplForm.name.trim();
    const repoUrl = tplForm.repoUrl.trim();
    const price = Number(tplForm.price.replace(/\D/g, '')) || 0;
    if (!name) {
      push('error', 'Shablon nomini kiriting');
      return;
    }
    if (!/^https?:\/\/.+/.test(repoUrl)) {
      push('error', 'Git repo havolasi http(s):// bilan boshlanishi kerak');
      return;
    }
    setTplSaving(true);
    try {
      const created = await addAdminTemplate({
        name,
        repoUrl,
        price,
        category: tplForm.category,
        description: tplForm.description.trim() || undefined,
      });
      setTemplates((prev) => [created, ...prev]);
      setTplForm(EMPTY_TEMPLATE);
      push('success', `«${name}» shabloni qo‘shildi 📦`);
    } catch (e) {
      push('error', `Qo‘shilmadi: ${e instanceof Error ? e.message : 'xatolik'}`);
    } finally {
      setTplSaving(false);
    }
  };

  const handleDeleteTemplate = async (tpl: AdminTemplate) => {
    await deleteAdminTemplate(tpl.id);
    setTemplates((prev) => prev.filter((t) => t.id !== tpl.id));
    push('info', `«${tpl.name}» o‘chirildi`);
  };

  return (
    <section className="space-y-5">
      {/* Sarlavha */}
      <div className="flex flex-wrap items-center justify-between gap-3">
        <div className="flex items-center gap-3">
          <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-amber-400/20 bg-gradient-to-br from-amber-500/20 to-orange-500/10 text-amber-300 shadow-lg shadow-amber-500/10">
            <Crown className="h-5 w-5" />
          </div>
          <div>
            <h2 className="text-xl font-bold text-white">👑 Admin Panel</h2>
            <p className="text-xs text-slate-500">Platforma boshqaruvi — to‘lovlar, karta, shablonlar va foydalanuvchilar</p>
          </div>
        </div>
        <button
          onClick={() => setAdminMode(false)}
          className="flex items-center gap-1.5 rounded-xl border border-white/10 bg-white/5 px-3.5 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10"
        >
          <ArrowLeft className="h-4 w-4" />
          Dashboardga qaytish
        </button>
      </div>

      {/* Tablar */}
      <div className="flex flex-wrap gap-2">
        {TABS.map((t) => {
          const active = tab === t.id;
          const Icon = t.icon;
          return (
            <button
              key={t.id}
              onClick={() => setTab(t.id)}
              className={`relative flex items-center gap-1.5 rounded-xl border px-3.5 py-2 text-sm font-semibold transition ${
                active
                  ? 'border-amber-400/40 bg-amber-500/10 text-amber-200 shadow-lg shadow-amber-500/10'
                  : 'border-white/10 bg-white/5 text-slate-300 hover:bg-white/10'
              }`}
            >
              <Icon className={`h-4 w-4 ${active ? 'text-amber-300' : 'text-slate-400'}`} />
              {t.label}
              {t.id === 'checks' && pendingCount > 0 && (
                <span className="flex h-5 min-w-5 items-center justify-center rounded-full bg-rose-500 px-1 text-[10px] font-bold text-white">
                  {pendingCount}
                </span>
              )}
            </button>
          );
        })}
      </div>

      {/* ============ TAB 1: TO'LOV CHEKLARI ============ */}
      {tab === 'checks' && (
        <div className="space-y-4">
          <div className="flex flex-wrap items-center gap-2">
            {CHECK_FILTERS.map((f) => (
              <button
                key={f.id}
                onClick={() => setCheckFilter(f.id)}
                className={`rounded-full border px-3 py-1 text-xs font-semibold transition ${
                  checkFilter === f.id
                    ? 'border-cyan-400/40 bg-cyan-500/10 text-cyan-300'
                    : 'border-white/10 bg-white/5 text-slate-400 hover:bg-white/10'
                }`}
              >
                {f.label}
              </button>
            ))}
          </div>

          {checksLoading ? (
            <div className="space-y-3">
              {[0, 1, 2].map((i) => (
                <div key={i} className="h-20 animate-pulse rounded-2xl border border-white/5 bg-white/[0.03]" />
              ))}
            </div>
          ) : filteredChecks.length === 0 ? (
            <EmptyState
              icon={<CreditCard className="h-7 w-7" />}
              title="Cheklar yo‘q"
              subtitle="Foydalanuvchilar chek yuborganda shu yerda ko‘rinadi."
            />
          ) : (
            <div className="space-y-3">
              {filteredChecks.map((check) => {
                const meta = CHECK_STATUS_META[check.status];
                return (
                  <div
                    key={check.id}
                    className="flex flex-col gap-3 rounded-2xl border border-white/10 bg-white/[0.04] p-4 backdrop-blur-xl transition hover:border-white/20 sm:flex-row sm:items-center"
                  >
                    <div className="flex min-w-0 flex-1 items-center gap-3">
                      <Avatar name={check.userName} size="sm" />
                      <div className="min-w-0">
                        <div className="flex flex-wrap items-center gap-2">
                          <p className="truncate text-sm font-semibold text-white">{check.userName}</p>
                          <Badge tone={meta.tone}>{meta.label}</Badge>
                        </div>
                        <p className="mt-0.5 text-xs text-slate-500">
                          {timeAgo(check.createdAt)} · ID: <span className="font-mono">{check.id}</span>
                        </p>
                      </div>
                    </div>

                    <div className="flex items-center gap-3 pl-10 sm:pl-0">
                      <span className="font-mono text-base font-bold text-emerald-300">
                        💰 {fmtNum(check.amount)} {check.currency}
                      </span>

                      {check.screenshotUrl && (
                        <button
                          onClick={() => setLightbox(check)}
                          className="flex h-9 items-center gap-1.5 rounded-lg border border-white/10 bg-white/5 px-2.5 text-xs font-medium text-slate-300 transition hover:border-cyan-400/30 hover:bg-cyan-400/10 hover:text-cyan-300"
                          title="Screenshotni kattalashtirib ko‘rish"
                        >
                          <Eye className="h-4 w-4" /> Ko‘rish
                        </button>
                      )}

                      {check.status === 'pending' && (
                        <div className="flex items-center gap-2">
                          <button
                            onClick={() => void handleResolve(check, 'approved')}
                            disabled={resolvingId === check.id}
                            className="flex items-center gap-1 rounded-lg bg-gradient-to-r from-emerald-500 to-teal-500 px-3 py-2 text-xs font-bold text-white shadow-lg shadow-emerald-500/25 transition hover:brightness-110 disabled:opacity-60"
                          >
                            {resolvingId === check.id ? (
                              <Loader2 className="h-3.5 w-3.5 animate-spin" />
                            ) : (
                              <Check className="h-3.5 w-3.5" />
                            )}
                            Tasdiqlash
                          </button>
                          <button
                            onClick={() => void handleResolve(check, 'rejected')}
                            disabled={resolvingId === check.id}
                            className="flex items-center gap-1 rounded-lg border border-rose-400/30 bg-rose-500/10 px-3 py-2 text-xs font-bold text-rose-300 transition hover:bg-rose-500/20 disabled:opacity-60"
                          >
                            <X className="h-3.5 w-3.5" /> Rad etish
                          </button>
                        </div>
                      )}
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </div>
      )}

      {/* ============ TAB 2: KARTA SOZLAMALARI ============ */}
      {tab === 'card' && (
        <div className="grid gap-5 lg:grid-cols-2">
          <div className="space-y-4 rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-xl">
            <h3 className="text-sm font-bold text-white">Karta ma'lumotlarini tahrirlash</h3>
            <Field label="Karta raqami" icon={CreditCard} hint="Foydalanuvchilarga to‘lov qilish uchun ko‘rsatiladi.">
              <input
                value={cardForm.cardNumber}
                onChange={(e) => setCardForm((f) => ({ ...f, cardNumber: e.target.value }))}
                placeholder="8600 0000 0000 0000"
                className={`${inputCls} font-mono`}
              />
            </Field>
            <Field label="Karta egasi (F.I.Sh.)" icon={User}>
              <input
                value={cardForm.cardHolder}
                onChange={(e) => setCardForm((f) => ({ ...f, cardHolder: e.target.value.toUpperCase() }))}
                placeholder="ISKANDAROV AZIZ"
                className={`${inputCls} font-mono uppercase`}
              />
            </Field>
            <Field label="Bank turi" icon={Landmark}>
              <select
                value={cardForm.bank}
                onChange={(e) => setCardForm((f) => ({ ...f, bank: e.target.value }))}
                className={`${inputCls} appearance-none`}
              >
                <option value="Uzcard">Uzcard</option>
                <option value="Humo">Humo</option>
                <option value="Visa">Visa</option>
                <option value="Mastercard">Mastercard</option>
              </select>
            </Field>
            <button
              onClick={() => void handleSaveCard()}
              disabled={cardSaving}
              className="flex w-full items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-violet-500 to-fuchsia-500 px-5 py-2.5 text-sm font-bold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 disabled:opacity-60"
            >
              {cardSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Save className="h-4 w-4" />}
              Karta sozlamalarini saqlash
            </button>
          </div>

          <div className="space-y-4">
            <p className="text-xs font-medium text-slate-400">Foydalanuvchilarga ko‘rinadigan ko‘rinish</p>
            {card && (
              <div className="relative overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-br from-[#14143a] via-[#1b1640] to-[#0e2a3a] p-6 shadow-xl shadow-violet-950/40">
                <div aria-hidden className="pointer-events-none absolute -right-10 -top-14 h-40 w-40 rounded-full bg-cyan-500/20 blur-3xl" />
                <div aria-hidden className="pointer-events-none absolute -bottom-16 -left-10 h-40 w-40 rounded-full bg-violet-500/20 blur-3xl" />
                <div className="relative flex items-center justify-between">
                  <span className="flex items-center gap-1.5 text-xs font-semibold uppercase tracking-widest text-slate-300">
                    <Landmark className="h-3.5 w-3.5" /> {card.bank ?? 'Bank karta'}
                  </span>
                  <span className="rounded-md border border-white/10 bg-white/10 px-2 py-0.5 font-mono text-[10px] text-slate-300">
                    UZCARD / HUMO
                  </span>
                </div>
                <p className="relative mt-5 font-mono text-xl font-bold tracking-wider text-white">
                  {card.cardNumber.replace(/\D/g, '').replace(/(\d{4})(?=\d)/g, '$1 ') || '•••• •••• •••• ••••'}
                </p>
                <div className="relative mt-4 flex items-end justify-between">
                  <div>
                    <p className="text-[10px] uppercase tracking-wider text-slate-400">Karta egasi</p>
                    <p className="mt-0.5 text-sm font-semibold uppercase tracking-wide text-white">{card.cardHolder}</p>
                  </div>
                  <button
                    onClick={() => void copyText(card.cardNumber.replace(/\s/g, '')).then((ok) => ok && push('success', 'Karta raqami nusxalandi'))}
                    className="flex items-center gap-1 rounded-lg border border-white/10 bg-white/10 px-2.5 py-1.5 text-xs font-semibold text-white transition hover:bg-white/20"
                  >
                    <Copy className="h-3.5 w-3.5" /> Nusxalash
                  </button>
                </div>
              </div>
            )}
            <p className="text-xs leading-relaxed text-slate-500">
              💡 O‘zgarishlar «Balansni To‘ldirish» oynasida darhol aks etadi.
            </p>
          </div>
        </div>
      )}

      {/* ============ TAB 3: SHABLON / GIT REPO ============ */}
      {tab === 'templates' && (
        <div className="grid gap-5 lg:grid-cols-2">
          <div className="space-y-4 rounded-2xl border border-white/10 bg-white/[0.04] p-5 backdrop-blur-xl">
            <h3 className="text-sm font-bold text-white">Yangi shablon qo‘shish</h3>
            <Field label="Git repozitoriy havolasi" icon={ExternalLink} hint="Masalan: https://github.com/user/repo">
              <input
                value={tplForm.repoUrl}
                onChange={(e) => setTplForm((f) => ({ ...f, repoUrl: e.target.value }))}
                placeholder="https://github.com/..."
                className={`${inputCls} font-mono`}
              />
            </Field>
            <Field label="Shablon nomi" icon={FolderGit2}>
              <input
                value={tplForm.name}
                onChange={(e) => setTplForm((f) => ({ ...f, name: e.target.value }))}
                placeholder="Masalan: Kino Bot Pro"
                className={inputCls}
              />
            </Field>
            <div className="grid grid-cols-2 gap-3">
              <Field label="Narxi (UZS)" icon={Wallet}>
                <input
                  value={tplForm.price}
                  onChange={(e) => setTplForm((f) => ({ ...f, price: e.target.value.replace(/[^\d]/g, '') }))}
                  inputMode="numeric"
                  placeholder="0 — bepul"
                  className={`${inputCls} font-mono`}
                />
              </Field>
              <Field label="Toifasi" icon={FolderGit2}>
                <select
                  value={tplForm.category}
                  onChange={(e) => setTplForm((f) => ({ ...f, category: e.target.value }))}
                  className={`${inputCls} appearance-none`}
                >
                  {TEMPLATE_CATEGORIES.map((c) => (
                    <option key={c} value={c}>
                      {c}
                    </option>
                  ))}
                </select>
              </Field>
            </div>
            <Field label="Tavsif (ixtiyoriy)" icon={FolderGit2}>
              <textarea
                value={tplForm.description}
                onChange={(e) => setTplForm((f) => ({ ...f, description: e.target.value }))}
                rows={3}
                placeholder="Shablon haqida qisqacha..."
                className={`${inputCls} resize-none leading-relaxed`}
              />
            </Field>
            <button
              onClick={() => void handleAddTemplate()}
              disabled={tplSaving}
              className="flex w-full items-center justify-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-5 py-2.5 text-sm font-bold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 disabled:opacity-60"
            >
              {tplSaving ? <Loader2 className="h-4 w-4 animate-spin" /> : <Plus className="h-4 w-4" />}
              Shablonni qo‘shish
            </button>
          </div>

          <div className="space-y-3">
            <p className="text-xs font-medium text-slate-400">Qo‘shilgan shablonlar ({templates.length})</p>
            {templates.length === 0 ? (
              <EmptyState
                icon={<FolderGit2 className="h-7 w-7" />}
                title="Hali shablonlar yo‘q"
                subtitle="Git repozitoriy orqali yangi shablon qo‘shing — u katalogda paydo bo‘ladi."
              />
            ) : (
              templates.map((t) => (
                <div
                  key={t.id}
                  className="flex items-center gap-3 rounded-2xl border border-white/10 bg-white/[0.04] p-4 backdrop-blur-xl transition hover:border-white/20"
                >
                  <div className="flex h-11 w-11 shrink-0 items-center justify-center rounded-xl border border-cyan-400/20 bg-cyan-500/10 text-lg">
                    📦
                  </div>
                  <div className="min-w-0 flex-1">
                    <div className="flex flex-wrap items-center gap-2">
                      <p className="truncate text-sm font-semibold text-white">{t.name}</p>
                      <Badge tone="cyan">{t.category}</Badge>
                      <span className="font-mono text-xs font-bold text-emerald-300">
                        {t.price > 0 ? `${fmtNum(t.price)} UZS` : 'Bepul'}
                      </span>
                    </div>
                    <p className="mt-0.5 truncate font-mono text-xs text-slate-500">{t.repoUrl}</p>
                    {t.description && <p className="mt-1 line-clamp-1 text-xs text-slate-400">{t.description}</p>}
                  </div>
                  <div className="flex shrink-0 items-center gap-2">
                    <a
                      href={t.repoUrl}
                      target="_blank"
                      rel="noreferrer noopener"
                      className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-300 transition hover:border-cyan-400/30 hover:bg-cyan-400/10 hover:text-cyan-300"
                      title="Reponi ochish"
                    >
                      <ExternalLink className="h-4 w-4" />
                    </a>
                    <button
                      onClick={() => void handleDeleteTemplate(t)}
                      className="flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/5 text-slate-300 transition hover:border-rose-400/30 hover:bg-rose-400/10 hover:text-rose-400"
                      title="O‘chirish"
                    >
                      <Trash2 className="h-4 w-4" />
                    </button>
                  </div>
                </div>
              ))
            )}
          </div>
        </div>
      )}

      {/* ============ TAB 4: FOYDALANUVCHILAR ============ */}
      {tab === 'users' && (
        <div className="overflow-hidden rounded-2xl border border-white/10 bg-white/[0.03] backdrop-blur-xl">
          {usersLoading ? (
            <div className="space-y-3 p-4">
              {[0, 1, 2].map((i) => (
                <div key={i} className="h-14 animate-pulse rounded-xl bg-white/[0.04]" />
              ))}
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full min-w-[640px] text-left text-sm">
                <thead>
                  <tr className="border-b border-white/10 text-xs uppercase tracking-wider text-slate-500">
                    <th className="px-4 py-3 font-semibold">Foydalanuvchi</th>
                    <th className="px-4 py-3 font-semibold">Rol</th>
                    <th className="px-4 py-3 font-semibold">Balans</th>
                    <th className="px-4 py-3 font-semibold">Botlar</th>
                    <th className="px-4 py-3 font-semibold">Ro‘yxatdan o‘tgan</th>
                  </tr>
                </thead>
                <tbody>
                  {users.map((u) => (
                    <tr key={u.id} className="border-b border-white/5 transition hover:bg-white/[0.03]">
                      <td className="px-4 py-3">
                        <div className="flex items-center gap-3">
                          <Avatar name={u.name} src={u.avatar} size="sm" />
                          <div className="min-w-0">
                            <p className="truncate font-semibold text-white">
                              {u.name}
                              {u.id === user?.id && (
                                <span className="ml-1.5 rounded-md border border-cyan-400/20 bg-cyan-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-cyan-300">
                                  SIZ
                                </span>
                              )}
                            </p>
                            {u.username && <p className="truncate font-mono text-xs text-slate-500">@{u.username}</p>}
                          </div>
                        </div>
                      </td>
                      <td className="px-4 py-3">
                        {u.role === 'admin' ? (
                          <Badge tone="amber">
                            <Crown className="h-3 w-3" /> Admin
                          </Badge>
                        ) : (
                          <Badge tone="slate">Foydalanuvchi</Badge>
                        )}
                      </td>
                      <td className="px-4 py-3 font-mono font-semibold text-emerald-300">💰 {fmtNum(u.balance)} UZS</td>
                      <td className="px-4 py-3">
                        <span className="flex items-center gap-1.5 font-mono text-slate-200">
                          <BotIcon className="h-3.5 w-3.5 text-violet-300" /> {u.botCount}
                        </span>
                      </td>
                      <td className="px-4 py-3 text-xs text-slate-400">
                        {u.joinedAt ? timeAgo(u.joinedAt) : '—'}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>
      )}

      {/* ===== Chek screenshot lightbox ===== */}
      <Modal
        open={!!lightbox}
        onClose={() => setLightbox(null)}
        title="To‘lov cheki — screenshot"
        subtitle={lightbox ? `«${lightbox.userName}» · ${fmtNum(lightbox.amount)} ${lightbox.currency}` : undefined}
        wide
      >
        {lightbox?.screenshotUrl && (
          <div className="flex justify-center">
            {/* eslint-disable-next-line @next/next/no-img-element */}
            <img
              src={lightbox.screenshotUrl}
              alt="To‘lov cheki screenshot"
              className="max-h-[65vh] w-full rounded-xl object-contain"
            />
          </div>
        )}
        {lightbox && lightbox.status === 'pending' && (
          <div className="mt-4 flex justify-end gap-3">
            <button
              onClick={() => void handleResolve(lightbox, 'rejected').then(() => setLightbox(null))}
              className="flex items-center gap-1.5 rounded-xl border border-rose-400/30 bg-rose-500/10 px-4 py-2 text-sm font-semibold text-rose-300 transition hover:bg-rose-500/20"
            >
              <XCircle className="h-4 w-4" /> Rad etish
            </button>
            <button
              onClick={() => void handleResolve(lightbox, 'approved').then(() => setLightbox(null))}
              className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-500 px-4 py-2 text-sm font-bold text-white shadow-lg shadow-emerald-500/25 transition hover:brightness-110"
            >
              <CheckCircle2 className="h-4 w-4" /> Tasdiqlash
            </button>
          </div>
        )}
      </Modal>
    </section>
  );
}
