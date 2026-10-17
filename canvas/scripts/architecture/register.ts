import { readFileSync } from "node:fs";
import type { CssCap, CssSpec, DebtEntry, EmittedSpec, ModuleSpec, Register, RuleSpec } from "./model.ts";

type Json = Record<string, unknown>;

class RegisterError extends Error {}

function record(value: unknown, where: string): Json {
  if (typeof value !== "object" || value === null || Array.isArray(value)) {
    throw new RegisterError(`${where} must be an object`);
  }
  return value as Json;
}

function list(value: unknown, where: string): readonly unknown[] {
  if (!Array.isArray(value)) throw new RegisterError(`${where} must be an array`);
  return value;
}

function text(value: unknown, where: string): string {
  if (typeof value !== "string") throw new RegisterError(`${where} must be a string`);
  return value;
}

function count(value: unknown, where: string): number {
  if (typeof value !== "number" || !Number.isInteger(value) || value < 0) {
    throw new RegisterError(`${where} must be a non-negative integer`);
  }
  return value;
}

function strings(value: unknown, where: string): readonly string[] {
  return list(value, where).map((entry, index) => text(entry, `${where}[${String(index)}]`));
}

function optionalStrings(source: Json, key: string, where: string): readonly string[] | undefined {
  return source[key] === undefined ? undefined : strings(source[key], `${where}.${key}`);
}

function optional<K extends string>(
  key: K,
  value: readonly string[] | undefined,
): Partial<Record<K, readonly string[]>> {
  return value === undefined ? {} : ({ [key]: value } as Partial<Record<K, readonly string[]>>);
}

function moduleSpec(value: unknown, where: string): ModuleSpec {
  const source = record(value, where);
  const entrypoint = source.entrypoint;
  const folders = optionalStrings(source, "folders", where);
  const files = optionalStrings(source, "files", where);
  const mayImport = optionalStrings(source, "mayImport", where);
  const callers = optionalStrings(source, "callers", where);
  return {
    kind: text(source.kind, `${where}.kind`),
    layer: text(source.layer, `${where}.layer`),
    lifetime: text(source.lifetime, `${where}.lifetime`),
    responsibility: text(source.responsibility, `${where}.responsibility`),
    ...optional("folders", folders),
    ...optional("files", files),
    ...optional("mayImport", mayImport),
    ...(source.mayImportAny === true ? { mayImportAny: true } : {}),
    ...(callers === undefined ? {} : { callers }),
    entrypoint: entrypoint === null || entrypoint === undefined ? null : text(entrypoint, `${where}.entrypoint`),
  };
}

function moduleMap(value: unknown, where: string): Readonly<Record<string, ModuleSpec>> {
  const source = record(value, where);
  const out: Record<string, ModuleSpec> = {};
  for (const [key, entry] of Object.entries(source)) out[key] = moduleSpec(entry, `${where}.${key}`);
  return out;
}

function caps(value: unknown, where: string): CssSpec["caps"] {
  const source = record(value, where);
  const out: Record<string, { cap: number; phase: string }> = {};
  for (const [key, entry] of Object.entries(source)) {
    const item = record(entry, `${where}.${key}`);
    out[key] = { cap: count(item.cap, `${where}.${key}.cap`), phase: text(item.phase, `${where}.${key}.phase`) };
  }
  return out;
}

function globalCaps(value: unknown, where: string): Readonly<Record<string, CssCap>> {
  const source = record(value, where);
  const out: Record<string, CssCap> = {};
  for (const [key, entry] of Object.entries(source)) {
    const cap = record(entry, `${where}.${key}`);
    out[key] = {
      global: count(cap.global, `${where}.${key}.global`),
      important: count(cap.important, `${where}.${key}.important`),
      phase: text(cap.phase, `${where}.${key}.phase`),
    };
  }
  return out;
}

function colorExceptions(value: unknown, where: string): CssSpec["colorExceptions"] {
  const source = record(value, where);
  const out: Record<string, { literals: number; reason: string; phase: string }> = {};
  for (const [key, entry] of Object.entries(source)) {
    const item = record(entry, `${where}.${key}`);
    out[key] = {
      literals: count(item.literals, `${where}.${key}.literals`),
      reason: text(item.reason, `${where}.${key}.reason`),
      phase: text(item.phase, `${where}.${key}.phase`),
    };
  }
  return out;
}

