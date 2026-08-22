module.exports = {
  apps: [
    {
      name: 'botmaker-backend',
      cwd: '/root/botmaker-platform/backend',
      script: './bin/botmaker-server',
      env: {
        PORT: 8085,
        PUBLIC_URL: 'https://nokori-uz.duckdns.org',
        DB_PATH: '/root/botmaker-platform/backend/botmaker.db'
      },
      restart_delay: 3000,
      autorestart: true
    },
    {
      name: 'botmaker-frontend',
      cwd: '/root/botmaker-platform/frontend',
      script: 'node_modules/next/dist/bin/next',
      args: 'start -p 3005',
      env: {
        NODE_ENV: 'production',
        PORT: 3005,
        NEXT_PUBLIC_API_BASE: 'https://nokori-uz.duckdns.org/botmaker-api'
      },
      restart_delay: 3000,
      autorestart: true
    }
  ]
};
