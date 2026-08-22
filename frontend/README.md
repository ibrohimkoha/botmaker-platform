# BotMaker AI — Frontend Dashboard

Telegram Botmaker platformasi uchun Next.js + Tailwind CSS asosidagi dashboard.

## Ishga tushirish

```bash
npm install
npm run dev
```

Dashboard `http://localhost:3000` da ochiladi.

## Backend API

Frontend avval Next.js proksi orqali (`/api/*` → `http://localhost:8085/api/*`),
bo‘lmasa to‘g‘ridan-to‘g‘ri `http://localhost:8085/api` ga ulanadi.

Boshqa manzil kerak bo‘lsa:

```bash
NEXT_PUBLIC_API_BASE=http://host:port npm run dev
```

Backend topilmasa dashboard demo ma’lumotlar bilan ishlaydi va bu haqida banner ko‘rsatadi.

## Kutayotgan API endpointlar

| Method | Endpoint               | Tavsif                          |
| ------ | ---------------------- | ------------------------------- |
| GET    | /api/stats             | Statistika                      |
| GET    | /api/bots              | Botlar ro‘yxati                 |
| POST   | /api/bots              | Bot yaratish                    |
| POST   | /api/bots/:id/start    | Botni ishga tushirish           |
| POST   | /api/bots/:id/stop     | Botni to‘xtatish                |
| DELETE | /api/bots/:id          | Botni o‘chirish                 |
| POST   | /api/broadcast         | Broadcast xabar yuborish        |

## Build

```bash
npm run build
```
