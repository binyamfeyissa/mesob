import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./src/pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/components/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/app/**/*.{js,ts,jsx,tsx,mdx}",
    "./src/features/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  theme: {
    extend: {
      colors: {
        // Mesob design tokens
        mesob: {
          blue: "#1B4FDE",
          gold: "#E6A817",
          dark: "#0F1B3D",
        },
      },
    },
  },
  plugins: [],
};

export default config;
