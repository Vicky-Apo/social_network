/** @type {import('tailwindcss').Config} */
module.exports = {
  content: [
    "./src/app/**/*.{js,ts,jsx,tsx}",
    "./src/pages/**/*.{js,ts,jsx,tsx}",
    "./src/components/**/*.{js,ts,jsx,tsx}",
  ],
  theme: {
    extend: {
      colors: {
        brand: {
          primary: "#00584b",
          accent: "#d8e267",
          contrast: "#f6d8f7",
          light: "#ffffff",
        },
      },
    },
  },
  plugins: [],
};
