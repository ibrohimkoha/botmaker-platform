/* BotMaker AI — statik ma'lumotlar, doimiylar va yordamchi funksiyalar */

import type { LucideIcon } from 'lucide-react';
import { Coins, Globe, KeyRound, User } from 'lucide-react';
import type {
  BotItem,
  CardSettings,
  FormState,
  PaymentCheck,
  PlatformUser,
  SettingKey,
  Stats,
  Template,
  TemplateId,
} from './types';

export const API_BASE = process.env.NEXT_PUBLIC_API_BASE ?? 'http://localhost:8085';
export const WEBHOOK_HOST = 'nokori-uz.duckdns.org';

/* ---- 6 o'rnatilgan shablon ---- */

export const TEMPLATES: Template[] = [
  {
    id: 'ai-chatbot',
    emoji: '🤖',
    category: 'AI',
    name: 'AI Aqlli Chatbot',
    short: 'AI Chatbot',
    tagline: 'DeepSeek AI yordamchisi',
    description:
      'DeepSeek AI asosidagi aqlli suhbatdosh — mijozlaringizga 24/7 savollarga javob beradi, savdo va maslahat beradi.',
    gradient: 'from-cyan-400 via-blue-500 to-violet-500',
    chip: 'border-cyan-400/20 bg-cyan-500/10 text-cyan-300',
    features: [
      'DeepSeek AI integratsiyasi (API kalit bilan)',
      '24/7 avtomatik javoblar',
      'Suhbat tarixi va kontekst',
      'Bir nechta tilni tushunadi (uz/ru/en)',
    ],
    settings: ['adminId', 'apiKey'],
  },
  {
    id: 'ecommerce',
    emoji: '🛒',
    category: 'Biznes',
    name: "E-Commerce & Online Do'kon",
    short: 'E-Commerce',
    tagline: 'Katalog, Savatcha, Click/Payme to‘lov',
    description:
      "Mahsulot katalogi, savatcha va Click/Payme orqali to'lov qabul qiluvchi to'liq online do'kon boti.",
    gradient: 'from-amber-400 via-orange-500 to-rose-500',
    chip: 'border-amber-400/20 bg-amber-500/10 text-amber-300',
    features: [
      'Mahsulot katalogi va qidiruv',
      'Savatcha va buyurtma tizimi',
      'Click / Payme to‘lov integratsiyasi',
      'Buyurtmalar kanalga tushadi',
    ],
    settings: ['adminId', 'channelId', 'currency'],
  },
  {
    id: 'feedback',
    emoji: '💬',
    category: 'Mijozlar',
    name: "Feedback & Qo'llab-quvvatlash",
    short: 'Feedback',
    tagline: 'Mijozlar bilan 2 tomonlama chat',
    description:
      "Mijozlar sizga xabar yozadi, siz bot orqali javob berasiz — qo'llab-quvvatlash xizmatini botga topshiring.",
    gradient: 'from-emerald-400 via-teal-500 to-cyan-500',
    chip: 'border-emerald-400/20 bg-emerald-500/10 text-emerald-300',
    features: [
      '2 tomonlama mijoz chat',
      'Xabarlar admin Telegramga yetib boradi',
      'Tezkor javob berish interfeysi',
      'Reyting va baholash',
    ],
    settings: ['adminId'],
  },
  {
    id: 'channel',
    emoji: '📢',
    category: 'Marketing',
    name: 'Kanal & Majburiy Obuna Menejeri',
    short: 'Kanal Menejeri',
    tagline: 'Zayafkalarni avto-tasdiqlash',
    description:
      "Kanalingizga majburiy obunani tekshiradi va zayafkalarni avtomatik tasdiqlaydi — obunachilar bazasini o'stiring.",
    gradient: 'from-fuchsia-400 via-pink-500 to-rose-500',
    chip: 'border-fuchsia-400/20 bg-fuchsia-500/10 text-fuchsia-300',
    features: [
      'Majburiy obuna tekshiruvi',
      'Zayafkalarni avto-tasdiqlash',
      'A‘zo soni statistikasi',
      'Cheklangan kontent himoyasi',
    ],
    settings: ['adminId', 'channelId'],
  },
  {
    id: 'visual-menu',
    emoji: '🧩',
    category: 'Konstruktor',
    name: 'Maxsus Vizual Menyu & Konstruktor',
    short: 'Vizual Menyu',
    tagline: 'Tugmalar va avtojavoblar',
    description:
      "Hech qanday kod yozmasdan tugmalar, menyular va avtojavoblar bilan o'z botingizni yig'ing.",
    gradient: 'from-violet-500 via-purple-500 to-fuchsia-500',
    chip: 'border-violet-400/20 bg-violet-500/10 text-violet-300',
    features: [
      'Vizual tugma konstruktori',
      'Avtojavob (auto-reply) sozlamalari',
      'Inline menyular',
      'Mavzuli bo‘limlar',
    ],
    settings: ['adminId'],
  },
  {
    id: 'cinema',
    emoji: '🎬',
    category: 'Media',
    name: 'Cinema & Media Botlari',
    short: 'Cinema',
    tagline: 'AniTez / AniXUltra shablonlari',
    description:
      'AniTez va AniXUltra shablonlari asosida media kontent, qidiruv va obuna boshqaruvchi kino botlar.',
    gradient: 'from-sky-400 via-indigo-500 to-violet-500',
    chip: 'border-sky-400/20 bg-sky-500/10 text-sky-300',
    features: [
      '~2ms webhook javob vaqti',
      'Kontent katalogi va qidiruv',
      'Obuna (subscription) tizimi',
      'Broadcast va statistika',
    ],
    settings: ['adminId', 'channelId'],
  },
];

