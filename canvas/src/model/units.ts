const units: Readonly<Record<string, string>> = {
  degC: "°C",
  degF: "°F",
  m3h: "m³/h",
  "m3/h": "m³/h",
  m3: "m³",
  "kWh/m3": "kWh/m³",
  pct: "%",
  percent: "%",
  ratio: "",
};

export function displayUnit(unit: string): string {
  return units[unit] ?? unit;
}

export function grouped(count: number): string {
  return count.toLocaleString("en-US").replaceAll(",", "\u202f");
}
