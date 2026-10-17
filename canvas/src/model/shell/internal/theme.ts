export type ThemePreference = "system" | "light" | "dark";

export type Theme = "light" | "dark";

const order: readonly ThemePreference[] = ["system", "light", "dark"];

export function resolveTheme(preference: ThemePreference, systemPrefersDark: boolean): Theme {
  if (preference === "system") {
    return systemPrefersDark ? "dark" : "light";
  }
  return preference;
}

export function nextPreference(preference: ThemePreference): ThemePreference {
  const index = order.indexOf(preference);
  return order[(index + 1) % order.length] ?? "system";
}

export function parsePreference(value: string | null): ThemePreference {
  return value === "light" || value === "dark" ? value : "system";
}

export function preferenceLabel(preference: ThemePreference): string {
  switch (preference) {
    case "system":
      return "Theme: system";
    case "light":
      return "Theme: light";
    case "dark":
      return "Theme: dark";
  }
}
