/** @type {import('next').NextConfig} */
const nextConfig = {
  reactStrictMode: true,
  // Backend API'ga proxy: /api/* -> http://localhost:8085/api/* (CORS muammosini oldini oladi)
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
