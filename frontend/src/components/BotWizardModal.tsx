'use client';

/* Yangi bot yaratish — 4 bosqichli wizard (shablon → token/nom →
   sozlamalar → webhook va ishga tushirish) */

import { useEffect, useMemo, useState } from 'react';
import {
  Bot as BotIcon,
  ChevronLeft,
  ChevronRight,
  Globe,
  KeyRound,
  Link2,
  Loader2,
  Rocket,
  Shield,
} from 'lucide-react';
import { apiFetch } from '../lib/api';
import {
  CURRENCIES,
  INITIAL_FORM,
  SETTING_META,
  TEMPLATES,
  WIZARD_STEPS,
  WEBHOOK_HOST,
  slugify,
  tplOf,
} from '../lib/data';
import { Field, Modal, Toggle, inputCls, useToast } from './ui';
import type { FormState, TemplateId } from '../lib/types';

interface Props {
  open: boolean;
  initialTemplate?: TemplateId;
  onClose: () => void;
  onCreated: () => void;
}

export default function BotWizardModal({ open, initialTemplate, onClose, onCreated }: Props) {
  const { push } = useToast();
  const [wizardStep, setWizardStep] = useState(0);
  const [form, setForm] = useState<FormState>(INITIAL_FORM);
  const [busy, setBusy] = useState(false);

  const activeTpl = tplOf(form.template);

  /* Modal ochilganda wizardni qayta boshlash (tanlangan shablon bilan) */
  useEffect(() => {
    if (open) {
      setForm({ ...INITIAL_FORM, template: initialTemplate ?? INITIAL_FORM.template });
      setWizardStep(0);
    }
  }, [open, initialTemplate]);

  const autoWebhookUrl = useMemo(
    () => `https://${WEBHOOK_HOST}/webhook/${slugify(form.name || 'bot')}`,
    [form.name],
  );

  const closeWizard = () => {
    onClose();
    setWizardStep(0);
    setForm(INITIAL_FORM);
  };

  const goNext = () => {
    if (wizardStep === 1) {
      if (!form.name.trim()) {
        push('error', 'Bot nomini kiriting');
        return;
      }
      if (!form.token.trim()) {
        push('error', 'Bot tokenni kiriting — @BotFather dan oling');
        return;
      }
    }
    if (wizardStep === 2) {
      if (!form.adminId.trim()) {
        push('error', 'Admin Telegram ID kiriting');
        return;
      }
      const tpl = tplOf(form.template);
      if (tpl.settings.includes('apiKey') && !form.apiKey.trim()) {
        push('error', 'AI API kaliti (DeepSeek) kiriting');
        return;
      }
    }
    setWizardStep((s) => Math.min(WIZARD_STEPS.length - 1, s + 1));
  };

  const handleCreate = async () => {
    setBusy(true);
    try {
      await apiFetch('/bots', {
        method: 'POST',
        body: JSON.stringify({
          name: form.name.trim(),
          template: form.template,
          template_name: tplOf(form.template).name,
          token: form.token.trim(),
          admin_id: form.adminId.trim() || undefined,
          api_key: form.apiKey.trim() || undefined,
          channel_id: form.channelId.trim() || undefined,
          currency: form.currency || undefined,
          webhook_url: form.useWebhook ? form.webhookUrl.trim() || autoWebhookUrl : undefined,
        }),
      });
      push('success', `«${form.name.trim()}» boti yaratildi va ishga tushirildi`);
      closeWizard();
      onCreated();
    } catch (e) {
      push('error', `Bot yaratilmadi: ${e instanceof Error ? e.message : 'noma’lum xatolik'}`);
    } finally {
      setBusy(false);
    }
  };

  return (
    <Modal
      open={open}
      onClose={closeWizard}
      title="Yangi Bot Yaratish"
      subtitle={`Qadam ${wizardStep + 1}/4 — ${WIZARD_STEPS[wizardStep]}`}
      wide
    >
      {/* Modal ochilganda tanlangan shablonni qo'llash */}

      {/* Progress */}
      <div className="mb-5 flex items-center gap-2">
        {WIZARD_STEPS.map((_, i) => (
          <div
            key={i}
            className={`h-1.5 flex-1 rounded-full transition-all duration-300 ${
              i <= wizardStep ? 'bg-gradient-to-r from-cyan-400 to-violet-500' : 'bg-white/10'
            }`}
          />
        ))}
      </div>

      {/* Qadam 1 — Shablon */}
      {wizardStep === 0 && (
        <div className="grid max-h-[46vh] gap-3 overflow-y-auto pr-1 sm:grid-cols-2">
          {TEMPLATES.map((t) => {
            const selected = form.template === t.id;
            return (
              <button
                key={t.id}
                onClick={() => setForm((f) => ({ ...f, template: t.id }))}
                className={`relative w-full rounded-xl border p-4 text-left transition-all ${
                  selected
                    ? 'border-cyan-400/50 bg-cyan-400/5 ring-2 ring-cyan-400/20'
                    : 'border-white/10 bg-white/[0.03] hover:border-white/25'
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div className="flex items-center gap-3">
                    <div
                      className={`flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br ${t.gradient} text-xl`}
                    >
                      {t.emoji}
                    </div>
                    <div>
                      <p className="text-sm font-semibold leading-tight text-white">{t.name}</p>
                      <p className="mt-0.5 text-xs text-slate-400">{t.tagline}</p>
                    </div>
                  </div>
                  <span
                    className={`shrink-0 rounded-full border px-2 py-0.5 text-[10px] font-semibold uppercase tracking-wide ${t.chip}`}
                  >
                    {t.category}
                  </span>
                </div>
              </button>
            );
          })}
        </div>
      )}

      {/* Qadam 2 — Token va nom */}
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
          <div className="flex gap-2.5 rounded-xl border border-white/10 bg-white/[0.03] p-3.5 text-xs leading-relaxed text-slate-400">
            <Shield className="mt-0.5 h-4 w-4 shrink-0 text-violet-300" />
            <span>
              Token olish uchun Telegram'da <code className="font-mono text-violet-200">@BotFather</code> ga murojaat
              qiling → <code className="font-mono text-violet-200">/newbot</code> → tokenni shu yerga joylang.
            </span>
          </div>
        </div>
      )}

      {/* Qadam 3 — Sozlamalar */}
      {wizardStep === 2 && (
        <div className="space-y-4">
          <div className="flex items-center gap-2.5 rounded-xl border border-white/10 bg-white/[0.03] p-3.5">
            <div
              className={`flex h-9 w-9 shrink-0 items-center justify-center rounded-lg bg-gradient-to-br ${activeTpl.gradient} text-lg`}
            >
              {activeTpl.emoji}
            </div>
            <div>
              <p className="text-sm font-semibold text-white">{activeTpl.name}</p>
              <p className="text-xs text-slate-400">{activeTpl.tagline}</p>
            </div>
          </div>
          {activeTpl.settings.map((key) => {
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
        </div>
      )}

      {/* Qadam 4 — Webhook va ishga tushirish */}
      {wizardStep === 3 && (
        <div className="space-y-4">
          <div className="rounded-xl border border-white/10 bg-white/[0.03] p-4">
            <p className="mb-3 text-[10px] font-semibold uppercase tracking-wider text-slate-500">Bot xulosasi</p>
            <ul className="space-y-1.5 text-sm text-slate-300">
              <li className="flex items-center gap-2">
                <span className="w-20 shrink-0 text-slate-500">Shablon:</span>
                <span>
                  {activeTpl.emoji} {activeTpl.name}
                </span>
              </li>
              <li className="flex items-center gap-2">
                <span className="w-20 shrink-0 text-slate-500">Nomi:</span>
                <span className="font-medium text-white">{form.name || '—'}</span>
              </li>
              <li className="flex items-center gap-2">
                <span className="w-20 shrink-0 text-slate-500">Username:</span>
                <span className="font-mono text-cyan-300">@{slugify(form.name)}</span>
              </li>
              <li className="flex items-center gap-2">
                <span className="w-20 shrink-0 text-slate-500">Admin ID:</span>
                <span className="font-mono">{form.adminId || '—'}</span>
              </li>
              {form.apiKey && (
                <li className="flex items-center gap-2">
                  <span className="w-20 shrink-0 text-slate-500">AI API:</span>
                  <span className="font-mono text-emerald-300">{form.apiKey.slice(0, 8)}…</span>
                </li>
              )}
            </ul>
          </div>

          <div className="flex items-center justify-between gap-3 rounded-xl border border-white/10 bg-white/[0.03] p-4">
            <div className="flex items-center gap-3">
              <Link2 className="h-4 w-4 shrink-0 text-cyan-400" />
              <div>
                <p className="text-sm font-medium text-white">Webhookni ulash</p>
                <p className="text-xs text-slate-400">Bot so‘rovlari shu manzilga yo‘naltiriladi</p>
              </div>
            </div>
            <Toggle checked={form.useWebhook} onChange={(v) => setForm((f) => ({ ...f, useWebhook: v }))} />
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
              Token Telegram API orqali <code className="font-mono text-cyan-200">setWebhook</code> bilan ulanadi va bot
              darhol ishga tushadi. Server: <span className="font-mono text-cyan-200">{WEBHOOK_HOST}</span> (VPS).
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

        {wizardStep < WIZARD_STEPS.length - 1 ? (
          <button
            onClick={goNext}
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
            {busy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Rocket className="h-4 w-4" />}
            Webhookni ulash va ishga tushirish
          </button>
        )}
      </div>
    </Modal>
  );
}