export const SETTING_META: Record<
  SettingKey,
  { label: string; placeholder: string; hint: string; icon: LucideIcon }
> = {
  adminId: {
    label: 'Admin Telegram ID',
    placeholder: 'Masalan: 521348907',
    hint: 'Faqat shu foydalanuvchi botni boshqarishi mumkin.',
    icon: User,
  },
  apiKey: {
    label: 'AI API kaliti (DeepSeek)',
    placeholder: 'sk-...',
    hint: 'DeepSeek API kaliti — AI javoblari shu orqali yaratiladi.',
    icon: KeyRound,
  },
  channelId: {
    label: 'Kanal / Guruh ID',
    placeholder: 'Masalan: @my_channel yoki -1001234567890',
    hint: 'Obuna tekshirish yoki buyurtmalar qabul qilinadigan kanal.',
    icon: Globe,
  },
  currency: {
    label: "Do'kon valyutasi",
    placeholder: 'UZS',
    hint: 'Mahsulot narxlari ko‘rsatiladigan valyuta.',
    icon: Coins,
  },
};

export const CURRENCIES = ['UZS', 'USD', 'RUB', 'EUR'];

export const WIZARD_STEPS = [
  'Shablon tanlang',
  'Bot tokeni va nomi',
  'Bot sozlamalari',
  'Webhookni ulash va ishga tushirish',
];

export const INITIAL_FORM: FormState = {
  template: 'ai-chatbot',
  name: '',
  token: '',
  adminId: '',
  apiKey: '',
  channelId: '',
  currency: 'UZS',
  webhookUrl: '',
  useWebhook: true,
};

/* ---- Demo ma'lumotlar (backend o'chirilganida) ---- */

export const DEMO_BOTS: BotItem[] = [
  {
    id: 'b1',
    name: 'AniTez Poster',
    username: 'anitez_poster_bot',
    template: 'cinema',
    adminId: 'u1',
    running: true,
    webhookActive: true,
    latency: 12,
    requests: 8421,
    aiResponses: 0,
    lastActivity: '2 daqiqa oldin',
  },
  {
    id: 'b2',
    name: "Online Do'kon Bot",
    username: 'shop_uz_bot',
    template: 'ecommerce',
    adminId: 'u2',
    running: true,
    webhookActive: true,
    latency: 8,
    requests: 2145,
    aiResponses: 0,
    lastActivity: '12 daqiqa oldin',
  },
  {
    id: 'b3',
    name: 'AI Yordamchi',
    username: 'deepseek_uz_bot',
    template: 'ai-chatbot',
    adminId: 'u3',
    running: false,
    webhookActive: false,
    latency: 24,
    requests: 1892,
    aiResponses: 3187,
    lastActivity: '3 kun oldin',
  },
];

export const DEMO_STATS: Stats = {
  totalBots: 3,
  activeWebhooks: 2,
  processedRequests: 12458,
  aiResponses: 3187,
  serverLoad: 47,
};

export const DEMO_USERS: PlatformUser[] = [
  {
    id: 'u1',
    name: 'Dilshod Karimov',
    username: 'dilshod_k',
    balance: 150000,
    botCount: 1,
    role: 'admin',
    joinedAt: '2026-07-02T10:00:00Z',
  },
  {
    id: 'u2',
    name: 'Aziza Rahimova',
    username: 'aziza_r',
    balance: 50000,
    botCount: 1,
    role: 'user',
    joinedAt: '2026-07-10T14:20:00Z',
  },
  {
    id: 'u3',
    name: 'Jasur Toshmatov',
    username: 'jasur_t',
    balance: 120000,
    botCount: 1,
    role: 'user',
    joinedAt: '2026-08-01T09:15:00Z',
  },
  {
    id: 'u4',
    name: 'Malika Yusupova',
    username: 'malika_y',
    balance: 0,
    botCount: 0,
    role: 'user',
    joinedAt: '2026-08-18T18:45:00Z',
  },
];

