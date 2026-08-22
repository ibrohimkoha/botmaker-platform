# 🤖 Universal Telegram BotMaker Platformasi — Project Status & Report

> Ushbu hujjat platformaning to'liq arxitekturasi, barcha bajarilgan ishlar, server va baza sozlamalari, API yo'nalishlari hamda yangi chatda loyihani muammosiz davom ettirish uchun to'liq hisobotni o'z ichiga oladi.

---

## 🌐 1. Jonli Manzillar va Havolalar

* **Frontend Dashboard (SaaS):** `https://nokori-uz.duckdns.org/botmaker`
* **Backend REST API:** `https://nokori-uz.duckdns.org/botmaker-api`
* **Telegram Webhook URL:** `https://nokori-uz.duckdns.org/webhook/<bot_token>`
* **GitHub Repozitoriy:** `https://github.com/ibrohimkoha/botmaker-platform.git`
* **Telegram Auth Bot:** `@botmakerauthbot` (`https://t.me/botmakerauthbot`)

---

## 🖥️ 2. Server va Infratuzilma (VPS)

* **Server IP:** `157.173.110.5` (Ubuntu Linux)
* **SSH Foydalanuvchi:** `root`
* **Loyiha Joylashuvi (VPS):** `/root/botmaker-platform`
* **Loyiha Joylashuvi (Local):** `/home/iskurama/Desktop/botmaker-platform`

### ⚡ Portlar va PM2 Jarayonlari:
| PM2 ID | Jarayon Nomi | Port | Vazifasi / Texnologiya |
|---|---|---|---|
| **39** | `botmaker-backend` | `8085` | Go 1.22+ yuqori tezlikdagi Webhook routeri va REST API |
| **42** | `botmaker-frontend` | `3050` | Next.js 14 Dashboard (`basePath: '/botmaker'`) |
| **35** | `nokori-go-web` | `3001` | Asosiy portal (tegilmagan) |
| **10** | `femutsu` | `3000` | Asosiy veb sayt (tegilmagan) |

> ⚠️ **Muhim Eslatma:** Port `3005` serverdagi boshqa Python xizmati tomonidan band. BotMaker Frontend porti: `3050`.

### 🛡️ Nginx Konfiguratsiyasi (`/etc/nginx/conf.d/duckdns.conf`):
* `/botmaker` -> `http://127.0.0.1:3050/botmaker` (Next.js)
* `/botmaker-api/` -> `http://127.0.0.1:8085/api/` (Go Backend)
* `/webhook/` -> `http://127.0.0.1:8085/webhook/` (Telegram Webhook Updates)

---

## 🔑 3. Kredensiallar va Tashqi Servislar

### 1. Google OAuth 2.0:
* **Client ID:** `531444252311-44j9qjp0ek0jbi5eggbmj9nrl317au54.apps.googleusercontent.com`
* **Client Secret:** `VPS ecosystem.config.js da sozlangan` (GOCSPX-***)
* **Authorized Origins:** `https://nokori-uz.duckdns.org`
* **Authorized Redirect URI:** `https://nokori-uz.duckdns.org/botmaker/api/auth/callback/google`

### 2. Telegram Auth:
* **Auth Bot Token:** `VPS ecosystem.config.js da sozlangan` (8849160724:***)
* **Auth Bot Username:** `@botmakerauthbot`
* **SuperAdmin Telegram ID:** `5415350162` (Iskurama) / `6149114216`
* **SuperAdmin Email:** `miraxmedovibrohim5@gmail.com`

### 3. Backblaze B2 Cloud Storage (Cheklarni saqlash):
* **Bucket Nomi:** `iskurama`
* **Bucket ID:** `f5c6d20ee6cb124baa0f0d14`
* **KeyID:** `005562e6b2bafd40000000002`
* **ApplicationKey:** `VPS ecosystem.config.js da sozlangan` (K005***)
* **Endpoint:** `s3.us-east-005.backblazeb2.com`
* **Region:** `us-east-005`

### 4. DeepSeek AI (Botlar uchun):
* **Base URL:** `https://api.deepseek.com`
* **API Key:** `VPS ecosystem.config.js da sozlangan` (sk-***)
* **Model:** `deepseek-chat` / `deepseek-v4-flash`

---

## 🗄️ 4. Ma'lumotlar Bazasi Strukturasi (SQLite — `botmaker.db`)

