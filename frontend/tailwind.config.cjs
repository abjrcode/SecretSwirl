/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  darkMode: 'media',
  theme: {
    debugScreens: {
      prefix: 'dbg-screen: ',
      position: ['bottom', 'right'],
    },
    extend: {
      colors: {
        'skin-primary': 'rgb(var(--color-fill) / <alpha-value>)',
        'skin-secondary': 'rgb(var(--color-text) / <alpha-value>)',
      },
    },
  },
  plugins: [require('tailwindcss-debug-screens')],
}
