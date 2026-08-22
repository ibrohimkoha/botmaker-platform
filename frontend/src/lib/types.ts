/* BotMaker AI SaaS platformasi — umumiy turlar */

export type TemplateId =
  | 'ai-chatbot'
  | 'ecommerce'
  | 'feedback'
  | 'channel'
  | 'visual-menu'
  | 'cinema';

export type SettingKey = 'adminId' | 'apiKey' | 'channelId' | 'currency';

export type Role = 'user' | 'admin';

export type AuthMethod = 'google' | 'telegram' | 'quick';

export type CheckStatus = 'pending' | 'approved' | 'rejected';

export interface Template {
  id: TemplateId;
  emoji: string;
  category: string;
  name: string;
  short: string;
  tagline: string;
  description: string;
  gradient: string;
  chip: string;
  features: string[];
  settings: SettingKey[];
}

export interface BotItem {
  id: string;
  name: string;
  username?: string;
  template: TemplateId;
  token?: string;
  adminId?: string;
  apiKey?: string;
  channelId?: string;
  currency?: string;
  webhookUrl?: string;
  running: boolean;
  webhookActive: boolean;
  latency?: number;
  requests?: number;
  aiResponses?: number;
  lastActivity?: string;
}

export interface Stats {
  totalBots: number;
  activeWebhooks: number;
  processedRequests: number;
  aiResponses: number;
  serverLoad: number | null;
}

export interface ToastMsg {
  id: number;
  type: 'success' | 'error' | 'info';
  message: string;
}

/* Ro'yxatdan o'tgan foydalanuvchi (session) */
export interface UserProfile {
  id: string;
  name: string;
  email?: string;
  username?: string;
  avatar?: string;
  balance: number;
  role: Role;
  authMethod: AuthMethod;
  joinedAt?: string;
}

/* Admin Panel — Foydalanuvchilar tab */
export interface PlatformUser {
  id: string;
  name: string;
  username?: string;
  avatar?: string;
  balance: number;
  botCount: number;
  role: Role;
  joinedAt?: string;
}

/* Admin Panel — To'lov cheklari tab */
export interface PaymentCheck {
  id: string;
  userId?: string;
  userName: string;
  amount: number;
  currency: string;
  screenshotUrl?: string;
  status: CheckStatus;
  createdAt: string;
  note?: string;
}

/* Admin Panel — Karta sozlamalari tab */
export interface CardSettings {
  cardNumber: string;
  cardHolder: string;
  bank?: string;
}

/* Admin Panel — Git repozitoriy asosidagi shablon */
export interface AdminTemplate {
  id: string;
  name: string;
  repoUrl: string;
  price: number;
  category: string;
  description?: string;
  createdAt: string;
}

/* Bot sozlamalari (GET/PUT /api/bots/:id/settings) */
export interface BotSettings {
  adminId: string;
  apiKey: string;
  channelId: string;
  currency: string;
  webhookUrl: string;
}

export interface FormState {
  template: TemplateId;
  name: string;
  token: string;
  adminId: string;
  apiKey: string;
  channelId: string;
  currency: string;
  webhookUrl: string;
  useWebhook: boolean;
}