export function demoCheckImage(amount: number): string {
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="640" height="900" viewBox="0 0 640 900">
  <rect width="640" height="900" fill="#0b0b1a"/>
  <rect x="30" y="30" width="580" height="840" rx="28" fill="#14142b" stroke="#2dd4bf" stroke-opacity="0.5" stroke-width="2"/>
  <text x="320" y="120" font-family="monospace" font-size="30" font-weight="bold" fill="#f8fafc" text-anchor="middle">TO'LOV CHEKI</text>
  <text x="320" y="170" font-family="monospace" font-size="18" fill="#94a3b8" text-anchor="middle">PAYMENT RECEIPT</text>
  <line x1="80" y1="210" x2="560" y2="210" stroke="#334155" stroke-width="2" stroke-dasharray="8 8"/>
  <text x="80" y="280" font-family="monospace" font-size="20" fill="#94a3b8">Summa:</text>
  <text x="560" y="280" font-family="monospace" font-size="24" font-weight="bold" fill="#34d399" text-anchor="end">${amount.toLocaleString('en-US')} UZS</text>
  <text x="80" y="360" font-family="monospace" font-size="20" fill="#94a3b8">Karta:</text>
  <text x="560" y="360" font-family="monospace" font-size="20" fill="#e2e8f0" text-anchor="end">8600 •••• •••• 9012</text>
  <text x="80" y="440" font-family="monospace" font-size="20" fill="#94a3b8">Holat:</text>
  <text x="560" y="440" font-family="monospace" font-size="20" fill="#fbbf24" text-anchor="end">Kutilmoqda</text>
  <text x="80" y="520" font-family="monospace" font-size="20" fill="#94a3b8">Vaqt:</text>
  <text x="560" y="520" font-family="monospace" font-size="20" fill="#e2e8f0" text-anchor="end">${new Date().toLocaleString('uz-UZ')}</text>
  <rect x="80" y="600" width="480" height="120" rx="16" fill="#1e293b"/>
  <text x="320" y="668" font-family="monospace" font-size="22" fill="#22d3ee" text-anchor="middle">Backblaze B2 — Screenshot</text>
  <text x="320" y="760" font-family="monospace" font-size="16" fill="#475569" text-anchor="middle">BotMaker AI · Platforma demo cheki</text>
</svg>`;
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

export const DEMO_CHECKS: PaymentCheck[] = [
  {
    id: 'c1',
    userId: 'u3',
    userName: 'Jasur Toshmatov',
    amount: 100000,
    currency: 'UZS',
    screenshotUrl: demoCheckImage(100000),
    status: 'pending',
    createdAt: '2026-08-22T11:30:00Z',
  },
  {
    id: 'c2',
    userId: 'u2',
    userName: 'Aziza Rahimova',
    amount: 50000,
    currency: 'UZS',
    screenshotUrl: demoCheckImage(50000),
    status: 'pending',
    createdAt: '2026-08-22T09:12:00Z',
  },
  {
    id: 'c3',
    userId: 'u3',
    userName: 'Jasur Toshmatov',
    amount: 20000,
    currency: 'UZS',
    screenshotUrl: demoCheckImage(20000),
    status: 'approved',
    createdAt: '2026-08-20T16:40:00Z',
  },
  {
    id: 'c4',
    userId: 'u1',
    userName: 'Dilshod Karimov',
    amount: 50000,
    currency: 'UZS',
    screenshotUrl: demoCheckImage(50000),
    status: 'rejected',
    createdAt: '2026-08-19T13:05:00Z',
  },
];

export const DEFAULT_CARD: CardSettings = {
  cardNumber: '8600 1234 5678 9012',
  cardHolder: 'BOTMAKER AI ADMIN',
  bank: 'Uzcard',
};

/* ---- Formatlash ---- */

export function fmtNum(n: number): string {
  return n.toLocaleString('en-US');
}

export function fmtMoney(n: number): string {
  return `💰 ${fmtNum(n)} UZS`;
}

const compactFmt = new Intl.NumberFormat('en', { notation: 'compact', maximumFractionDigits: 1 });

export function fmtCompact(n: number): string {
  return compactFmt.format(n);
}

export function slugify(s: string): string {
  return s.toLowerCase().replace(/[^a-z0-9]+/g, '-').replace(/^-+|-+$/g, '') || 'bot';
}

export function tplOf(id: TemplateId): Template {
  return TEMPLATES.find((t) => t.id === id) ?? TEMPLATES[0];
}

export function num(v: unknown): number {
  const n = Number(v);
  return Number.isFinite(n) ? n : 0;
}

/* Backend shablon nomini frontend TemplateId ga moslashtirish */
export function mapTemplate(raw: string): TemplateId {
  const s = raw.toLowerCase();
  if (s.includes('anitez') || s.includes('anix') || s.includes('cinema') || s.includes('media')) return 'cinema';
  if (s.includes('ai') || s.includes('deepseek') || s.includes('chatbot')) return 'ai-chatbot';
  if (s.includes('shop') || s.includes('store') || s.includes('ecom') || s.includes('commerce')) return 'ecommerce';
  if (s.includes('support') || s.includes('feedback')) return 'feedback';
  if (s.includes('channel') || s.includes('obuna') || s.includes('subscribe')) return 'channel';
  if (s.includes('menu') || s.includes('visual')) return 'visual-menu';
  return 'cinema';
}

export function normalizeStats(raw: unknown): Stats {
  const r = (raw ?? {}) as Record<string, unknown>;
  const loadRaw = r.server_load ?? r.serverLoad ?? r.cpu ?? r.load;
  const loadN = Number(loadRaw);
  return {
    totalBots: num(r.total_bots ?? r.totalBots ?? r.bots),
    activeWebhooks: num(r.active_webhooks ?? r.activeWebhooks ?? r.webhooks),
    processedRequests: num(r.processed_requests ?? r.processedRequests ?? r.requests),
    aiResponses: num(r.ai_responses ?? r.aiResponses ?? r.ai_answers),
    serverLoad: Number.isFinite(loadN) ? Math.min(100, Math.max(0, loadN)) : null,
  };
}

export function normalizeBots(raw: unknown): BotItem[] {
  const list = Array.isArray(raw) ? raw : ((raw as Record<string, unknown>)?.bots as unknown[] | undefined) ?? [];
  return list.map((b, i) => {
    const row = (b ?? {}) as Record<string, unknown>;
    const webhook = row.webhook as Record<string, unknown> | undefined;
    const lastRaw = row.last_activity ?? row.updated_at;
    const latencyRaw = row.latency ?? webhook?.latency_ms;
    return {
      id: String(row.id ?? row.bot_id ?? row.botId ?? `bot-${i}`),
      name: String(row.name ?? row.bot_name ?? 'Nomsiz bot'),
      username: row.username ? String(row.username) : undefined,
      template: mapTemplate(String(row.template ?? 'cinema')),
      token: row.token ? String(row.token) : undefined,
      adminId: row.admin_id ?? row.adminId ? String(row.admin_id ?? row.adminId) : undefined,
      apiKey: row.api_key ?? row.apiKey ? String(row.api_key ?? row.apiKey) : undefined,
      channelId: row.channel_id ?? row.channelId ? String(row.channel_id ?? row.channelId) : undefined,
      currency: row.currency ? String(row.currency) : undefined,
      webhookUrl: row.webhook_url ?? row.webhookUrl ? String(row.webhook_url ?? row.webhookUrl) : undefined,
      running: Boolean(row.running ?? row.is_running ?? row.active),
      webhookActive: Boolean(row.webhook_active ?? row.webhookActive ?? webhook?.active ?? true),
      latency: latencyRaw != null ? num(latencyRaw) : undefined,
      requests: row.requests ?? row.webhook_requests ? num(row.requests ?? row.webhook_requests) : undefined,
      aiResponses: row.ai_responses ?? row.aiResponses ? num(row.ai_responses ?? row.aiResponses) : undefined,
      lastActivity: lastRaw ? String(lastRaw) : undefined,
    };
  });
}

/* Avatarlar: ism bosh harflaridan SVG data-URL generatsiya */
export function initialsAvatar(name: string): string {
  const clean = name.trim() || '?';
  const initials =
    clean
      .split(/\s+/)
      .slice(0, 2)
      .map((p) => p[0]?.toUpperCase() ?? '')
      .join('') || '?';
  const hue = [...clean].reduce((a, ch) => a + ch.charCodeAt(0), 0) % 360;
  const svg = `<svg xmlns="http://www.w3.org/2000/svg" width="96" height="96"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop offset="0" stop-color="hsl(${hue},85%,55%)"/><stop offset="1" stop-color="hsl(${(hue + 60) % 360},85%,45%)"/></linearGradient></defs><rect width="96" height="96" rx="24" fill="url(#g)"/><text x="48" y="61" font-family="Inter,Arial,sans-serif" font-size="38" font-weight="700" fill="white" text-anchor="middle">${initials}</text></svg>`;
  return `data:image/svg+xml;utf8,${encodeURIComponent(svg)}`;
}

export function timeAgo(iso: string): string {
  const then = new Date(iso).getTime();
  if (!Number.isFinite(then)) return '—';
  const diff = Date.now() - then;
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'hozir';
  if (mins < 60) return `${mins} daqiqa oldin`;
  const hours = Math.floor(mins / 60);
  if (hours < 24) return `${hours} soat oldin`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days} kun oldin`;
  return new Date(iso).toLocaleDateString('uz-UZ');
}
