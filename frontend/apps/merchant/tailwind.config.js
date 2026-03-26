import daisyui from 'daisyui'

/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{vue,ts,tsx}'],
  theme: {
    extend: {},
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
