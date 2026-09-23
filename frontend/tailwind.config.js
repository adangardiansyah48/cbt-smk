/** @type {import('tailwindcss').Config} */
export default {
  content: ["./index.html", "./src/**/*.{vue,ts,tsx,js}"],
  theme: {
    extend: {
      colors: { brand: "#0d9488", surface: "#0f172a" },
      fontFamily: { sans: ["Segoe UI", "Inter", "system-ui", "sans-serif"] },
    },
  },
  plugins: [require("@tailwindcss/typography")],
}
