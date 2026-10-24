const units: Readonly<Record<string, string>> = {
  degC: "°C",
  Cel: "°C",
  degF: "°F",
  K: "K",
  "%": "%",
  pct: "%",
  percent: "%",
  "%RH": "% RH",
  "%/kPa": "%/kPa",
  "kW/%": "kW/%",
  permille: "‰",
  ppm: "ppm",
  pH: "pH",
  "1": "",
  rel: "",
  frac: "",
  ratio: "",
  bool: "",
  code: "",
  on: "",
  count: "count",
  units: "units",
  "units/h": "units/h",
  "packs/h": "packs/h",
  "1/h": "1/h",
  s: "s",
  ms: "ms",
  min: "min",
  h: "h",
  Hz: "Hz",
  rpm: "rpm",
  A: "A",
  V: "V",
  W: "W",
  kW: "kW",
  kWh: "kWh",
  "kWh/m3": "kWh/m³",
  kPa: "kPa",
  bar: "bar",
  Pa: "Pa",
  m: "m",
  "m/h": "m/h",
  "mm/s": "mm/s",
  m3: "m³",
  m3h: "m³/h",
  "m3/h": "m³/h",
  "L/h": "L/h",
  "L/min": "L/min",
  "kg/s": "kg/s",
  "t/h": "t/h",
  "mg/L": "mg/L",
  "N.m": "N·m",
  "currency/h": "currency/h",
  "currency/kWh": "currency/kWh",
};

export function displayUnit(unit: string): string {
  return units[unit] ?? unit;
}

export function presentedUnits(): readonly string[] {
  return Object.keys(units);
}

export function codeTitle(unit: string): string {
  return unit === "" || displayUnit(unit) === unit ? "" : `unit code: ${unit}`;
}

export function withUnit(value: string, unit: string): string {
  const symbol = displayUnit(unit);
  return symbol === "" ? value : `${value} ${symbol}`;
}

export function grouped(count: number): string {
  return count.toLocaleString("en-US").replaceAll(",", " ");
}

export function groupedFigure(figure: string): string {
  const [whole = "", ...rest] = figure.split(".");
  const spaced = whole.replace(/\B(?=(\d{3})+(?!\d))/g, " ");
  return [spaced, ...rest].join(".");
}