1. **`users`** — Platforma foydalanuvchilari (`id`, `telegram_id`, `google_id`, `email`, `full_name`, `avatar_url`, `balance`, `role`, `created_at`, `updated_at`).
2. **`bots`** — Yaratilgan Telegram botlar (`id`, `name`, `username`, `token`, `template`, `webhook_url`, `status`, `created_at`, `updated_at`).
3. **`bot_settings`** — Bot parametrlari (`bot_id`, `key` [admin_id, channel_id, api_key, currency, welcome_message], `value`, `updated_at`).
4. **`deposits`** — Balans to'ldirish cheklari (`id`, `user_id`, `amount`, `receipt_url` [B2 URL], `status` [pending/approved/rejected], `reject_reason`, `created_at`, `approved_at`).
5. **`payment_cards`** — To'lov qabul qiluvchi admin kartalari (`id`, `card_number`, `card_holder`, `bank_name`, `is_active`, `created_at`).
6. **`custom_repo_templates`** — Dinamik Git repo shablonlari (`id`, `name`, `title`, `description`, `git_repo_url`, `category`, `price`, `created_at`).
7. **`broadcasts`, `stats`, `orders`, `tickets`** — Bot statistikasi, xabarlar, e-commerce va support jadvallari.

> 🛡️ **Xavfsizlik Qoidasi:** Serverdagi asosiy PostgreSQL `nokori_go` bazasiga va mavjud anime/kino botlariga mutlaqo tegish taqiqlangan. BotMaker to'liq mustaqil SQLite bazada ishlaydi.

---

## 🚀 5. Bajarilgan Barcha Vazifalar (Changelog v1.0 — v3.1)

1. **Multi-Agent Tizimi:** DeepSeek V4-Flash orqali 2 ta ixtisoslashgan agent (Agent 1: Go Backend, Agent 2: Next.js Frontend) orqali arzon va tezkor ishlab chiqildi.
2. **Go Webhook Engine:** Telegram botlarini polling o'rniga chaqmoqdek tez Webhook orqali ishlash mexanizmi (~1-3ms javob tezligi).
3. **Mavjud Bot Shablonlari:**
   - 🤖 `ai_assistant` (DeepSeek AI bilan aqlli suhbat)
   - 🛒 `ecommerce_shop` (Savatcha, mahsulotlar va buyurtmalar)
   - 💬 `feedback_support` (2 tomonlama admin-mijoz chati)
   - 📢 `channel_manager` (Kanalga a'zolikni majburiy tekshirish)
   - 🧩 `custom_builder` (Visual tugmalar va javoblar konstruktori)
   - 🎬 `anitez` va `anixultra` (Media va kino botlari)
4. **Foydalanuvchi va Balans Tizimi (User Panel):**
   - Google 1-Click va Telegram `@botmakerauthbot` orqali xavfsiz kirish.
   - Bot yaratish uchun tizimga kirish majburiyligi.
   - Shaxsiy balans (`💰 ... UZS`), Uzcard/Humo karta raqamiga pul o'tkazib Backblaze B2 ga to'lov chekini yuklash va admin tasdig'iga jo'natish.
5. **Administrator Paneli (Admin Panel):**
   - **Cheklar nazorati:** Rasm screenshotlarini ko'rish, 1-tugma bilan "Tasdiqlash" (foydalanuvchi balansiga avtomatik pul qo'shadi) yoki "Rad etish".
   - **Karta sozlamalari:** Admin o'z karta raqamini istalgan payt o'zgartira oladi.
   - **Dinamik Repo Shablonlar:** Admin yangi bot shablonlarini Git repo orqali qo'sha oladi.
   - **Foydalanuvchilar va Botlar monitoringi.**
6. **Xatoliklar bartaraf etildi:**
   - Telegram Webhook 400 xatosi (HTTPS domen tekshiruvi orqali to'g'rilandi).
   - Bot sozlamalaridagi 405 xatosi (PUT/PATCH sozlamalar handlerlari qo'shildi).
   - Broadcast 400 xatosi (moslashuvchan int/string/all payload orqali tuzatildi).
   - Bot sozlamalarini qayta ochganda avtomatik yuklash va to'ldirish ta'minlandi.

---

## 🔮 6. Keyingi Bosqichlar (Yangi Chatda Qilinishi Mumkin Bo'lgan Ishlar):

1. **Click / Payme Avtomatik To'lov Integratsiyasi:** Chek yuklash bilan bir qatorda Click / Payme orqali to'g'ridan-to'g'ri avtomatik hisob to'ldirish.
2. **Yangi Shablonlarni Ishga Tushirish Dvigateli:** Git repo orqali qo'shilgan yangi repozitoriylarni dinamik `git clone` qilib, konteynerda ishga tushirish mexanizmi.
3. **Foydalanuvchi SMS / Bot Bildirishnomalari:** To'lov tasdiqlanganda foydalanuvchining Telegramiga bildirishnoma yuborish.

---

## 🛠️ Loyihani Qayta Ishga Tushirish / Yangilash Buyruqlari

```bash
# 1. VPS ga ulanish va kodni yangilash:
ssh root@157.173.110.5
cd /root/botmaker-platform && git pull origin main

# 2. Go Backendni yig'ish va qayta ishga tushirish:
cd /root/botmaker-platform/backend
go build -o bin/botmaker-server ./cmd/server
pm2 restart botmaker-backend

# 3. Next.js Frontendni yig'ish va qayta ishga tushirish:
cd /root/botmaker-platform/frontend
npm run build
pm2 restart botmaker-frontend
```
