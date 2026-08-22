import type { Config } from 'tailwindcss';

const config: Config = {
  content: ['./src/**/*.{js,ts,jsx,tsx,mdx}'],
  theme: {
    extend: {
      fontFamily: {
        sans: [
          'Inter',
          'ui-sans-serif',
          'system-ui',
          '-apple-system',
          'Segoe UI',
          'Roboto',
          'Helvetica Neue',
          'Arial',
          'sans-serif',
        ],
        mono: ['ui-monospace', 'SFMono-Regular', 'Menlo', 'Monaco', 'Consolas', 'monospace'],
      },
      colors: {
        night: {
          DEFAULT: '#05050f',
          soft: '#0b0b1a',
        },
        neon: {
          cyan: '#22d3ee',
          violet: '#a855f7',
          fuchsia: '#e879f9',
          pink: '#f472b6',
        },
      },
      boxShadow: {
        glow: '0 0 24px rgba(139, 92, 246, 0.25)',
        'glow-lg': '0 0 48px rgba(139, 92, 246, 0.35)',
        'glow-cyan': '0 0 24px rgba(34, 211, 238, 0.25)',
      },
      keyframes: {
        'pulse-soft': {
          '0%, 100%': { opacity: '1' },
          '50%': { opacity: '0.45' },
        },
        shimmer: {
          '0%': { backgroundPosition: '-200% 0' },
          '100%': { backgroundPosition: '200% 0' },
        },
      },
      animation: {
        'pulse-soft': 'pulse-soft 2.4s ease-in-out infinite',
        shimmer: 'shimmer 2.2s linear infinite',
      },
    },
  },
  plugins: [],
};

export default config;
