const hash = (digit: string) => `sha256:${digit.repeat(64)}`;

function plant(id: string, version: string, extra: Record<string, unknown>) {
  return {
    domain: { id, version, title: id, digest: hash("a"), file: `domains/${id}.domain.json` },
    device_kinds: [],
    runs: [],
    ...extra,
  };
}

const edit = {
  verdict: "editable",
  reason: "composition",
  route: { code: "edit_composition", sentence: "Edit the composition, then compose it again." },
};

const source = (file: string, id: string, status: string) => ({ file, id, version: "0.1.0", status });

const factoryPlant = plant("factory-utilities", "0.1.0", {
  composition: source("compositions/factory-utilities.composition.json", "factory-utilities", "in_sync"),
  runs: [
    {
      id: "w-1",
      mode: "evaluate",
      recorded: true,
      domain_version: "0.1.0",
      domain_digest: hash("a"),
      same_as_current: true,
    },
    {
      id: "w-2",
      mode: "explore",
      recorded: true,
      domain_version: "0.1.0",
      domain_digest: hash("b"),
      same_as_current: false,
    },
  ],
  editable: edit,
});

const twinPlant = plant("arduino-thermal-twin", "0.1.0", {
  twin: source("twins/fan.twin.json", "arduino-thermal-twin", "in_sync"),
  device_kinds: [
    { kind: "fan", status: "found", file: "kinds/fan.kind.json", digest: hash("c") },
    { kind: "thermal_sensor", status: "found", file: "kinds/thermal_sensor.kind.json", digest: hash("d") },
  ],
  editable: {
    verdict: "read_only",
    reason: "twin",
    route: {
      code: "open_device_behavior",
      sentence: "Generated from the twin choices twins/fan.twin.json: change it in the device editor.",
    },
  },
});

const examplePlant = plant("chiller-pilot", "0.1.0", {
  editable: {
    verdict: "read_only",
    reason: "example",
    route: {
      code: "create_from_example",
      sentence: "A shipped example: create a plant from it under a new id, then edit that composition.",
    },
  },
});

const handPlant = plant("alarm-playback", "0.1.0", {
  editable: {
    verdict: "read_only",
    reason: "hand_authored",
    route: {
      code: "copy_domain_file",
      sentence: "Authored as a domain file: copy the file as a new domain id and edit the copy.",
    },
  },
});

const outOfSyncPlant = plant("campus-3", "0.1.0", {
  composition: {
    ...source("compositions/campus-3.composition.json", "campus-3", "out_of_sync"),
    reason: "version differs",
  },
  editable: {
    verdict: "editable",
    reason: "composition_out_of_sync",
    route: { code: "recompose", sentence: "The composition no longer produces this domain file: compose it again." },
  },
});

function device(kind: string, extra: Record<string, unknown>) {
  return {
    kind,
    file: `kinds/${kind}.kind.json`,
    digest: hash("c"),
    twins_including: [],
    has_own_twin: false,
    implemented_by: [],
    ...extra,
  };
}

const inFanTwin = [{ twin: "arduino-thermal-twin", file: "twins/fan.twin.json", kinds: 2 }];

export const lineageWire = {
  schema_version: "0.1",
  plants: [factoryPlant, twinPlant, examplePlant, handPlant, outOfSyncPlant],
  device_kinds: [
    device("fan", { twins_including: inFanTwin }),
    device("thermal_sensor", { twins_including: inFanTwin }),
    device("heater", {
      twins_including: [{ twin: "heater-twin", file: "twins/heater.twin.json", kinds: 1 }],
      has_own_twin: true,
    }),
    device("lonely", {}),
  ],
  unreadable: [{ file: "kinds/broken.kind.json", code: "kind_file_unreadable", message: "not JSON" }],
  future_member: { ignored: true },
};

const reading = {
  channel: "temperature",
  unit: "Cel",
  value_type: "number",
  range: { min: 0, max: 50 },
  resolution: 1,
  noise: { model: "gaussian", sigma: 0.1 },
  cadence: { mode: "periodic", period_s: 10 },
  absence: "signal",
  basis: "library_instrument",
  instrument: { id: "streamsim/thermistor", version: 1, ref: "streamsim/thermistor@1" },
  recorded_in: "twin_choices",
};

