'use client';

/* Broadcast — bot obunachilariga ommaviy xabar yuborish */

import { useEffect, useState } from 'react';
import { AlertTriangle, Loader2, Megaphone, MessageSquare, Send } from 'lucide-react';
import { apiFetch } from '../lib/api';
import { Field, Modal, inputCls, useToast } from './ui';
import type { BotItem } from '../lib/types';

export default function BroadcastModal({
  open,
  onClose,
  bots,
  initialTarget,
}: {
  open: boolean;
  onClose: () => void;
  bots: BotItem[];
  initialTarget?: string;
}) {
  const { push } = useToast();
  const [target, setTarget] = useState<string>('all');
  const [text, setText] = useState('');
  const [busy, setBusy] = useState(false);

  /* Modal ochilganda bot-spetsifik targetni o'rnatish */
  useEffect(() => {
    if (open && initialTarget) setTarget(initialTarget);
  }, [open, initialTarget]);

  const send = async () => {
    if (!text.trim()) return;
    setBusy(true);
    try {
      await apiFetch('/broadcast', {
        method: 'POST',
        body: JSON.stringify({
          bot_id: target === 'all' ? undefined : target,
          target,
          message: text,
        }),
      });
      push('success', target === 'all' ? 'Broadcast barcha botlarga yuborildi' : 'Broadcast muvaffaqiyatli yuborildi');
      onClose();
      setText('');
    } catch (e) {
      push('error', `Broadcast yuborilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal open={open} onClose={onClose} title="Broadcast — Xabar yuborish" subtitle="Bot obunachilariga ommaviy xabar jo‘natiladi">
      <div className="space-y-4">
        <Field label="Qaysi bot orqali" icon={Megaphone}>
          <select
            value={target}
            onChange={(e) => setTarget(e.target.value)}
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
            value={text}
            onChange={(e) => setText(e.target.value)}
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
          onClick={onClose}
          className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10"
        >
          Bekor qilish
        </button>
        <button
          onClick={() => void send()}
          disabled={busy || !text.trim()}
          className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-cyan-500 to-violet-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Send className="h-4 w-4" />}
          Yuborish
        </button>
      </div>
    </Modal>
  );
}
