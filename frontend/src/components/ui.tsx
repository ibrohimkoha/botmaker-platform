'use client';

/* Umumiy UI primitivlari: Toast, Modal, Toggle, LiveDot, Field, Badge */

import { createContext, useCallback, useContext, useMemo, useState } from 'react';
import type { ReactNode } from 'react';
import { AlertTriangle, CheckCircle2, Sparkles, X } from 'lucide-react';
import type { LucideIcon } from 'lucide-react';
import type { ToastMsg } from '../lib/types';

export const inputCls =
  'w-full rounded-xl border border-white/10 bg-white/[0.04] px-3.5 py-2.5 text-sm text-white placeholder:text-slate-500 outline-none transition focus:border-cyan-400/50 focus:bg-white/[0.06] focus:ring-2 focus:ring-cyan-400/20';

/* ============================================================
   TOAST TIZIMI
============================================================ */

const ToastContext = createContext<{ push: (type: ToastMsg['type'], message: string) => void } | null>(
  null,
);

export function ToastProvider({ children }: { children: ReactNode }) {
  const [toasts, setToasts] = useState<ToastMsg[]>([]);
  const push = useCallback((type: ToastMsg['type'], message: string) => {
    const id = Date.now() + Math.random();
    setToasts((prev) => [...prev, { id, type, message }]);
    window.setTimeout(() => {
      setToasts((prev) => prev.filter((t) => t.id !== id));
    }, 4200);
  }, []);
  const value = useMemo(() => ({ push }), [push]);

  return (
    <ToastContext.Provider value={value}>
      {children}
      <ToastViewport toasts={toasts} dismiss={(id) => setToasts((prev) => prev.filter((t) => t.id !== id))} />
    </ToastContext.Provider>
  );
}

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error('useToast ToastProvider ichida ishlatilishi kerak');
  return ctx;
}

function ToastViewport({ toasts, dismiss }: { toasts: ToastMsg[]; dismiss: (id: number) => void }) {
  return (
    <div className="fixed bottom-4 right-4 z-[70] flex w-full max-w-sm flex-col gap-2 px-4 sm:px-0">
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
            onClick={() => dismiss(t.id)}
            className="text-current opacity-60 transition hover:opacity-100"
            aria-label="Yopish"
          >
            <X className="h-4 w-4" />
          </button>
        </div>
      ))}
    </div>
  );
}

/* ============================================================
   MODAL
============================================================ */

export function Modal({
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
  children: ReactNode;
  wide?: boolean;
}) {
  if (!open) return null;
  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center p-4">
      <div className="absolute inset-0 bg-black/70 backdrop-blur-sm" onClick={onClose} />
      <div
        className={`animate-modal-in relative w-full ${wide ? 'max-w-2xl' : 'max-w-lg'} max-h-[90vh] overflow-y-auto rounded-2xl border border-white/10 bg-[#0b0b1a]/95 p-6 shadow-2xl shadow-violet-950/50 backdrop-blur-2xl`}
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

/* ============================================================
   TOGGLE
============================================================ */

export function Toggle({ checked, onChange }: { checked: boolean; onChange: (v: boolean) => void }) {
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

/* ============================================================
   LIVE DOT
============================================================ */

export function LiveDot({ on = true }: { on?: boolean }) {
  return (
    <span className="relative flex h-1.5 w-1.5">
      {on && <span className="absolute inline-flex h-full w-full animate-ping rounded-full bg-emerald-400 opacity-60" />}
      <span className={`relative inline-flex h-1.5 w-1.5 rounded-full ${on ? 'bg-emerald-400' : 'bg-slate-500'}`} />
    </span>
  );
}

/* ============================================================
   FIELD (label + input wrapper)
============================================================ */

export function Field({
  label,
  icon: Icon,
  hint,
  children,
}: {
  label: string;
  icon?: LucideIcon;
  hint?: string;
  children: ReactNode;
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

/* ============================================================
   BADGE
============================================================ */

const badgeTones: Record<string, string> = {
  green: 'border-emerald-400/20 bg-emerald-500/10 text-emerald-300',
  red: 'border-rose-400/20 bg-rose-500/10 text-rose-300',
  amber: 'border-amber-400/20 bg-amber-500/10 text-amber-300',
  cyan: 'border-cyan-400/20 bg-cyan-500/10 text-cyan-300',
  violet: 'border-violet-400/20 bg-violet-500/10 text-violet-300',
  slate: 'border-slate-500/20 bg-slate-500/10 text-slate-400',
};

export function Badge({ children, tone = 'slate' }: { children: ReactNode; tone?: string }) {
  return (
    <span
      className={`inline-flex items-center gap-1 rounded-full border px-2 py-0.5 text-[11px] font-medium ${badgeTones[tone] ?? badgeTones.slate}`}
    >
      {children}
    </span>
  );
}

/* ============================================================
   BO'SH HOLAT
============================================================ */

export function EmptyState({
  icon,
  title,
  subtitle,
  action,
}: {
  icon: ReactNode;
  title: string;
  subtitle: string;
  action?: ReactNode;
}) {
  return (
    <div className="flex flex-col items-center gap-3 rounded-2xl border border-dashed border-white/15 bg-white/[0.02] px-6 py-12 text-center">
      <div className="flex h-14 w-14 items-center justify-center rounded-2xl border border-white/10 bg-white/5 text-slate-400">
        {icon}
      </div>
      <p className="font-semibold text-white">{title}</p>
      <p className="max-w-sm text-sm text-slate-500">{subtitle}</p>
      {action && <div className="mt-1">{action}</div>}
    </div>
  );
}

/* ============================================================
   AVATAR
============================================================ */

export function Avatar({
  name,
  src,
  size = 'md',
}: {
  name: string;
  src?: string;
  size?: 'sm' | 'md' | 'lg';
}) {
  const dims =
    size === 'sm' ? 'h-8 w-8 rounded-lg text-sm' : size === 'lg' ? 'h-12 w-12 rounded-xl text-lg' : 'h-10 w-10 rounded-xl text-base';
  const initial = (name.trim() || '?').slice(0, 1).toUpperCase();
  if (src) {
    // eslint-disable-next-line @next/next/no-img-element
    return <img src={src} alt={name} className={`${dims} shrink-0 object-cover`} />;
  }
  return (
    <span
      className={`${dims} flex shrink-0 items-center justify-center bg-gradient-to-br from-cyan-400 to-violet-500 font-bold text-white shadow-lg shadow-violet-500/20`}
    >
      {initial}
    </span>
  );
}
