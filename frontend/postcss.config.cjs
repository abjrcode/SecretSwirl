module.exports = {
  plugins: {
    "postcss-functions": {
      functions: {
        toRgb(hex) {
          let hexValue = hex.replace("#", "");

          if (hexValue.length === 3) {
            hexValue = hexValue.replace(/./g, "$&$&");
          }

          const r = parseInt(hexValue.substring(0, 2), 16);
          const g = parseInt(hexValue.substring(2, 4), 16);
          const b = parseInt(hexValue.substring(4, 6), 16);
          return `${r} ${g} ${b}`;
        },
      },
    },
    tailwindcss: {},
    autoprefixer: {},
  },
};
