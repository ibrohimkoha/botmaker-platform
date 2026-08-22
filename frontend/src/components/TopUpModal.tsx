'use client';

/* Balansni to'ldirish modali:
   — Admin karta raqami/egasi (nusxalash tugmasi bilan)
   — To'lov summasi (tezkor variantlar bilan)
   — Chek screenshot yuklash (dropzone + darhol preview)
   — Backblaze B2 ga yuklab, admin tasdig'iga yuborish */

import { useCallback, useEffect, useRef, useState } from 'react';
import {
  Banknote,
  Check,
  Copy,
  CreditCard,
  ImagePlus,
  Landmark,
  Loader2,
  Send,
  UploadCloud,
  X,
} from 'lucide-react';
import { copyText, fetchCardSettings, submitPaymentCheck, uploadScreenshot } from '../lib/api';
import { fmtNum } from '../lib/data';
import { useSession } from '../lib/store';
import { Field, Modal, inputCls, useToast } from './ui';
import type { CardSettings } from '../lib/types';

const QUICK_AMOUNTS = [50000, 100000, 200000, 500000];

/* Karta raqamini 4 xonali guruhlarga ajratish */
function formatCardNumber(v: string): string {
  const digits = v.replace(/\D/g, '').slice(0, 16);
  return digits.replace(/(\d{4})(?=\d)/g, '$1 ');
}

