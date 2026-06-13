export const colors = {
  // Backgrounds
  bg: "#F2F3F7",
  surface: "#FFFFFF",
  surfaceSecondary: "#F8F8FB",

  // Brand
  primary: "#3B6CF8",
  primaryLight: "rgba(59,108,248,0.12)",
  navy: "#0F1B3D",
  gold: "#E6A817",

  // Text
  text: "#111827",
  textSecondary: "#6B7280",
  textTertiary: "#9CA3AF",

  // Semantic
  success: "#059669",
  successLight: "#D1FAE5",
  error: "#DC2626",
  errorLight: "#FEE2E2",
  warning: "#F59E0B",
  warningLight: "#FEF3C7",

  // UI
  border: "#E5E7EB",
  divider: "#F3F4F6",
  shadow: "#000000",
  overlay: "rgba(0,0,0,0.5)",
};

export const spacing = {
  xs: 4,
  sm: 8,
  md: 16,
  lg: 24,
  xl: 32,
  xxl: 48,
};

export const radius = {
  sm: 10,
  md: 14,
  lg: 20,
  xl: 28,
  full: 999,
};

export const shadow = {
  sm: {
    shadowColor: colors.shadow,
    shadowOffset: { width: 0, height: 2 },
    shadowOpacity: 0.06,
    shadowRadius: 6,
    elevation: 2,
  },
  md: {
    shadowColor: colors.shadow,
    shadowOffset: { width: 0, height: 4 },
    shadowOpacity: 0.08,
    shadowRadius: 12,
    elevation: 4,
  },
  lg: {
    shadowColor: colors.shadow,
    shadowOffset: { width: 0, height: 8 },
    shadowOpacity: 0.12,
    shadowRadius: 24,
    elevation: 8,
  },
};

export const typography = {
  display: { fontSize: 32, fontWeight: "800" as const, color: colors.text },
  title: { fontSize: 28, fontWeight: "800" as const, color: colors.text },
  heading: { fontSize: 20, fontWeight: "700" as const, color: colors.text },
  subheading: { fontSize: 17, fontWeight: "600" as const, color: colors.text },
  body: { fontSize: 16, fontWeight: "400" as const, color: colors.text },
  bodyBold: { fontSize: 16, fontWeight: "600" as const, color: colors.text },
  caption: { fontSize: 13, fontWeight: "400" as const, color: colors.textSecondary },
  captionBold: { fontSize: 13, fontWeight: "600" as const, color: colors.textSecondary },
  micro: { fontSize: 11, fontWeight: "400" as const, color: colors.textTertiary },
  label: {
    fontSize: 11,
    fontWeight: "600" as const,
    color: colors.textTertiary,
    letterSpacing: 0.8,
    textTransform: "uppercase" as const,
  },
};
