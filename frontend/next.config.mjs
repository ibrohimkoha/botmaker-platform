/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Ilova Nginx'da /botmaker/ marshruti ostida xizmat qiladi — assetlar va sahifalar shu yo'ldan yuklanadi
  basePath: '/botmaker',
  // Backend API'ga proxy: /api/* -> http://localhost:8085/api/* (CORS muammosini oldini oladi)
  // basePath o'rnatilganda Next.js rewrite source'iga avtomatik prefiks qo'shadi: /botmaker/api/* -> backend
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8085/api/:path*',
      },
    ];
  },
};

export default nextConfig;
