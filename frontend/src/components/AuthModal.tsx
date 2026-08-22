'use client';

/* Kirish modali: Google 1-Click, Telegram Login Widget yoki Tezkor ID.
   Real OAuth ma'lumotlari sozlanmagan bo'lsa demo login ishlaydi. */

import { useCallback, useEffect, useRef, useState } from 'react';
import { Check, Crown, KeyRound, ShieldCheck, Zap } from 'lucide-react';
import { useSession } from '../lib/store';
import { useToast } from './ui';
import { Modal, inputCls } from './ui';
import type { AuthMethod, Role, UserProfile } from '../lib/types';

const GOOGLE_CLIENT_ID =
  process.env.NEXT_PUBLIC_GOOGLE_CLIENT_ID ||
  '531444252311-44j9qjp0ek0jbi5eggbmj9nrl317au54.apps.googleusercontent.com';
const TG_BOT_USERNAME = process.env.NEXT_PUBLIC_TG_BOT_USERNAME || 'botmakerauthbot';

const SUPER_ADMIN_IDS = ['5415350162', '6149114216'];
const SUPER_ADMIN_EMAILS = ['miraxmedovibrohim5@gmail.com'];

/* ---- Tashqi servislar uchun minimal turlar ---- */

interface GoogleCredentialResponse {
  credential: string;
}

interface GoogleGsi {
  accounts?: {
    id?: {
      initialize: (opts: { client_id: string; callback: (r: GoogleCredentialResponse) => void }) => void;
      renderButton: (el: HTMLElement | null, opts: Record<string, unknown>) => void;
      prompt: () => void;
    };
  };
}

interface TelegramAuthUser {
  id: number;
  first_name: string;
  last_name?: string;
  username?: string;
  photo_url?: string;
  auth_date?: number;
  hash?: string;
}

/* ---- Yordamchilar ---- */

function loadScript(src: string): Promise<void> {
  return new Promise((resolve, reject) => {
    if (document.querySelector(`script[src="${src}"]`)) {
      resolve();
      return;
    }
    const s = document.createElement('script');
    s.src = src;
    s.async = true;
    s.onload = () => resolve();
    s.onerror = () => reject(new Error('Script yuklanmadi'));
    document.head.appendChild(s);
  });
}

function decodeJwt(token: string): Record<string, unknown> | null {
  try {
    const payload = token.split('.')[1];
    if (!payload) return null;
    const b64 = payload.replace(/-/g, '+').replace(/_/g, '/');
    const json = decodeURIComponent(
      atob(b64)
        .split('')
        .map((c) => '%' + ('00' + c.charCodeAt(0).toString(16)).slice(-2))
        .join(''),
    );
    return JSON.parse(json) as Record<string, unknown>;
  } catch {
    return null;
  }
}

/* ---- Logotiplar ---- */

function GoogleIcon() {
  return (
    <svg className="h-5 w-5" viewBox="0 0 24 24" aria-hidden>
      <path
        fill="#4285F4"
        d="M22.56 12.25c0-.78-.07-1.53-.2-2.25H12v4.26h5.92a5.06 5.06 0 0 1-2.2 3.32v2.77h3.57c2.08-1.92 3.27-4.74 3.27-8.1z"
      />
      <path
        fill="#34A853"
        d="M12 23c2.97 0 5.46-.98 7.28-2.66l-3.57-2.77c-.98.66-2.23 1.06-3.71 1.06-2.86 0-5.29-1.93-6.16-4.53H2.18v2.84A11 11 0 0 0 12 23z"
      />
      <path
        fill="#FBBC05"
        d="M5.84 14.1a6.6 6.6 0 0 1 0-4.2V7.06H2.18a11 11 0 0 0 0 9.88l3.66-2.84z"
      />
      <path
        fill="#EA4335"
        d="M12 5.38c1.62 0 3.06.56 4.21 1.64l3.15-3.15A11 11 0 0 0 2.18 7.06l3.66 2.84c.87-2.6 3.3-4.52 6.16-4.52z"
      />
    </svg>
  );
}

