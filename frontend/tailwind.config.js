/** @type {import('tailwindcss').Config} */
export default {
  content: ['./index.html', './src/**/*.{js,ts,jsx,tsx}'],
  theme: {
    extend: {
      colors: {
        primary: '#0F0F12',
        card: '#1A1A20',
        hover: '#242430',
        border: '#2A2A38',
        accent: {
          blue: '#4F8EF7',
          green: '#3ECF8E',
          red: '#E5534B',
        },
        text: {
          primary: '#F0F0F5',
          muted: '#8888AA',
        },
      },
    },
  },
  plugins: [],
}
