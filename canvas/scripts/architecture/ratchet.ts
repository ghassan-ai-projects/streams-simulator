import type { ModuleSpec, Register } from "./model.ts";
import { parseRegister } from "./register.ts";
import { introducingCommit, mergeBase, show } from "./git.ts";

export interface RatchetBase {
  readonly commit: string;
  readonly label: string;
}

export function registerBase(path: string, reference = "main"): RatchetBase | null {
  const introduced = introducingCommit(path);
  if (introduced !== null && introduced !== "")
    return { commit: introduced, label: `register as first committed in ${introduced.slice(0, 7)}` };
  const base = mergeBase(reference);
  return base === null ? null : { commit: base, label: `register at the branch point ${base.slice(0, 7)}` };
}

export function registerAt(commit: string, path: string): Register | null {
  const text = show(commit, path);
  if (text === null) return null;
  return parseRegister(JSON.parse(text) as unknown, `${commit}:${path}`);
}

export function registerRegressions(previous: Register, current: Register): string[] {
  return [
    ...capRegressions(previous, current),
    ...allowListRegressions(previous, current),
    ...moduleRegressions(previous, current),
    ...featureRegressions(previous, current),
    ...emittedRegressions(previous, current),
    ...probeRegressions(previous, current),
  ];
}

function moduleRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  for (const [name, before] of Object.entries(previous.modules)) {
    const now = current.modules[name];
    if (now === undefined) out.push(`module ${name} is gone from the register`);
    else out.push(...moduleAllowance(name, before, now));
  }
  for (const [name, now] of Object.entries(current.modules)) {
    if (previous.modules[name] === undefined && now.mayImportAny === true) {
      out.push(`new module ${name} may import anything`);
    }
  }
  for (const support of current.testSupport) {
    if (!previous.testSupport.includes(support)) out.push(`test support ${support} was added`);
  }
  return out;
}

function moduleAllowance(name: string, before: ModuleSpec, now: ModuleSpec): string[] {
  const out: string[] = [];
  for (const imported of now.mayImport ?? []) {
    if (!(before.mayImport ?? []).includes(imported)) out.push(`module ${name} may now import ${imported}`);
  }
  if (now.mayImportAny === true && before.mayImportAny !== true) out.push(`module ${name} may now import anything`);
  if (now.entrypoint === null && before.entrypoint !== null)
    out.push(`module ${name} lost its entrypoint ${before.entrypoint}`);
  return out;
}

function featureRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  for (const [folder, owner] of Object.entries(previous.features)) {
    const now = current.features[folder];
    if (now === undefined) out.push(`feature folder ${folder} is gone`);
    else if (now !== owner) out.push(`feature folder ${folder} moved from ${owner} to ${now}`);
  }
  for (const folder of Object.keys(current.features)) {
    if (previous.features[folder] !== undefined) continue;
    if (!Object.keys(previous.features).some((known) => folder.startsWith(`${known}/`))) {
      out.push(`feature folder ${folder} is not narrower than a registered folder`);
    }
  }
  return out;
}

function emittedRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  for (const before of previous.emittedCss.order) {
    if (!current.emittedCss.order.some((entry) => entry.name === before.name)) {
      out.push(`emitted stylesheet sentinel ${before.name} is gone`);
    }
  }
  const counts: readonly (keyof Register["emittedCss"])[] = ["important", "zIndex", "positioned"];
  for (const key of counts) out.push(...emittedCount(key, previous, current));
  return out;
}

function emittedCount(key: keyof Register["emittedCss"], previous: Register, current: Register): string[] {
  const was = previous.emittedCss[key];
  const now = current.emittedCss[key];
  if (typeof was !== "number" || typeof now !== "number" || was === now) return [];
  if (key === "important" && now < was) return [];
  const noted = current.emittedCss.note !== "" && current.emittedCss.note !== previous.emittedCss.note;
  return noted
    ? []
    : [
        `the built stylesheet's ${key} count changed ${String(was)} -> ${String(now)} without a note naming the correction`,
      ];
}

function probeRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  if (previous.rules.protocolProductionOnly && !current.rules.protocolProductionOnly) {
    out.push("protocol scanning no longer covers tests");
  }
  if (previous.rules.protocolOwner !== current.rules.protocolOwner) out.push("the protocol owner module changed");
  for (const rule of previous.rules.depcruiseRules) {
    if (!current.rules.depcruiseRules.some((entry) => entry.name === rule.name))
      out.push(`dependency-cruiser rule ${rule.name} is no longer cross-checked`);
  }
  return out;
}

