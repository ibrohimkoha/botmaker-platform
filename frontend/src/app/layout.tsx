import type { Metadata, Viewport } from 'next';
import './globals.css';

export const metadata: Metadata = {
  title: '🤖 BotMaker AI — Professional Telegram Botlar Konstruktori',
  description:
    'Bozor narxidan 10x arzonroq, yuqori sifatli va Webhook asosida chaqmoqdek tez ishlovchi Telegram botlar yarating — AI chatbot, e-commerce, feedback, obuna menejeri va boshqa shablonlar.',
};

export const viewport: Viewport = {
  themeColor: '#05050f',
  width: 'device-width',
  initialScale: 1,
};

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="uz">
      <head>
        <link rel="preconnect" href="https://fonts.googleapis.com" />
        <link rel="preconnect" href="https://fonts.gstatic.com" crossOrigin="anonymous" />
        <link
          href="https://fonts.googleapis.com/css2?family=Inter:wght@400;500;600;700;800&display=swap"
          rel="stylesheet"
        />
      </head>
      <body className="min-h-screen antialiased">{children}</body>
    </html>
  );
}
