import type { DomainDocument } from "../app/ports";
import type { PortSpec } from "../model/plant";

export function domainDocument(value: object): DomainDocument {
  return value as DomainDocument;
}

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

export const chillerPorts: { readonly ports: readonly PortSpec[] } = {
  ports: [
    { name: "cond_in", type: "cooling_water_temp", direction: "in" },
    { name: "spare_in", type: "chilled_water", direction: "in" },
    { name: "heat_out", type: "heat_rejection", direction: "out" },
    { name: "power_out", type: "electrical", direction: "out" },
  ],
};

export const towerPorts: { readonly ports: readonly PortSpec[] } = {
  ports: [
    { name: "heat_in", type: "heat_rejection", direction: "in" },
    { name: "power_in", type: "electrical", direction: "in" },
    { name: "basin_out", type: "cooling_water_temp", direction: "out" },
  ],
};

const beltLinks = [
  ["F-1.feed_out", "B-1.feed_in"],
  ["B-1.moving_out", "F-1.moving_in"],
  ["B-1.share_out", "D-1.share_in"],
  ["B-1.stop_ok_out", "D-1.stop_ok_in"],
  ["D-1.drag_out", "B-1.drag_in"],
  ["D-1.run_cap_out", "B-1.run_cap_in"],
  ["D-1.enabled_out", "B-1.drives_in"],
] as const;

export const beltDomain = {
  id: "belt",
  title: "Belt",
  site: { name: "Belt", areas: [{ name: "Belt" }, { name: "Transfer", parent: "Belt" }] },
  kinds: [
    { name: "feeder", channels: [{ name: "feeder.feed", unit: "kg/s" }] },
    { name: "belt", channels: [{ name: "belt.tail_speed", unit: "rpm" }] },
    {
      name: "drive",
      channels: ["current", "winding", "vibration", "run_status", "run_command", "torque", "slip"].map((name) => ({
        name: `drive.${name}`,
        unit: name === "winding" ? "degC" : "",
      })),
    },
  ],
  instances: [
    { kind: "belt", names: ["B-1"], area: "Transfer", provenance: { component: "lib/belt" } },
    { kind: "drive", names: ["D-1"], area: "Transfer", provenance: { component: "lib/drive" } },
    { kind: "feeder", names: ["F-1"], area: "Transfer", provenance: { component: "lib/feeder" } },
  ],
  links: beltLinks.map(([from, to]) => ({ from: from.split(".")[0], to: to.split(".")[0], type: "link" })),
};

export const beltComposition = { connections: beltLinks.map(([from, to]) => ({ from, to })) };

const beltPort = (name: string, type: string): PortSpec => ({
  name,
  type,
  direction: name.endsWith("_in") ? "in" : "out",
});

export const beltSpecs = new Map([
  [
    "lib/feeder",
    { ports: [beltPort("moving_in", "run_state"), beltPort("feed_out", "bulk_flow")], category: "prod" as const },
  ],
  [
    "lib/belt",
    {
      ports: [
        beltPort("feed_in", "bulk_flow"),
        beltPort("drag_in", "drag_force"),
        beltPort("run_cap_in", "force_capacity"),
        beltPort("drives_in", "run_state"),
        beltPort("moving_out", "run_state"),
        beltPort("share_out", "load_torque"),
        beltPort("stop_ok_out", "run_state"),
      ],
      category: "prod" as const,
    },
  ],
  [
    "lib/drive",
    {
      ports: [
        beltPort("share_in", "load_torque"),
        beltPort("stop_ok_in", "run_state"),
        beltPort("drag_out", "drag_force"),
        beltPort("run_cap_out", "force_capacity"),
        beltPort("enabled_out", "run_state"),
      ],
      category: "prod" as const,
    },
  ],
]);

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

export function componentShown(
  spec: { readonly ports: readonly PortSpec[] },
  document: object = {},
  entry: object = {},
) {
  const ports = spec.ports.map(({ name, type, direction, channel }) => ({
    name,
    type,
    direction: direction === "out" ? "provides" : "requires",
    ...(channel === undefined ? {} : { state: channel }),
  }));
  return { entry, document: { ...document, ports } };
}
