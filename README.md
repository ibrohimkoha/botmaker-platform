# 🤖 Telegram BotMaker AI Platform

Yuqori unumdorlikka ega, arzon va zamonaviy **Telegram Botmaker** platformasi.
Go Webhook Engine va Next.js 14 Dashboard asosida yaratilgan.

---

## 🌟 Imkoniyatlar

- ⚡ **Yagona Go Webhook Dispatcher:** Barcha botlar 1 ta HTTP server orqali `1-3ms` latensiyada ishlaydi.
- 🎨 **Zamonaviy Neon Dashboard:** Next.js 14 + Tailwind CSS + Lucide Icons.
- 📦 **Mavjud Bot Shablonlari:**
  - `AniTez` — Tezkor Anime & Kino qidirish boti (kod, nom, top ro'yxat, obuna tekshiruvi).
  - `AniXUltra` — VIP & Seriallar Cinema boti (kategoriyalar, epizodlar, admin boshqaruvi).
- 🚀 **Boshqaruv & Analitika:**
  - 1-tugma bilan bot yaratish va Webhook ulash (`setWebhook`).
  - Botlarni yoqish/to'xtatish (Start/Stop).
  - Ommaviy xabar yuborish (Broadcast).
  - Real-vaqtli statistika (Foydalanuvchilar, qidiruvlar, so'rovlar).

---

## 🛠 Texnologik Stack

- **Backend:** Go (Golang 1.25), `gopkg.in/telebot.v3`, SQLite/GORM.
- **Frontend:** Next.js 14, React 18, Tailwind CSS, TypeScript.
- **Infratuzilma:** Webhook Routing, Nginx Proxy, Systemd / PM2.