function TelegramIcon() {
  return (
    <svg className="h-5 w-5" viewBox="0 0 24 24" aria-hidden>
      <path
        fill="#229ED9"
        d="M11.94 2a10 10 0 1 0 0 20 10 10 0 0 0 0-20zm4.61 6.9c-.16.83-1.1 6.34-1.55 8.4-.2.9-.58 1.2-.95 1.23-.8.06-1.41-.53-2.2-1.04-1.22-.8-1.91-1.3-3.1-2.08-1.37-.9-.48-1.4.3-2.2.2-.21 3.66-3.36 3.73-3.65.01-.03.02-.15-.06-.2-.07-.06-.19-.04-.27-.02-.11.02-1.9 1.2-5.36 3.54-.5.35-.96.52-1.37.51-.45-.01-1.32-.26-1.97-.47-.79-.26-1.42-.4-1.37-.85.03-.23.34-.47.94-.72 3.7-1.61 6.17-2.67 7.4-3.19 3.53-1.47 4.26-1.72 4.74-1.73.1 0 .34.02.49.14.12.1.15.23.17.33.02.1.04.34.03.52z"
      />
    </svg>
  );
}

/* ---- Modal ---- */

export default function AuthModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { user, login } = useSession();
  const { push } = useToast();
  const googleRef = useRef<HTMLDivElement>(null);
  const tgRef = useRef<HTMLDivElement>(null);

  /* Balansni saqlab qolish uchun oldingi profil bilan birlashtirish */
  const doLogin = useCallback(
    (profile: Omit<UserProfile, 'balance' | 'joinedAt'>) => {
      const prev = user;
      const isSuperAdmin =
        SUPER_ADMIN_IDS.includes(profile.id) ||
        (profile.email && SUPER_ADMIN_EMAILS.includes(profile.email.toLowerCase()));

      const finalRole: Role = isSuperAdmin ? 'admin' : (profile.role ?? 'user');

      const merged: UserProfile = {
        ...profile,
        role: finalRole,
        balance: prev && prev.name === profile.name ? prev.balance : 0,
        joinedAt: prev && prev.name === profile.name ? prev.joinedAt : new Date().toISOString(),
      };
      login(merged);
      push('success', `Xush kelibsiz, ${profile.name}! 👋 ${finalRole === 'admin' ? '(👑 Admin huquqi berildi)' : ''}`);
      onClose();
    },
    [user, login, push, onClose],
  );

  /* Google 1-Click (GIS) */
  useEffect(() => {
    if (!open || !GOOGLE_CLIENT_ID || !googleRef.current) return;
    let cancelled = false;
    loadScript('https://accounts.google.com/gsi/client')
      .then(() => {
        if (cancelled) return;
        const g = (window as unknown as { google?: GoogleGsi }).google;
        if (!g?.accounts?.id) return;
        g.accounts.id.initialize({
          client_id: GOOGLE_CLIENT_ID,
          callback: (resp) => {
            const payload = decodeJwt(resp.credential);
            if (!payload) return;
            doLogin({
              id: String(payload.sub ?? 'google-user'),
              name: String(payload.name ?? payload.given_name ?? 'Google Foydalanuvchi'),
              email: payload.email ? String(payload.email) : undefined,
              avatar: payload.picture ? String(payload.picture) : undefined,
              role: 'user',
              authMethod: 'google',
            });
          },
        });
        g.accounts.id.renderButton(googleRef.current, {
          theme: 'filled_black',
          size: 'large',
          width: 320,
          text: 'continue_with',
        });
      })
      .catch(() => {
        /* GIS yuklanmasa */
      });
    return () => {
      cancelled = true;
    };
  }, [open, doLogin]);

  /* Telegram Login Widget */
  useEffect(() => {
    if (!open || !TG_BOT_USERNAME || !tgRef.current) return;
    const host = tgRef.current;
    if (host.querySelector('script')) return;
    (window as unknown as Record<string, unknown>).onTelegramAuth = (u: TelegramAuthUser) => {
      doLogin({
        id: String(u.id),
        name: [u.first_name, u.last_name].filter(Boolean).join(' ') || 'Telegram Foydalanuvchi',
        username: u.username,
        avatar: u.photo_url,
        role: 'user',
        authMethod: 'telegram',
      });
    };
    const s = document.createElement('script');
    s.src = 'https://telegram.org/js/telegram-widget.js?22';
    s.async = true;
    s.setAttribute('data-telegram-login', TG_BOT_USERNAME);
    s.setAttribute('data-size', 'large');
    s.setAttribute('data-onauth', 'onTelegramAuth');
    s.setAttribute('data-request-access', 'write');
    host.appendChild(s);
  }, [open, doLogin]);

  const handleGoogleDirect = () => {
    const redirectUri = encodeURIComponent('https://nokori-uz.duckdns.org/botmaker/api/auth/callback/google');
    const authUrl = `https://accounts.google.com/o/oauth2/v2/auth?client_id=${GOOGLE_CLIENT_ID}&redirect_uri=${redirectUri}&response_type=code&scope=openid%20email%20profile&access_type=offline&prompt=consent`;
    window.location.href = authUrl;
  };

  return (
    <Modal open={open} onClose={onClose} title="Tizimga kirish" subtitle="Bot yaratish va platformadan foydalanish uchun kiring">
      <div className="space-y-4">
        {/* Google 1-Click */}
        <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
          <p className="mb-3 flex items-center justify-between text-sm font-semibold text-white">
            <span className="flex items-center gap-2">
              <GoogleIcon /> Google hisobi orqali kirish
            </span>
            <span className="rounded-md border border-cyan-400/20 bg-cyan-500/10 px-1.5 py-0.5 text-[10px] font-semibold text-cyan-300">
              1-CLICK
            </span>
          </p>
          <div ref={googleRef} className="flex justify-center overflow-hidden rounded-xl" />
          <button
            onClick={handleGoogleDirect}
            className="mt-2.5 flex w-full items-center justify-center gap-2 rounded-xl border border-white/10 bg-white/5 px-4 py-2.5 text-xs font-semibold text-slate-300 transition hover:bg-white/10 hover:text-white"
          >
            <GoogleIcon /> Google orqali to‘g‘ridan-to‘g‘ri kirish
          </button>
        </div>

        {/* Telegram */}
        <div className="rounded-2xl border border-white/10 bg-white/[0.03] p-4">
          <p className="mb-3 flex items-center justify-between text-sm font-semibold text-white">
            <span className="flex items-center gap-2">
              <TelegramIcon /> Telegram orqali kirish
            </span>
            <span className="text-xs text-cyan-400">@{TG_BOT_USERNAME}</span>
          </p>
          <div ref={tgRef} className="flex justify-center" />
          <a
            href={`https://t.me/${TG_BOT_USERNAME}?start=auth`}
            target="_blank"
            rel="noopener noreferrer"
            className="mt-2.5 flex w-full items-center justify-center gap-2 rounded-xl border border-sky-400/20 bg-sky-500/10 px-4 py-2.5 text-xs font-semibold text-sky-300 transition hover:bg-sky-500/20"
          >
            <TelegramIcon /> @{TG_BOT_USERNAME} botiga o‘tish
          </a>
        </div>

        <div className="flex items-center gap-2.5 rounded-xl border border-emerald-400/20 bg-emerald-500/5 px-3.5 py-3 text-xs leading-relaxed text-emerald-100/80">
          <ShieldCheck className="mt-0.5 h-4 w-4 shrink-0 text-emerald-300" />
          <span>
            <Check className="mr-1 inline h-3 w-3" />
            Kirish ma'lumotlaringiz xavfsiz himoyalangan.
          </span>
        </div>
      </div>
    </Modal>
  );
}
