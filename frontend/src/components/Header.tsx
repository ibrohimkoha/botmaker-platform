'use client';

/* Sahifa sarlavhasi: logotip, server holati, foydalanuvchi profili,
   balans, balansni to'ldirish tugmasi va Admin Panel kaliti */

import { useState } from 'react';
import {
  Bot as BotIcon,
  ChevronDown,
  Crown,
  Gauge,
  LogIn,
  LogOut,
  Plus,
  RefreshCw,
  Server,
  Wallet,
} from 'lucide-react';
import { fmtNum, WEBHOOK_HOST } from '../lib/data';
import { useSession } from '../lib/store';
import { Avatar, LiveDot, Toggle } from './ui';

interface HeaderProps {
  demoMode: boolean;
  loading: boolean;
  webhookLatency: number;
  onRefresh: () => void;
  onOpenCreate: () => void;
  onOpenTopUp: () => void;
  onOpenLogin: () => void;
}

export default function Header({
  demoMode,
  loading,
  webhookLatency,
  onRefresh,
  onOpenCreate,
  onOpenTopUp,
  onOpenLogin,
}: HeaderProps) {
  const { user, adminMode, setAdminMode, logout } = useSession();
  const [menuOpen, setMenuOpen] = useState(false);

  const closeMenu = () => setMenuOpen(false);

  return (
    <header className="sticky top-0 z-40 -mx-4 border-b border-white/5 bg-[#05050f]/70 px-4 backdrop-blur-xl sm:-mx-6 sm:px-6 lg:-mx-8 lg:px-8">
      <div className="flex h-16 items-center justify-between gap-3">
        {/* Logotip */}
        <div className="flex min-w-0 items-center gap-3">
          <div className="relative flex h-10 w-10 shrink-0 items-center justify-center rounded-xl bg-gradient-to-br from-cyan-400 via-violet-500 to-fuchsia-500 shadow-lg shadow-violet-500/30">
            <BotIcon className="h-5 w-5 text-white" />
            <span className="absolute -right-1 -top-1 h-2.5 w-2.5 rounded-full bg-emerald-400 ring-2 ring-[#05050f]" />
          </div>
          <div className="min-w-0">
            <h1 className="truncate text-sm font-bold leading-tight tracking-tight text-white sm:text-base lg:text-lg">
              🤖 <span className="text-gradient">BotMaker AI</span>
            </h1>
            <p className="mt-0.5 hidden max-w-2xl truncate text-[11px] leading-tight text-slate-500 xl:block">
              Bozor narxidan 10x arzonroq, yuqori sifatli va Webhook asosida chaqmoqdek tez ishlovchi Telegram
              botlar yarating.
            </p>
          </div>
        </div>

        {/* O'ng blok */}
        <div className="flex shrink-0 items-center gap-2 sm:gap-3">
          {/* VPS Server holati */}
          <div className="hidden items-center gap-2 rounded-full border border-white/10 bg-white/5 px-3 py-1.5 backdrop-blur lg:flex">
            <LiveDot on={!demoMode} />
            <Server className="h-3.5 w-3.5 text-slate-400" />
            <span className="font-mono text-xs text-slate-300">{WEBHOOK_HOST}</span>
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
          <div className="hidden items-center gap-1.5 rounded-full border border-emerald-400/20 bg-emerald-500/10 px-3 py-1.5 md:flex">
            <Gauge className="h-3.5 w-3.5 text-emerald-300" />
            <span className="text-xs font-medium text-emerald-300">
              Webhook: <span className="font-semibold">Active {webhookLatency}ms</span>
            </span>
          </div>

          <button
            onClick={onRefresh}
            disabled={loading}
            className="flex h-9 w-9 items-center justify-center rounded-xl border border-white/10 bg-white/5 text-slate-300 transition hover:bg-white/10 hover:text-white disabled:opacity-50"
            title="Ma'lumotlarni yangilash"
          >
            <RefreshCw className={`h-4 w-4 ${loading ? 'animate-spin' : ''}`} />
          </button>

          {/* Balansni to'ldirish (faqat tizimga kirilganda) */}
          {user && (
            <button
              onClick={onOpenTopUp}
              className="flex h-9 items-center gap-1.5 rounded-xl border border-emerald-400/30 bg-emerald-500/10 px-3 text-sm font-semibold text-emerald-300 transition hover:bg-emerald-500/20 active:scale-[0.98]"
              title="Balansni to'ldirish"
            >
              <Wallet className="h-4 w-4" />
              <span className="hidden md:inline">Balansni to‘ldirish</span>
            </button>
          )}

          {/* 👑 Admin Panel kaliti */}
          {user?.role === 'admin' && (
            <div className="flex items-center gap-2 rounded-full border border-amber-400/20 bg-amber-500/10 px-3 py-1.5">
              <Crown className="h-3.5 w-3.5 text-amber-300" />
              <span className="hidden text-xs font-medium text-amber-200 md:inline">Admin Panel</span>
              <Toggle checked={adminMode} onChange={setAdminMode} />
            </div>
          )}

          {/* Foydalanuvchi profili */}
          {user ? (
            <div className="relative">
              <button
                onClick={() => setMenuOpen((v) => !v)}
                className="flex items-center gap-2 rounded-xl border border-white/10 bg-white/5 py-1 pl-1 pr-2 transition hover:bg-white/10"
                aria-label="Profil menyusi"
              >
                <Avatar name={user.name} src={user.avatar} size="sm" />
                <span className="hidden max-w-[110px] truncate text-sm font-semibold text-white sm:block">
                  {user.name}
                </span>
                <span className="hidden items-center gap-1 rounded-md border border-emerald-400/20 bg-emerald-500/10 px-1.5 py-0.5 font-mono text-[11px] font-semibold text-emerald-300 xl:flex">
                  💰 {fmtNum(user.balance)} UZS
                </span>
                <ChevronDown
                  className={`h-4 w-4 shrink-0 text-slate-400 transition-transform ${menuOpen ? 'rotate-180' : ''}`}
                />
              </button>

              {menuOpen && (
                <>
                  <div className="fixed inset-0 z-40" onClick={closeMenu} />
                  <div className="animate-modal-in absolute right-0 top-full z-50 mt-2 w-72 rounded-2xl border border-white/10 bg-[#0b0b1a]/95 p-4 shadow-2xl shadow-violet-950/50 backdrop-blur-2xl">
                    <div className="flex items-center gap-3">
                      <Avatar name={user.name} src={user.avatar} size="lg" />
                      <div className="min-w-0">
                        <p className="truncate text-sm font-bold text-white">
                          {user.name}
                          {user.role === 'admin' && <Crown className="ml-1 inline h-3.5 w-3.5 text-amber-300" />}
                        </p>
                        {user.email ? (
                          <p className="truncate text-xs text-slate-500">{user.email}</p>
                        ) : user.username ? (
                          <p className="truncate font-mono text-xs text-slate-500">@{user.username}</p>
                        ) : (
                          <p className="text-xs text-slate-500">ID: {user.id}</p>
                        )}
                      </div>
                    </div>

                    <div className="mt-4 flex items-center justify-between rounded-xl border border-emerald-400/20 bg-emerald-500/10 px-3 py-2.5">
                      <div>
                        <p className="text-[11px] text-slate-400">Balans</p>
                        <p className="text-sm font-bold text-emerald-300">💰 {fmtNum(user.balance)} UZS</p>
                      </div>
                      <button
                        onClick={() => {
                          closeMenu();
                          onOpenTopUp();
                        }}
                        className="rounded-lg bg-gradient-to-r from-emerald-500 to-teal-500 px-3 py-1.5 text-xs font-semibold text-white shadow-lg shadow-emerald-500/25 transition hover:brightness-110"
                      >
                        To‘ldirish
                      </button>
                    </div>

                    {user.role === 'admin' && (
                      <div className="mt-3 flex items-center justify-between rounded-xl border border-amber-400/20 bg-amber-500/10 px-3 py-2.5">
                        <span className="flex items-center gap-1.5 text-sm font-medium text-amber-200">
                          <Crown className="h-4 w-4 text-amber-300" /> Admin Panel
                        </span>
                        <Toggle
                          checked={adminMode}
                          onChange={(v) => {
                            setAdminMode(v);
                            closeMenu();
                          }}
                        />
                      </div>
                    )}

                    <button
                      onClick={() => {
                        closeMenu();
                        logout();
                      }}
                      className="mt-3 flex w-full items-center justify-center gap-1.5 rounded-xl border border-rose-400/20 bg-rose-500/10 px-4 py-2 text-sm font-medium text-rose-300 transition hover:bg-rose-500/20"
                    >
                      <LogOut className="h-4 w-4" /> Chiqish
                    </button>
                  </div>
                </>
              )}
            </div>
          ) : (
            <button
              onClick={onOpenLogin}
              className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-3.5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 active:scale-[0.98]"
            >
              <LogIn className="h-4 w-4" />
              Kirish
            </button>
          )}

          <button
            onClick={onOpenCreate}
            className="hidden items-center gap-1.5 rounded-xl bg-gradient-to-r from-violet-500 to-fuchsia-500 px-3.5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 active:scale-[0.98] sm:flex"
          >
            <Plus className="h-4 w-4" />
            Yangi Bot
          </button>
        </div>
      </div>
    </header>
  );
}
