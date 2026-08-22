'use client';

/* Bot sozlamalari modali:
   — Modal ochilganda GET /api/bots/:id/settings orqali saqlangan
     sozlamalarni yuklab, inputlarga to'ldiradi (Admin ID, Kanal ID,
     Webhook URL, API kalit, valyuta).
   — «Saqlash» bosilganda PUT /api/bots/:id/settings orqali saqlaydi
     va muvaffaqiyat haqida toast ko'rsatadi. */

import { useEffect, useState } from 'react';
import { Check, Link2, Loader2 } from 'lucide-react';
import { fetchBotSettings, saveBotSettings } from '../lib/api';
import { CURRENCIES, SETTING_META, WEBHOOK_HOST, tplOf } from '../lib/data';
import { Field, Modal, inputCls, useToast } from './ui';
import type { BotItem, BotSettings } from '../lib/types';

const EMPTY: BotSettings = { adminId: '', apiKey: '', channelId: '', currency: 'UZS', webhookUrl: '' };

export default function BotSettingsModal({
  bot,
  onClose,
  onSaved,
}: {
  bot: BotItem | null;
  onClose: () => void;
  onSaved: (updated: BotItem) => void;
}) {
  const { push } = useToast();
  const [form, setForm] = useState<BotSettings>(EMPTY);
  const [loading, setLoading] = useState(false);
  const [busy, setBusy] = useState(false);

  /* Modal ochilganda GET /api/bots/:id/settings orqali sozlamalarni yuklash */
  useEffect(() => {
    if (!bot) return;
    let cancelled = false;
    setLoading(true);
    void fetchBotSettings(bot.id, {
      adminId: bot.adminId,
      apiKey: bot.apiKey,
      channelId: bot.channelId,
      currency: bot.currency,
      webhookUrl: bot.webhookUrl,
    }).then((s) => {
      if (cancelled) return;
      setForm(s);
      setLoading(false);
    });
    return () => {
      cancelled = true;
    };
  }, [bot]);

  if (!bot) return null;
  const tpl = tplOf(bot.template);

  const save = async () => {
    setBusy(true);
    try {
      const via = await saveBotSettings(bot.id, form);
      const savedVia =
        via === 'api' ? 'server' : via === 'legacy' ? 'server (legacy)' : 'lokal (demo rejim)';
      push('success', `«${bot.name}» sozlamalari saqlandi ✅ (${savedVia})`);
      onSaved({
        ...bot,
        adminId: form.adminId.trim() || undefined,
        apiKey: form.apiKey.trim() || undefined,
        channelId: form.channelId.trim() || undefined,
        currency: form.currency.trim() || undefined,
        webhookUrl: form.webhookUrl.trim() || undefined,
      });
      onClose();
    } catch (e) {
      push('error', `Sozlamalar saqlanmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal
      open={!!bot}
      onClose={onClose}
      title="Bot sozlamalari"
      subtitle={`«${bot.name}» sozlamalari — serverdan yuklanmoqda`}
    >
      {loading ? (
        <div className="space-y-4 py-2">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-16 animate-pulse rounded-xl bg-white/[0.04]" />
          ))}
          <p className="flex items-center justify-center gap-2 text-xs text-slate-500">
            <Loader2 className="h-3.5 w-3.5 animate-spin" />
            GET /api/bots/{bot.id}/settings yuklanmoqda...
          </p>
        </div>
      ) : (
        <div className="space-y-4">
          {tpl.settings.map((key) => {
            const meta = SETTING_META[key];
            const Icon = meta.icon;
            return (
              <Field key={key} label={meta.label} icon={Icon} hint={meta.hint}>
                {key === 'currency' ? (
                  <select
                    value={form.currency}
                    onChange={(e) => setForm((f) => ({ ...f, currency: e.target.value }))}
                    className={`${inputCls} appearance-none`}
                  >
                    {CURRENCIES.map((c) => (
                      <option key={c} value={c}>
                        {c}
                      </option>
                    ))}
                  </select>
                ) : (
                  <input
                    value={form[key]}
                    onChange={(e) => setForm((f) => ({ ...f, [key]: e.target.value }))}
                    placeholder={meta.placeholder}
                    inputMode={key === 'adminId' ? 'numeric' : undefined}
                    className={`${inputCls} ${key === 'apiKey' ? 'font-mono' : ''}`}
                  />
                )}
              </Field>
            );
          })}

          <Field
            label="Webhook URL"
            icon={Link2}
            hint="Bot so‘rovlari qabul qilinadigan manzil — GET /api/bots/:id/settings orqali yuklandi."
          >
            <input
              value={form.webhookUrl}
              onChange={(e) => setForm((f) => ({ ...f, webhookUrl: e.target.value }))}
              placeholder={`https://${WEBHOOK_HOST}/webhook/...`}
              className={`${inputCls} font-mono`}
            />
          </Field>
        </div>
      )}

      <div className="mt-6 flex justify-end gap-3">
        <button
          onClick={onClose}
          className="rounded-xl border border-white/10 bg-white/5 px-4 py-2 text-sm font-medium text-slate-200 transition hover:bg-white/10"
        >
          Bekor qilish
        </button>
        <button
          onClick={() => void save()}
          disabled={busy || loading}
          className="flex items-center gap-1.5 rounded-xl bg-gradient-to-r from-violet-500 to-fuchsia-500 px-5 py-2 text-sm font-semibold text-white shadow-lg shadow-violet-500/25 transition hover:brightness-110 disabled:cursor-not-allowed disabled:opacity-60"
        >
          {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Check className="h-4 w-4" />}
          Saqlash (PUT /api/bots/{bot.id}/settings)
        </button>
      </div>
    </Modal>
  );
}
