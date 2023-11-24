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
    logs: false,
  },
  darkMode: "media",
  theme: {
    debugScreens: {
      prefix: "dbg-screen: ",
      position: ["bottom", "right"],
    },
    extend: {
      transitionProperty: {
        "max-height": "max-height",
      },
      keyframes: {
        wiggle: {
          "0%, 100%": { transform: "rotate(-3deg)" },
          "50%": { transform: "rotate(3deg)" },
        },
        shake: {
          "0%": {
            transform: "translateX(0)",
          },
          "25%": {
            transform: "translateX(-5px)",
          },
          "50%": {
            transform: "translateX(5px)",
          },
          "75%": {
            transform: "translateX(-5px)",
          },
          "100%": {
            transform: "translateX(5px)",
          },
        },
        slideInLeft: {
          from: {
            transform: "translate3d(-100%, 0, 0)",
          },
          to: {
            transform: "translate3d(0, 0, 0)",
          },
        },
        slideOutLeft: {
          from: {
            transform: "translate3d(0, 0, 0)",
          },
          to: {
            transform: "translate3d(-110%, 0, 0)",
          },
        },
        slideInRight: {
          from: {
            transform: "translate3d(100%, 0, 0)",
          },
          to: {
            transform: "translate3d(0, 0, 0)",
          },
        },
        slideOutRight: {
          from: {
            transform: "translate3d(0, 0, 0)",
          },
          to: {
            transform: "translate3d(110%, 0, 0)",
          },
        },
      },
      animation: {
        wiggle: "wiggle 1s ease-in-out infinite",
        shake: "shake 0.250s ease-in-out 1",
      },
      colors: {
        "skin-primary": "rgb(var(--color-fill) / <alpha-value>)",
        "skin-secondary": "rgb(var(--color-text) / <alpha-value>)",
      },
    },
  },
  plugins: [require("daisyui"), require("tailwindcss-debug-screens")],
}