function capRegressions(previous: Register, current: Register): string[] {
  return [
    ...cssCapRegressions(previous, current),
    ...floorRegressions(previous, current),
    ...debtRegressions(previous, current),
  ];
}

function cssCapRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  for (const [path, entry] of Object.entries(current.css.caps)) {
    const before = previous.css.caps[path];
    if (before === undefined) {
      if (entry.cap > current.css.defaultCap)
        out.push(`new css cap ${path} is above the default ${String(current.css.defaultCap)}`);
    } else if (entry.cap > before.cap) {
      out.push(`css cap ${path} rose ${String(before.cap)} -> ${String(entry.cap)}`);
    }
  }
  for (const [path, entry] of Object.entries(current.css.globalCaps)) {
    out.push(...allowanceRegression(path, entry, previous.css.globalCaps[path]));
  }
  for (const [path, entry] of Object.entries(current.css.colorExceptions)) {
    const before = previous.css.colorExceptions[path];
    if (before === undefined) out.push(`new colour exception ${path}`);
    else if (entry.literals > before.literals) out.push(`colour exception ${path} rose`);
  }
  return out;
}

function allowanceRegression(
  path: string,
  entry: { readonly global: number; readonly important: number },
  before: { readonly global: number; readonly important: number } | undefined,
): string[] {
  if (before === undefined) return [`new :global/!important allowance ${path}`];
  if (entry.global > before.global || entry.important > before.important) return [`allowance ${path} rose`];
  return [];
}

function floorRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  for (const [glob, entry] of Object.entries(current.floors)) {
    const before = previous.floors[glob];
    if (before === undefined) continue;
    for (const [metric, value] of Object.entries(entry)) {
      const was = before[metric];
      if (was !== undefined && value < was)
        out.push(`coverage floor ${glob} ${metric} fell ${String(was)} -> ${String(value)}`);
    }
  }
  return out;
}

function debtRegressions(previous: Register, current: Register): string[] {
  const out: string[] = [];
  for (const entry of current.debt) {
    const before = previous.debt.find(
      (old) => old.rule === entry.rule && old.file === entry.file && old.target === entry.target,
    );
    if (before === undefined)
      out.push(`new debt ${entry.rule} ${entry.file}${entry.target === undefined ? "" : ` <- ${entry.target}`}`);
    else if (entry.count > before.count)
      out.push(`debt ${entry.rule} ${entry.file} rose ${String(before.count)} -> ${String(entry.count)}`);
  }
  return out;
}

function allowListRegressions(previous: Register, current: Register): string[] {
  return [
    ...grew("boolean fact allow-list", previous.rules.booleanFactNames, current.rules.booleanFactNames),
    ...grew("fake-timer test allow-list", previous.rules.fakeTimerTests, current.rules.fakeTimerTests),
    ...grew("global stylesheet allow-list", previous.css.globalOwners, current.css.globalOwners),
    ...grew("global entrypoint allow-list", previous.css.globalEntrypoints, current.css.globalEntrypoints),
    ...grew("vendor stylesheet allow-list", previous.css.vendorFiles, current.css.vendorFiles),
    ...shrank("simulator action detection list", previous.rules.actionNames, current.rules.actionNames),
    ...shrank("restricted app import list", previous.rules.restrictedAppImports, current.rules.restrictedAppImports),
    ...shrank("restricted app global list", previous.rules.restrictedAppGlobals, current.rules.restrictedAppGlobals),
    ...shrank("clock global list", previous.rules.clockGlobals, current.rules.clockGlobals),
    ...shrank("restricted ui import list", previous.rules.restrictedUiImports, current.rules.restrictedUiImports),
  ];
}

function grew(what: string, before: readonly string[], after: readonly string[]): string[] {
  return after.filter((entry) => !before.includes(entry)).map((entry) => `${what} gained ${entry}`);
}

function shrank(what: string, before: readonly string[], after: readonly string[]): string[] {
  return before.filter((entry) => !after.includes(entry)).map((entry) => `${what} lost ${entry}`);
}
