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
export interface Shape {
  readonly extracted?: boolean;
  readonly graphLayer?: string;
  readonly caps?: Readonly<Record<string, Cap>>;
  readonly globalCaps?: Readonly<Record<string, Allowance>>;
  readonly vendorFiles?: readonly string[];
  readonly debt?: readonly Debt[];
}

export function register(shape: Shape): Register {
  const graph = shape.extracted === true;
  return parseRegister(
    {
      version: 1,
      source: "src",
      modules: fixtureModules(graph, shape.graphLayer),
      css: {
        globalEntrypoints: [],
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

export function compare(name: string, before: Register, after: Register, expected: readonly string[]): string[] {
  const actual = registerRegressions(before, after).sort();
  if (JSON.stringify(actual) === JSON.stringify([...expected].sort())) return [];
  return [`ratchet ${name}: expected [${expected.join("; ")}], found [${actual.join("; ")}]`];
}