export default function TopUpModal({ open, onClose }: { open: boolean; onClose: () => void }) {
  const { user } = useSession();
  const { push } = useToast();
  const [card, setCard] = useState<CardSettings | null>(null);
  const [amount, setAmount] = useState('');
  const [file, setFile] = useState<File | null>(null);
  const [preview, setPreview] = useState<string | null>(null);
  const [busy, setBusy] = useState(false);
  const [dragOver, setDragOver] = useState(false);
  const [copied, setCopied] = useState<'card' | 'holder' | null>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  /* Modal ochilganda karta sozlamalarini yuklash */
  useEffect(() => {
    if (!open) return;
    setCard(null);
    setAmount('');
    setFile(null);
    setPreview(null);
    setCopied(null);
    void fetchCardSettings().then((c) => setCard(c));
  }, [open]);

  const amountNum = Math.max(0, Number(amount.replace(/\D/g, '')) || 0);

  const handleCopy = async (text: string, key: 'card' | 'holder') => {
    const ok = await copyText(text);
    if (ok) {
      setCopied(key);
      push('success', key === 'card' ? 'Karta raqami nusxalandi' : 'Karta egasi nusxalandi');
      window.setTimeout(() => setCopied(null), 2000);
    } else {
      push('error', 'Nusxalash imkonsiz — qo‘lda ko‘chiring');
    }
  };

  const onFile = useCallback((f: File | undefined) => {
    if (!f) return;
    if (!f.type.startsWith('image/')) {
      push('error', 'Faqat rasm fayllari qabul qilinadi (PNG/JPG)');
      return;
    }
    if (f.size > 8 * 1024 * 1024) {
      push('error', 'Fayl hajmi 8 MB dan oshmasligi kerak');
      return;
    }
    setFile(f);
    const reader = new FileReader();
    reader.onload = () => setPreview(String(reader.result));
    reader.readAsDataURL(f);
  }, [push]);

  const submit = async () => {
    if (amountNum < 1000) {
      push('error', 'To‘lov summasini kiriting (kamida 1 000 UZS)');
      return;
    }
    if (!file || !preview) {
      push('error', 'To‘lov cheki (screenshot) yuklang');
      return;
    }
    setBusy(true);
    try {
      const { url, storage } = await uploadScreenshot(file);
      await submitPaymentCheck({
        amount: amountNum,
        currency: 'UZS',
        screenshotUrl: url,
        userId: user?.id,
        userName: user?.name ?? 'Anonim foydalanuvchi',
      });
      push(
        'success',
        storage === 'b2'
          ? 'Chek Backblaze B2 ga yuklandi va admin tasdig‘iga yuborildi ✅'
          : 'Chek admin tasdig‘iga yuborildi ✅ (demo rejim — Backblaze B2 sozlanmagan)',
      );
      onClose();
    } catch (e) {
      push('error', `Yuborilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal
      open={open}
      onClose={onClose}
      title="Balansni To‘ldirish"
      subtitle="Karta orqali o‘tkazma qiling va chekni yuklab yuboring"
    >
      <div className="space-y-4">
        {/* Karta ma'lumotlari */}
        <div>
          <p className="mb-2 flex items-center gap-1.5 text-xs font-medium text-slate-400">
            <CreditCard className="h-3.5 w-3.5 text-cyan-400/70" />
            O‘tkazma qilinadigan karta (Uzcard / Humo)
          </p>

          {card ? (
            <div className="relative overflow-hidden rounded-2xl border border-white/10 bg-gradient-to-br from-[#14143a] via-[#1b1640] to-[#0e2a3a] p-5 shadow-xl shadow-violet-950/40">
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
                {formatCardNumber(card.cardNumber)}
              </p>
              <div className="relative mt-4 flex items-end justify-between">
                <div>
                  <p className="text-[10px] uppercase tracking-wider text-slate-400">Karta egasi</p>
                  <p className="mt-0.5 text-sm font-semibold uppercase tracking-wide text-white">
                    {card.cardHolder}
                  </p>
                </div>
                <div className="flex gap-2">
                  <button
                    onClick={() => handleCopy(card.cardNumber.replace(/\s/g, ''), 'card')}
                    className="flex items-center gap-1 rounded-lg border border-white/10 bg-white/10 px-2.5 py-1.5 text-xs font-semibold text-white transition hover:bg-white/20"
                    title="Karta raqamini nusxalash"
                  >
                    {copied === 'card' ? <Check className="h-3.5 w-3.5 text-emerald-300" /> : <Copy className="h-3.5 w-3.5" />}
                    Nusxalash
                  </button>
                  <button
                    onClick={() => handleCopy(card.cardHolder, 'holder')}
                    className="flex items-center gap-1 rounded-lg border border-white/10 bg-white/10 px-2.5 py-1.5 text-xs font-semibold text-white transition hover:bg-white/20"
                    title="Karta egasini nusxalash"
                  >
                    {copied === 'holder' ? <Check className="h-3.5 w-3.5 text-emerald-300" /> : <Copy className="h-3.5 w-3.5" />}
                    Ega
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <div className="h-32 animate-pulse rounded-2xl border border-white/5 bg-white/[0.03]" />
          )}
        </div>

        {/* Summa */}
        <Field label="To‘lov summasi" icon={Banknote} hint="O‘tkazma qilgan summani kiriting — chek shu summaga tekshiriladi.">
          <div className="relative">
            <input
              value={amount}
              onChange={(e) => setAmount(e.target.value.replace(/[^\d]/g, ''))}
              inputMode="numeric"
              placeholder="Masalan: 50000"
              className={`${inputCls} pr-16 font-mono`}
            />
            <span className="absolute right-3.5 top-1/2 -translate-y-1/2 font-mono text-xs font-semibold text-slate-400">
              UZS
            </span>
          </div>
          <div className="mt-2 flex flex-wrap gap-2">
            {QUICK_AMOUNTS.map((a) => (
              <button
                key={a}
                onClick={() => setAmount(String(a))}
                className={`rounded-lg border px-2.5 py-1 text-xs font-semibold transition ${
                  amountNum === a
                    ? 'border-emerald-400/40 bg-emerald-500/10 text-emerald-300'
                    : 'border-white/10 bg-white/5 text-slate-300 hover:bg-white/10'
                }`}
              >
                {fmtNum(a)}
              </button>
            ))}
          </div>
        </Field>

        {/* Chek yuklash */}
        <div>
          <p className="mb-2 flex items-center gap-1.5 text-xs font-medium text-slate-400">
            <ImagePlus className="h-3.5 w-3.5 text-cyan-400/70" />
            To‘lov cheki (screenshot / rasm)
          </p>

          {preview ? (
            <div className="relative overflow-hidden rounded-2xl border border-emerald-400/20 bg-emerald-500/5 p-2">
              {/* eslint-disable-next-line @next/next/no-img-element */}
              <img src={preview} alt="Chek preview" className="max-h-56 w-full rounded-xl object-contain" />
              <div className="mt-2 flex items-center justify-between px-1">
                <span className="flex items-center gap-1.5 truncate text-xs text-slate-400">
                  <Check className="h-3.5 w-3.5 shrink-0 text-emerald-400" />
                  <span className="truncate">{file?.name}</span>
                </span>
                <div className="flex shrink-0 gap-2">
                  <button
                    onClick={() => inputRef.current?.click()}
                    className="rounded-lg border border-white/10 bg-white/5 px-2.5 py-1 text-xs font-semibold text-slate-200 transition hover:bg-white/10"
                  >
                    Almashtirish
                  </button>
                  <button
                    onClick={() => {
                      setFile(null);
                      setPreview(null);
                    }}
                    className="flex h-7 w-7 items-center justify-center rounded-lg border border-rose-400/20 bg-rose-500/10 text-rose-300 transition hover:bg-rose-500/20"
                    aria-label="Rasmni olib tashlash"
                  >
                    <X className="h-3.5 w-3.5" />
                  </button>
                </div>
              </div>
            </div>
          ) : (
            <button
              type="button"
              onClick={() => inputRef.current?.click()}
              onDragOver={(e) => {
                e.preventDefault();
                setDragOver(true);
              }}
              onDragLeave={() => setDragOver(false)}
              onDrop={(e) => {
                e.preventDefault();
                setDragOver(false);
                onFile(e.dataTransfer.files[0]);
              }}
              className={`flex w-full flex-col items-center gap-2 rounded-2xl border-2 border-dashed px-6 py-8 text-center transition ${
                dragOver
                  ? 'border-cyan-400/60 bg-cyan-400/10'
                  : 'border-white/15 bg-white/[0.02] hover:border-white/30 hover:bg-white/[0.04]'
              }`}
            >
              <span className="flex h-12 w-12 items-center justify-center rounded-2xl border border-white/10 bg-white/5 text-cyan-300">
                <UploadCloud className="h-6 w-6" />
              </span>
              <span className="text-sm font-semibold text-white">
                Chek rasmni shu yerga tashlang yoki bosing
              </span>
              <span className="text-xs text-slate-500">PNG / JPG · 8 MB gacha · Preview darhol ko‘rinadi</span>
            </button>
          )}
          <input
            ref={inputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={(e) => onFile(e.target.files?.[0])}
          />
        </div>

        {/* Xulosa */}
        <div className="flex items-center justify-between rounded-xl border border-white/10 bg-white/[0.03] px-4 py-3">
          <span className="text-sm text-slate-400">Jami to‘lanadigan</span>
          <span className="font-mono text-lg font-bold text-emerald-300">
            💰 {amountNum > 0 ? fmtNum(amountNum) : '0'} UZS
          </span>
        </div>

        <button
          onClick={() => void submit()}
          disabled={busy}
          className="flex w-full items-center justify-center gap-2 rounded-xl bg-gradient-to-r from-emerald-500 to-teal-500 px-5 py-3 text-sm font-bold text-white shadow-lg shadow-emerald-500/25 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          Chekni yuborish
        </button>
        <p className="text-center text-[11px] text-slate-500">
          Chek admin tomonidan tasdiqlangach balansingizga qo‘shiladi.
        </p>
      </div>
    </Modal>
  );
}
