/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.tsx"],
  daisyui: {
    themes: [
      "light",
      "dark",
      "cupcake",
      "corporate",
      "bumblebee",
      "emerald",
      "synthwave",
      "retro",
      "cyberpunk",
      "valentine",
      "halloween",
      "garden",
      "forest",
      "aqua",
      "lofi",
      "pastel",
      "fantasy",
      "wireframe",
      "black",
      "luxury",
      "dracula",
      "lemonade",
    ],
  },
  darkMode: "media",
  theme: {
    debugScreens: {
      prefix: "dbg-screen: ",
      position: ["bottom", "right"],
    },
    extend: {
      colors: {
        "skin-primary": "rgb(var(--color-fill) / <alpha-value>)",
        "skin-secondary": "rgb(var(--color-text) / <alpha-value>)",
      },
    },
  },
  plugins: [require("daisyui"), require("tailwindcss-debug-screens")],
}
