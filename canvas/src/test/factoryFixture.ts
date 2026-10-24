import { success } from "../app/ports";
import { FakeBridge } from "./fakeBridge";
import { observation } from "./runFixture";
import { sampleWorkspace } from "./workspaceFixture";

interface FactoryGroup {
  readonly kind: string;
  readonly names: readonly string[];
  readonly channels: readonly (readonly [name: string, unit: string])[];
}

const groups: readonly FactoryGroup[] = [
  {
    kind: "weather",
    names: ["WX-1"],
    channels: [
      ["weather.dry_bulb_temp", "degC"],
      ["weather.wet_bulb_temp", "degC"],
      ["weather.humidity_rh", "%RH"],
    ],
  },
  { kind: "site_calendar", names: ["CAL-1"], channels: [["site_calendar.site_load", "kW"]] },
  {
    kind: "molder",
    names: ["M-1", "M-2"],
    channels: [
      ["molder.power", "kW"],
      ["molder.vibration", "mm/s"],
      ["molder.air_flow", "m3/h"],
      ["molder.units_made", "count"],
    ],
  },
  {
    kind: "cnc",
    names: ["M-3"],
    channels: [
      ["cnc.power", "kW"],
      ["cnc.vibration", "mm/s"],
      ["cnc.air_flow", "m3/h"],
      ["cnc.units_made", "count"],
    ],
  },
  {
    kind: "packaging_line",
    names: ["M-4"],
    channels: [
      ["packaging_line.power", "kW"],
      ["packaging_line.vibration", "mm/s"],
      ["packaging_line.air_flow", "m3/h"],
      ["packaging_line.units_made", "count"],
    ],
  },
  {
    kind: "ahu",
    names: ["AHU-1"],
    channels: [
      ["ahu.supply_air_temp", "degC"],
      ["ahu.cooling_load", "kW"],
    ],
  },
  {
    kind: "compressor_lead",
    names: ["AC-1"],
    channels: [
      ["compressor_lead.air_flow", "m3/h"],
      ["compressor_lead.discharge_temp", "degC"],
      ["compressor_lead.power", "kW"],
      ["compressor_lead.vibration", "mm/s"],
      ["compressor_lead.run_status", ""],
      ["compressor_lead.load_status", ""],
      ["compressor_lead.alarm_state", "on"],
    ],
  },
  {
    kind: "compressor_lag",
    names: ["AC-2"],
    channels: [
      ["compressor_lag.air_flow", "m3/h"],
      ["compressor_lag.discharge_temp", "degC"],
      ["compressor_lag.power", "kW"],
      ["compressor_lag.vibration", "mm/s"],
      ["compressor_lag.run_status", ""],
      ["compressor_lag.load_status", ""],
      ["compressor_lag.alarm_state", "on"],
    ],
  },
  {
    kind: "air_receiver",
    names: ["R-1"],
    channels: [
      ["air_receiver.header_pressure", "kPa"],
      ["air_receiver.outlet_flow", "m3/h"],
      ["air_receiver.condensate_level", "m"],
      ["air_receiver.drain_status", ""],
    ],
  },
  {
    kind: "air_dryer",
    names: ["D-1"],
    channels: [
      ["air_dryer.dew_point", "degC"],
      ["air_dryer.pressure_drop", "kPa"],
      ["air_dryer.power", "kW"],
      ["air_dryer.run_status", ""],
    ],
  },
  {
    kind: "chiller",
    names: ["CH-1"],
    channels: [
      ["chiller.chw_supply_temp", "degC"],
      ["chiller.chw_return_temp", "degC"],
      ["chiller.cond_temp", "degC"],
      ["chiller.power", "kW"],
    ],
  },
  {
    kind: "cooling_tower",
    names: ["CT-1"],
    channels: [
      ["cooling_tower.basin_temp", "degC"],
      ["cooling_tower.fan_power", "kW"],
    ],
  },
  {
    kind: "chw_pump",
    names: ["P-1", "P-2", "P-3"],
    channels: [
      ["chw_pump.flow", "m3/h"],
      ["chw_pump.power", "kW"],
    ],
  },
  {
    kind: "cw_pump",
    names: ["P-4"],
    channels: [
      ["cw_pump.flow", "m3/h"],
      ["cw_pump.power", "kW"],
    ],
  },
  {
    kind: "air_meter",
    names: ["MB-AIR"],
    channels: [
      ["air_meter.power", "kW"],
      ["air_meter.demand", "kW"],
      ["air_meter.energy", "kWh"],
      ["air_meter.current", "A"],
    ],
  },
  {
    kind: "cooling_meter",
    names: ["MB-COOL"],
    channels: [
      ["cooling_meter.power", "kW"],
      ["cooling_meter.demand", "kW"],
      ["cooling_meter.energy", "kWh"],
      ["cooling_meter.current", "A"],
    ],
  },
  {
    kind: "process_meter",
    names: ["MB-PROC"],
    channels: [
      ["process_meter.power", "kW"],
      ["process_meter.demand", "kW"],
      ["process_meter.energy", "kWh"],
      ["process_meter.current", "A"],
    ],
  },
  {
    kind: "main_meter",
    names: ["MM-1"],
    channels: [
      ["main_meter.power", "kW"],
      ["main_meter.demand", "kW"],
      ["main_meter.energy", "kWh"],
      ["main_meter.current", "A"],
    ],
  },
  {
    kind: "transformer",
    names: ["TX-1"],
    channels: [
      ["transformer.load", "kW"],
      ["transformer.winding_temp", "degC"],
      ["transformer.oil_temp", "degC"],
      ["transformer.ambient_temp", "degC"],
    ],
  },
  {
    kind: "generator",
    names: ["G-1"],
    channels: [
      ["generator.output", "kW"],
      ["generator.running", ""],
      ["generator.fuel_level", "m"],
      ["generator.engine_temp", "degC"],
    ],
  },
];

