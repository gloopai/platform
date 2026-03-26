import daisyui from 'daisyui'

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx}'],
  theme: {
    extend: {
      colors: {
        brand: {
          50: '#f0f9ff',
          100: '#e0f2fe',
          200: '#bae6fd',
          300: '#7dd3fc',
          400: '#38bdf8',
          500: '#0ea5e9',
          600: '#0284c7',
          700: '#0369a1',
          800: '#075985',
          900: '#0c4a6e',
        },
      },
      fontFamily: {
        sans: ['Inter', 'ui-sans-serif', 'system-ui', 'sans-serif'],
      },
      backgroundImage: {
        'grid-slate': `linear-gradient(to right, rgb(15 23 42 / 0.06) 1px, transparent 1px),
          linear-gradient(to bottom, rgb(15 23 42 / 0.06) 1px, transparent 1px)`,
      },
      boxShadow: {
        glow: '0 0 80px -20px rgb(14 165 233 / 0.35)',
      },
    },
  },
  plugins: [daisyui],
  daisyui: {
    themes: [
      {
        gloopmono: {
          primary: '#111111',
          'primary-content': '#ffffff',
          secondary: '#2a2a2a',
          'secondary-content': '#ffffff',
          accent: '#3a3a3a',
          'accent-content': '#ffffff',
          neutral: '#1f1f1f',
          'neutral-content': '#f5f5f5',
          'base-100': '#ffffff',
          'base-200': '#f7f7f7',
          'base-300': '#ececec',
          'base-content': '#111111',
          info: '#3b82f6',
          success: '#16a34a',
          warning: '#d97706',
          error: '#dc2626',
        },
      },
    ],
    darkTheme: 'gloopmono',
    logs: false,
  },
}