function cssSpec(value: unknown, where: string): CssSpec {
  const source = record(value, where);
  return {
    globalEntrypoints: strings(source.globalEntrypoints, `${where}.globalEntrypoints`),
    vendorFiles: strings(source.vendorFiles, `${where}.vendorFiles`),
    defaultCap: count(source.defaultCap, `${where}.defaultCap`),
    caps: caps(source.caps, `${where}.caps`),
    globalCaps: globalCaps(source.globalCaps, `${where}.globalCaps`),
    colorExceptions: colorExceptions(source.colorExceptions, `${where}.colorExceptions`),
  };
}

function numericMap(value: unknown, where: string): Readonly<Record<string, number>> {
  const source = record(value, where);
  const out: Record<string, number> = {};
  for (const [key, entry] of Object.entries(source)) out[key] = count(entry, `${where}.${key}`);
  return out;
}

function floors(value: unknown, where: string): Register["floors"] {
  const source = record(value, where);
  const out: Record<string, Record<string, number>> = {};
  for (const [key, entry] of Object.entries(source)) out[key] = { ...numericMap(entry, `${where}.${key}`) };
  return out;
}

function ruleSpec(value: unknown, where: string): RuleSpec {
  const source = record(value, where);
  const rules = list(source.depcruiseRules, `${where}.depcruiseRules`).map((entry, index) => {
    const item = record(entry, `${where}.depcruiseRules[${String(index)}]`);
    return {
      name: text(item.name, `${where}.depcruiseRules.name`),
      severity: text(item.severity, `${where}.depcruiseRules.severity`),
    };
  });
  return {
    actionNames: strings(source.actionNames, `${where}.actionNames`),
    protocolOwner: text(source.protocolOwner, `${where}.protocolOwner`),
    protocolProductionOnly: source.protocolProductionOnly === true,
    restrictedAppImports: strings(source.restrictedAppImports, `${where}.restrictedAppImports`),
    restrictedAppGlobals: strings(source.restrictedAppGlobals, `${where}.restrictedAppGlobals`),
    restrictedUiImports: strings(source.restrictedUiImports, `${where}.restrictedUiImports`),
    clockGlobals: strings(source.clockGlobals ?? [], `${where}.clockGlobals`),
    booleanFactNames: strings(source.booleanFactNames ?? [], `${where}.booleanFactNames`),
    fakeTimerTests: strings(source.fakeTimerTests ?? [], `${where}.fakeTimerTests`),
    depcruiseRules: rules,
  };
}

function emittedSpec(value: unknown, where: string): EmittedSpec {
  const source = record(value, where);
  const order = list(source.order, `${where}.order`).map((entry, index) => {
    const item = record(entry, `${where}.order[${String(index)}]`);
    return {
      name: text(item.name, `${where}.order.name`),
      selector: text(item.selector, `${where}.order.selector`),
    };
  });
  return {
    order,
    important: count(source.important, `${where}.important`),
    zIndex: count(source.zIndex, `${where}.zIndex`),
    positioned: count(source.positioned, `${where}.positioned`),
    note: source.note === undefined ? "" : text(source.note, `${where}.note`),
  };
}

function debtList(value: unknown, where: string): readonly DebtEntry[] {
  return list(value, where).map((entry, index) => {
    const item = record(entry, `${where}[${String(index)}]`);
    const target = item.target;
    return {
      rule: text(item.rule, `${where}.rule`),
      file: text(item.file, `${where}.file`),
      ...(target === undefined ? {} : { target: text(target, `${where}.target`) }),
      count: count(item.count, `${where}.count`),
      phase: text(item.phase, `${where}.phase`),
      ...(item.reason === undefined ? {} : { reason: text(item.reason, `${where}.reason`) }),
    };
  });
}

export function parseRegister(value: unknown, where: string): Register {
  const source = record(value, where);
  return {
    version: count(source.version, `${where}.version`),
    source: text(source.source, `${where}.source`),
    modules: moduleMap(source.modules, `${where}.modules`),
    testSupport: strings(source.testSupport ?? [], `${where}.testSupport`),
    css: cssSpec(source.css, `${where}.css`),
    floors: floors(source.floors ?? {}, `${where}.floors`),
    rules: ruleSpec(source.rules, `${where}.rules`),
    emittedCss: emittedSpec(source.emittedCss, `${where}.emittedCss`),
    debt: debtList(source.debt ?? [], `${where}.debt`),
  };
}

export function readRegister(path: string): Register {
  const text = readFileSync(path, "utf8");
  return parseRegister(JSON.parse(text) as unknown, path);
}
