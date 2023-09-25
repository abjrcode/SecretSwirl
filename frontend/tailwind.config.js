/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./index.html", "./src/**/*.{js,ts,jsx,tsx}"],
  theme: {
    debugScreens: {
      prefix: "dbg-screen: ",
      position: ["bottom", "right"],
    },
    extend: {},
  },
  plugins: [require("tailwindcss-debug-screens")],
};