const factoryDomain = {
  id: "factory-utilities",
  title: "Factory utilities",
  site: { name: "Factory", areas: [{ name: "Factory" }] },
  kinds: groups.map((group) => ({
    name: group.kind,
    channels: group.channels.map(([name, unit]) => ({ name, unit })),
  })),
  instances: groups.map((group) => ({ kind: group.kind, names: group.names, area: "Factory" })),
  links: [],
};

export const factoryPath = "domains/factory-utilities.domain.json";
const startNs = 1_767_225_600e9;

interface Series {
  readonly entity: string;
  readonly channel: string;
  readonly unit: string;
}

function factorySeries(): readonly Series[] {
  return groups.flatMap((group) =>
    group.names.flatMap((entity) => group.channels.map(([channel, unit]) => ({ entity, channel, unit }))),
  );
}

function valueOf(index: number, second: number): number {
  return Math.round((20 + 5 * Math.sin(second / 300 + index)) * 100) / 100;
}

export function factoryJournal(fromS: number, toS: number, stepS = 10): object[] {
  const series = factorySeries();
  const entries: object[] = [];
  for (let second = fromS; second < toS; second += stepS) {
    series.forEach((one, index) => {
      const seq = entries.length + 1;
      entries.push(observation(seq, startNs + second * 1e9, one.entity, one.channel, valueOf(index, second), one.unit));
    });
    if (second % 60 === 0) entries.push(clockAt(entries.length + 1, second));
  }
  return entries;
}

function clockAt(seq: number, second: number): object {
  const atNs = startNs + second * 1e9;
  return { stream: "observer", type: "clock.advanced", seq, at_ns: atNs, to_ns: atNs };
}

export function factoryBridge(): FakeBridge {
  const fake = new FakeBridge();
  fake.sessionOutcome = success(true);
  fake.workspaceOutcome = success({
    ...sampleWorkspace,
    domains: [{ path: factoryPath, id: factoryDomain.id, title: "Factory utilities", instances: 23 }],
    runs: [],
  });
  fake.addFile(factoryPath, factoryDomain);
  return fake;
}

export function singleReadings(entity: string, fromS: number, toS: number, stepS = 10): object[] {
  const series = factorySeries();
  const entries: object[] = [];
  for (let second = fromS; second < toS; second += stepS) {
    series.forEach((one, index) => {
      if (one.entity !== entity) return;
      const atNs = startNs + second * 1e9;
      entries.push(observation(entries.length + 1, atNs, one.entity, one.channel, valueOf(index, second), one.unit));
    });
  }
  return entries;
}

export function clockOnly(second: number): object {
  return clockAt(1, second);
}
