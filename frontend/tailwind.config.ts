import type { Config } from "tailwindcss";

const config: Config = {
  content: [
    "./pages/**/*.{js,ts,jsx,tsx,mdx}",
    "./components/**/*.{js,ts,jsx,tsx,mdx}",
    "./app/**/*.{js,ts,jsx,tsx,mdx}",
  ],
  darkMode: "class",
  theme: {
    extend: {
      colors: {
        // Asana-inspired dark theme colors
        background: {
          DEFAULT: "#1e1f21",
          secondary: "#252628",
          tertiary: "#2e2f31",
          hover: "#353638",
        },
        foreground: {
          DEFAULT: "#f5f4f3",
          secondary: "#a2a0a2",
          tertiary: "#6d6e6f",
        },
        // Accent colors (Asana coral/salmon)
        accent: {
          DEFAULT: "#f06a6a",
          hover: "#e85555",
          light: "#ffeae9",
        },
        // Success/Info colors
        success: {
          DEFAULT: "#5da283",
          light: "#e7f4ef",
        },
        warning: {
          DEFAULT: "#f1bd6c",
          light: "#fef6e9",
        },
        error: {
          DEFAULT: "#e8384f",
          light: "#fdecef",
        },
        info: {
          DEFAULT: "#4573d2",
          light: "#eaf0fc",
        },
        // Border colors
        border: {
          DEFAULT: "#424244",
          light: "#565758",
          focus: "#8f8f91",
        },
        // Sidebar colors
        sidebar: {
          DEFAULT: "#2e2f31",
          hover: "#353638",
          active: "#3d3e40",
        },
      },
      fontFamily: {
        sans: ['"DM Sans"', 'system-ui', 'sans-serif'],
        mono: ['"JetBrains Mono"', 'monospace'],
      },
      boxShadow: {
        'card': '0 1px 3px rgba(0, 0, 0, 0.3)',
        'card-hover': '0 4px 12px rgba(0, 0, 0, 0.4)',
        'modal': '0 8px 24px rgba(0, 0, 0, 0.5)',
      },
      animation: {
        'fade-in': 'fadeIn 0.2s ease-out',
        'slide-in': 'slideIn 0.3s ease-out',
        'scale-in': 'scaleIn 0.2s ease-out',
      },
      keyframes: {
        fadeIn: {
          '0%': { opacity: '0' },
          '100%': { opacity: '1' },
        },
        slideIn: {
          '0%': { transform: 'translateY(-10px)', opacity: '0' },
          '100%': { transform: 'translateY(0)', opacity: '1' },
        },
        scaleIn: {
          '0%': { transform: 'scale(0.95)', opacity: '0' },
          '100%': { transform: 'scale(1)', opacity: '1' },
        },
      },
    },
  },
  plugins: [],
};
export default config;
