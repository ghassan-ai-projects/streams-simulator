export const domain = {
  id: "mini",
  title: "Mini plant",
  site: {
    name: "Mini",
    areas: [{ name: "Mini" }, { name: "Cooling", parent: "Mini" }, { name: "Power", parent: "Mini" }],
  },
  kinds: [
    {
      name: "chiller",
      state: [{ name: "cop" }],
      channels: [{ name: "chiller.power", unit: "kW" }],
      faults: [{ id: "chiller__fouling" }],
      effectors: [{ name: "start" }],
    },
    { name: "tower", channels: [{ name: "tower.basin_temp", unit: "degC" }] },
    { name: "meter", channels: [{ name: "meter.power", unit: "kW" }] },
  ],
  instances: [
    { kind: "chiller", names: ["CH-1"], area: "Cooling", provenance: { component: "lib/chiller" } },
    { kind: "tower", names: ["CT-1"], area: "Cooling", provenance: { component: "lib/tower" } },
    { kind: "meter", names: ["MB-1"], area: "Power" },
  ],
  links: [
    { from: "CH-1", to: "CT-1", type: "tower__heat_in__chiller" },
    { from: "CT-1", to: "CH-1", type: "chiller__cond_in__tower" },
    { from: "CH-1", to: "MB-1", type: "meter__loads_in__chiller" },
  ],
};

export const composition = {
  connections: [
    { from: "CH-1.heat_out", to: "CT-1.heat_in" },
    { from: "CT-1.basin_out", to: "CH-1.cond_in" },
    { from: "CH-1.power_out", to: "MB-1.loads_in" },
  ],
};

export const chillerPorts = {
  ports: [
    { name: "cond_in", type: "cooling_water_temp", direction: "consumes" },
    { name: "spare_in", type: "chilled_water", direction: "consumes" },
    { name: "heat_out", type: "heat_rejection", direction: "provides" },
    { name: "power_out", type: "electrical", direction: "provides" },
  ],
};

export const towerPorts = {
  ports: [
    { name: "heat_in", type: "heat_rejection", direction: "consumes" },
    { name: "power_in", type: "electrical", direction: "consumes" },
    { name: "basin_out", type: "cooling_water_temp", direction: "provides" },
  ],
};

export function campus(halls: number) {
  const names = (base: string) => Array.from({ length: halls }, (_, i) => `h${String(i + 1).padStart(2, "0")}/${base}`);
  return {
    id: "campus",
    title: "Campus",
    site: { name: "Campus", areas: [{ name: "Campus" }] },
    kinds: [
      { name: "line", channels: [{ name: "line.power", unit: "kW" }] },
      { name: "calendar", channels: [{ name: "calendar.shift", unit: "" }] },
    ],
    instances: [
      { kind: "calendar", names: ["CAL-1"], area: "Campus" },
      { kind: "line", names: [...names("M-1"), ...names("M-2")], area: "Campus" },
    ],
    links: names("M-1").flatMap((id) => [
      { from: "CAL-1", to: id, type: "line__shift_in__calendar" },
      { from: "CAL-1", to: id.replace("M-1", "M-2"), type: "line__shift_in__calendar" },
      { from: id, to: id.replace("M-1", "M-2"), type: "line__feed__line" },
    ]),
  };
}
