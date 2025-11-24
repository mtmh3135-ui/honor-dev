/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./index.html",
    "./src/**/*.{js,ts,jsx,tsx}", // 👈 penting agar Tailwind scan semua file
  ],
  theme: {
    extend: {},
  },
  plugins: [],
}
export default {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    extend: {
      colors: {
        softgreen: "#92E3A9",
      },
    },
  },
  plugins: [],
};