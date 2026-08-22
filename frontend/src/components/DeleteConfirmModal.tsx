'use client';

/* Botni o'chirishni tasdiqlash modali */

import { useState } from 'react';
import { Loader2, Trash2 } from 'lucide-react';
import { apiFetch } from '../lib/api';
import { Modal, useToast } from './ui';
import type { BotItem } from '../lib/types';

export default function DeleteConfirmModal({
  bot,
  onClose,
  onDeleted,
}: {
  bot: BotItem | null;
  onClose: () => void;
  onDeleted: () => void;
}) {
  const { push } = useToast();
  const [busy, setBusy] = useState(false);

  const confirmDelete = async () => {
    if (!bot) return;
    setBusy(true);
    try {
      await apiFetch(`/bots/${bot.id}`, { method: 'DELETE' });
      push('success', `«${bot.name}» o‘chirildi`);
      onClose();
      onDeleted();
    } catch (e) {
      push('error', `O‘chirishda xatolik: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal
      open={!!bot}
      onClose={onClose}
      title="Botni o‘chirish"
      subtitle="Bu amalni ortga qaytarib bo‘lmaydi"
    >
      <p className="text-sm leading-relaxed text-slate-300">
        <b className="text-white">{bot?.name}</b> boti va uning barcha ma’lumotlari (obunachilar, statistika)
        o‘chiriladi. Davom etasizmi?
      </p>
      <div className="mt-6 flex justify-end gap-3">
        <button
          onClick={onClose}
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
  );
}
