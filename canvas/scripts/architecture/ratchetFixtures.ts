import { parseRegister } from "./register.ts";
import { registerRegressions } from "./ratchet.ts";
import type { Register } from "./model.ts";

interface Cap {
  readonly cap: number;
  readonly phase: string;
}
interface Allowance {
  readonly global: number;
  readonly important: number;
  readonly phase: string;
}
interface Debt {
  readonly rule: string;
  readonly file: string;
  readonly target?: string;
  readonly count: number;
  readonly phase: string;
}
interface Shape {
  readonly extracted?: boolean;
  readonly graphLayer?: string;
  readonly caps?: Readonly<Record<string, Cap>>;
  readonly globalCaps?: Readonly<Record<string, Allowance>>;
  readonly vendorFiles?: readonly string[];
  readonly debt?: readonly Debt[];
}

const phase = "P1";
const oldNodes = "src/ui/plant/nodes.module.css";
const newNodes = "src/ui/graph/internal/nodes.module.css";
const oldControls = "src/ui/plant/GraphControls.module.css";
const vendor = "src/ui/graph/internal/vendor.module.css";

export function runRatchetFixtures(): string[] {
  return [
    ...validMoveFixture(),
    ...unrelatedMoveFixture(),
    ...vendorCapFixture(),
    ...vendorSourceFixture(),
    ...crossLayerFixture(),
  ];
}

function validMoveFixture(): string[] {
  const base = register({
    caps: { [oldNodes]: { cap: 695, phase } },
    globalCaps: {
      [oldNodes]: { global: 70, important: 30, phase },
      [oldControls]: { global: 0, important: 2, phase },
    },
    debt: [
      { rule: "boolean-flag-param", file: "src/ui/plant/flow.ts", count: 1, phase },
      {
        rule: "css-feature-ownership",
        file: "src/ui/shared.module.css",
        target: "src/ui/plant/Outline.tsx",
        count: 1,
        phase,
      },
    ],
  });
  const moved = register({
    extracted: true,
    caps: { [newNodes]: { cap: 695, phase } },
    globalCaps: {
      [newNodes]: { global: 0, important: 30, phase },
      [vendor]: { global: 6, important: 2, phase },
    },
    vendorFiles: [vendor],
    debt: [
      { rule: "boolean-flag-param", file: "src/ui/graph/internal/flow.ts", count: 1, phase },
      {
        rule: "css-feature-ownership",
        file: "src/ui/shared.module.css",
        target: "src/ui/graph/internal/Outline.tsx",
        count: 1,
        phase,
      },
    ],
  });
  return compare("same-basename extraction and approved vendor move", base, moved, []);
}

function unrelatedMoveFixture(): string[] {
  const base = register({
    caps: { [oldNodes]: { cap: 695, phase } },
    globalCaps: { [oldNodes]: { global: 70, important: 30, phase } },
    debt: [{ rule: "boolean-flag-param", file: "src/ui/plant/flow.ts", count: 1, phase }],
  });
  const unrelated = register({
    extracted: true,
    caps: { "src/ui/graph/internal/renamed.module.css": { cap: 400, phase } },
    globalCaps: { "src/ui/graph/internal/other.module.css": { global: 0, important: 2, phase } },
    debt: [{ rule: "boolean-flag-param", file: "src/ui/graph/internal/other.ts", count: 1, phase }],
  });
  return compare("unrelated paths do not inherit moved allowances", base, unrelated, [
    "new :global/!important allowance src/ui/graph/internal/other.module.css",
    "new css cap src/ui/graph/internal/renamed.module.css is above the default 300",
    "new debt boolean-flag-param src/ui/graph/internal/other.ts",
  ]);
}

function vendorCapFixture(): string[] {
  const base = register({
    globalCaps: { [oldControls]: { global: 0, important: 2, phase } },
  });
  const badVendor = register({
    extracted: true,
    globalCaps: { [vendor]: { global: 7, important: 3, phase } },
    vendorFiles: [vendor],
  });
  return compare("vendor registration stays fixed to its approved allowance", base, badVendor, [
    "new :global/!important allowance src/ui/graph/internal/vendor.module.css",
    "vendor stylesheet allow-list gained src/ui/graph/internal/vendor.module.css",
  ]);
}

function vendorSourceFixture(): string[] {
  const noSourceVendor = register({
    extracted: true,
    globalCaps: {
      [newNodes]: { global: 0, important: 30, phase },
      [vendor]: { global: 6, important: 2, phase },
    },
    vendorFiles: [vendor],
  });
  return compare(
    "vendor registration requires the moved source allowance",
    register({
      caps: { [oldNodes]: { cap: 695, phase } },
      globalCaps: { [oldNodes]: { global: 70, important: 30, phase } },
    }),
    noSourceVendor,
    [
      "new :global/!important allowance src/ui/graph/internal/vendor.module.css",
      "vendor stylesheet allow-list gained src/ui/graph/internal/vendor.module.css",
    ],
  );
}

function crossLayerFixture(): string[] {
  const base = register({
    caps: { [oldNodes]: { cap: 695, phase } },
    globalCaps: { [oldNodes]: { global: 0, important: 30, phase } },
  });
  const crossLayer = register({
    extracted: true,
    graphLayer: "model",
    caps: { [newNodes]: { cap: 695, phase } },
    globalCaps: { [newNodes]: { global: 0, important: 30, phase } },
  });
  return compare("cross-layer paths do not inherit moved allowances", base, crossLayer, [
    "new css cap src/ui/graph/internal/nodes.module.css is above the default 300",
    "new :global/!important allowance src/ui/graph/internal/nodes.module.css",
  ]);
}

function register(shape: Shape): Register {
  const graph = shape.extracted === true;
  return parseRegister(
    {
      version: 1,
      source: "src",
      modules: fixtureModules(graph, shape.graphLayer),
      css: {
        globalEntrypoints: [],
        globalOwners: [],
        vendorFiles: shape.vendorFiles ?? [],
        defaultCap: 300,
        caps: shape.caps ?? {},
        globalCaps: shape.globalCaps ?? {},
        colorExceptions: {},
      },
      rules: emptyRules(),
      emittedCss: { order: [], important: 0, zIndex: 0, positioned: 0, note: "" },
      debt: shape.debt ?? [],
    },
    "ratchet fixture",
  );
}

function fixtureModules(graph: boolean, graphLayer = "ui"): Register["modules"] {
  const modules: Record<string, unknown> = {
    ui: { kind: "ui", layer: "ui", lifetime: "components", responsibility: "ui", folders: ["src/ui"] },
  };
  if (graph)
    modules["ui/graph"] = {
      kind: "ui",
      layer: graphLayer,
      lifetime: "components",
      responsibility: "graph",
      folders: ["src/ui/graph"],
    };
  return modules as Register["modules"];
}

function emptyRules(): Register["rules"] {
  return {
    actionNames: [],
    protocolOwner: "",
    protocolProductionOnly: false,
    restrictedAppImports: [],
    restrictedAppGlobals: [],
    restrictedUiImports: [],
    clockGlobals: [],
    booleanFactNames: [],
    fakeTimerTests: [],
    depcruiseRules: [],
  };
}

function compare(name: string, before: Register, after: Register, expected: readonly string[]): string[] {
  const actual = registerRegressions(before, after).sort();
  if (JSON.stringify(actual) === JSON.stringify([...expected].sort())) return [];
  return [`ratchet ${name}: expected [${expected.join("; ")}], found [${actual.join("; ")}]`];
}