export const assumptionsWire = {
  schema_version: "0.1",
  domain: { id: "arduino-thermal-twin", version: "0.1.0", digest: hash("a") },
  origin: { kind: "twin", twin: { id: "arduino-thermal-twin", version: "0.1.0", status: "in_sync" } },
  kinds: [
    {
      kind: "thermal_sensor",
      implements: "thermal_sensor",
      instances: [{ names: ["zone-01"], area: "thermal-bench" }],
      readings: [
        { ...reading, equals_binary_suggestion: false },
        {
          ...reading,
          channel: "humidity",
          unit: "%",
          basis: "plain_channel",
          instrument: undefined,
          equals_binary_suggestion: true,
        },
      ],
      dynamics: [
        {
          target: "temperature_value",
          tier: "F1",
          form: "first_order_lag",
          gain: 1,
          time_constant_s: 180,
          clamp: { min: 0, max: 50 },
          inputs: [],
        },
      ],
      commands: [
        {
          name: "set_led",
          risk_class: "low",
          safe_stop: false,
          arguments: [],
          response: "first_order",
          recorded_in: "twin_choices",
          sets: [],
          drives: [
            { from: "fan-01", kind: "thermal_sensor", metric: "temperature", gain: -0.02, time_constant_s: 180 },
          ],
          effect: { dead_time_s: 0, time_constant_s: 180, state_changes: 1 },
          acknowledgement: { latency_ms_mean: 40, latency_ms_sigma: 10, failure_modes: [] },
          equals_binary_suggestion: true,
          statement: "Moves temperature toward the commanded value.",
        },
      ],
      faults: [
        {
          id: "thermal_sensor__temperature__offset",
          negative_class: false,
          onset_shape: "step",
          affects: ["temperature__bias"],
        },
      ],
    },
  ],
  not_modeled: [
    {
      code: "no_calibration_evidence",
      scope: "domain",
      sentence: "No calibration or hardware validation is attached.",
    },
    {
      code: "first_order_response_only",
      scope: "domain",
      sentence: "A command moves a reading directly or through a first-order lag.",
    },
  ],
};

const set = (names: string[]) => ({ count: names.length, names, id_templates: [] });

export const diffWire = {
  schema_version: "0.1",
  a: {
    id: "factory-utilities",
    version: "0.1.0",
    title: "Factory",
    digest: hash("a"),
    source: { kind: "run", label: "w-1" },
  },
  b: {
    id: "factory-utilities",
    version: "0.1.0",
    title: "Factory",
    digest: hash("b"),
    source: { kind: "file", label: "domains/factory-utilities.domain.json" },
  },
  identity: { same_id: true, same_version: true, same_digest: false, same_id_and_version_different_digest: true },
  kinds: [
    {
      kind: "thermal_sensor",
      status: "changed",
      instances: { before: set(["zone-01"]), after: set(["zone-01"]) },
      relations: [
        { type: "fan_drives_thermal_sensor", from_kind: "fan", to_kind: "thermal_sensor", links: ["fan-01>zone-01"] },
      ],
      changes: [
        {
          code: "gain_changed",
          subject: "thermal_sensor/dynamics/temperature_value/inputs/duty@fan_drives_thermal_sensor",
          section: "dynamics",
          field: "coef",
          before: -0.02,
          after: -0.05,
          sentence: "thermal_sensor/dynamics/temperature_value: coef changed from -0.02 to -0.05.",
        },
      ],
    },
  ],
  domain_changes: [],
  summary: { changes: 1, kinds_changed: 1, kinds_added: 0, kinds_removed: 0 },
  unchanged_summary: { kinds: 2, readings: 5, dynamics: 0, commands: 4, faults: 14 },
};

export function lineageOf(...plants: readonly object[]) {
  return { ...lineageWire, plants };
}

export function miniLineagePlant(extra: Record<string, unknown>) {
  return plant("mini", "0.1.0", extra);
}

export const miniEditable = edit;
