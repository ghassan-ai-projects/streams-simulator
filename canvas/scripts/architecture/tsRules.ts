import ts from "typescript";
import type { ModuleSpec, Register, Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { SourceFile } from "./sources.ts";
import { descendants, importsOf, isProduction, lineOf } from "./sources.ts";

function allowListFor(file: SourceFile, register: Register): readonly string[] {
  const spec = register.modules[file.module];
  const base = spec?.mayImportAny === true ? Object.keys(register.modules) : (spec?.mayImport ?? []);
  return file.isTest ? [...base, ...register.testSupport] : [...base];
}

export function moduleRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    if (file.module === "") {
      out.push(violation("module-registered", file.path, 1, "file belongs to no registered module"));
      continue;
    }
    for (const imported of importsOf(file, register)) {
      if (aliasSpecifier(imported.specifier)) {
        out.push(
          violation(
            "import-resolvable",
            file.path,
            imported.line,
            `${imported.specifier} is not a relative path the scan can follow`,
          ),
        );
        continue;
      }
      if (imported.module === null || imported.module === file.module) continue;
      if (!allowListFor(file, register).includes(imported.module)) {
        out.push(
          violation(
            "import-direction",
            file.path,
            imported.line,
            `${file.module} may not import ${imported.module}`,
            imported.target ?? imported.specifier,
          ),
        );
      }
      out.push(...entrypointRule(file, register, imported));
    }
  }
  return out;
}

function aliasSpecifier(specifier: string): boolean {
  return /^(?:@\/|~\/|src\/)/.test(specifier);
}

function entrypointRule(
  file: SourceFile,
  register: Register,
  imported: {
    readonly module: string | null;
    readonly target: string | null;
    readonly specifier: string;
    readonly line: number;
  },
): Violation[] {
  if (imported.target === null || imported.module === null || imported.module === file.module) return [];
  if (imported.target.includes("/internal/")) {
    return [
      violation("private-path", file.path, imported.line, `${imported.module} internals are private`, imported.target),
    ];
  }
  const entrypoint = register.modules[imported.module]?.entrypoint;
  if (entrypoint !== null && entrypoint !== undefined && imported.target !== entrypoint) {
    return [
      violation(
        "module-entrypoint",
        file.path,
        imported.line,
        `${imported.module} is imported through ${entrypoint}`,
        imported.target,
      ),
    ];
  }
  return [];
}

export function targetRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const [path, target] of Object.entries(register.targets)) {
    if (register.targetModules[target] === undefined) {
      out.push(violation("target-owner", path, 1, `target ${target} is not a registered target module`));
    }
  }
  for (const file of files) {
    if (targetOf(file.path, register) === null) {
      out.push(violation("target-missing", file.path, 1, "the file has no target module assignment"));
    }
  }
  return out;
}

function targetOf(path: string, register: Register): string | null {
  if (register.targets[path] !== undefined) return register.targets[path] ?? null;
  const folders = Object.keys(register.targets)
    .filter((folder) => path.startsWith(`${folder}/`))
    .sort((a, b) => b.length - a.length);
  return folders[0] === undefined ? null : (register.targets[folders[0]] ?? null);
}

/** The accepted dependency contract, enforced for every module that has been extracted. */
/** Foundations and the composition root may import whatever they wire; only product modules are bounded. */
function bounded(owner: ModuleSpec): boolean {
  return owner.kind !== "composition" && owner.kind !== "test-support";
}

export function targetEdgeRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    const from = targetOf(file.path, register);
    const contract = from === null ? undefined : contractFor(from, register, file);
    if (from === null || contract === undefined) continue;
    out.push(...edgeViolations(file, from, contract, register));
  }
  return out;
}

function contractFor(from: string, register: Register, file: SourceFile): ModuleSpec | undefined {
  const owner = register.modules[from];
  if (owner === undefined || !bounded(owner) || file.isTest) return undefined;
  return register.targetModules[from];
}

function edgeViolations(file: SourceFile, from: string, contract: ModuleSpec, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const imported of importsOf(file, register)) {
    if (imported.target === null) continue;
    const to = targetOf(imported.target, register);
    if (to === null || to === from || register.testSupport.includes(to)) continue;
    if ((contract.mayImport ?? []).includes(to)) continue;
    out.push(violation("target-edge", file.path, imported.line, `${from} may not import ${to}`, imported.target));
  }
  return out;
}

export function callerRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  const graph = callGraph(files, register);
  for (const [name, spec] of Object.entries(register.modules)) {
    if (spec.callers === undefined) continue;
    const callers = graph.get(name) ?? new Set<string>();
    for (const caller of callers) {
      if (!(spec.callers ?? []).includes(caller)) {
        out.push(violation("callers-recorded", name, 1, `${caller} calls this module but is not recorded`));
      }
    }
    for (const recorded of spec.callers ?? []) {
      if (!callers.has(recorded)) {
        out.push(
          violation("callers-recorded", name, 1, `${recorded} is recorded as a caller but does not import this module`),
        );
      }
    }
  }
  return out;
}

function callGraph(files: readonly SourceFile[], register: Register): Map<string, Set<string>> {
  const graph = new Map<string, Set<string>>();
  for (const file of files) {
    for (const imported of importsOf(file, register)) {
      if (imported.module === null || imported.module === file.module) continue;
      graph.set(imported.module, (graph.get(imported.module) ?? new Set<string>()).add(file.module));
    }
  }
  return graph;
}

export function protocolRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    if (register.rules.protocolProductionOnly && !isProduction(file.path)) continue;
    if (file.module === register.rules.protocolOwner) continue;
    out.push(...protocolViolations(file, register));
  }
  return out;
}

function protocolViolations(file: SourceFile, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const node of descendants(file.source)) {
    if (ts.isStringLiteralLike(node)) out.push(...protocolViolation(file, register, node.text, node));
    if (!ts.isTemplateExpression(node)) continue;
    out.push(...protocolViolation(file, register, node.head.text, node));
    for (const span of node.templateSpans)
      out.push(...protocolViolation(file, register, span.literal.text, span.literal));
  }
  return out;
}

function protocolViolation(file: SourceFile, register: Register, value: string, node: ts.Node): Violation[] {
  const name = protocolName(value, register);
  if (name === null) return [];
  return [
    violation(
      "protocol-owner",
      file.path,
      lineOf(file.source, node),
      `${name} belongs in ${register.rules.protocolOwner}`,
    ),
  ];
}

function protocolName(value: string, register: Register): string | null {
  if (register.rules.actionNames.includes(value)) return `simulator action ${value}`;
  return value.startsWith("/api/") ? `bridge route ${value}` : null;
}

export function restrictedRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    if (file.module === "ui") out.push(...uiImportViolations(file, register));
    if (file.module !== "app" || !isProduction(file.path)) continue;
    for (const imported of importsOf(file, register)) {
      if (restricted(imported.specifier, register.rules.restrictedAppImports)) {
        out.push(
          violation("app-restricted-import", file.path, imported.line, `app may not import ${imported.specifier}`),
        );
      }
    }
    out.push(...globalViolations(file, register));
  }
  return out;
}

function restricted(specifier: string, names: readonly string[]): boolean {
  return names.some((name) => specifier === name || specifier.startsWith(`${name}/`));
}

function uiImportViolations(file: SourceFile, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const imported of importsOf(file, register)) {
    for (const forbidden of register.rules.restrictedUiImports) {
      if (!imported.specifier.includes(forbidden)) continue;
      out.push(
        violation("ui-restricted-import", file.path, imported.line, `the UI may not import ${imported.specifier}`),
      );
    }
  }
  return out;
}

function globalViolations(file: SourceFile, register: Register): Violation[] {
  const out: Violation[] = [];
  for (const node of descendants(file.source)) {
    if (!ts.isIdentifier(node)) continue;
    if (!register.rules.restrictedAppGlobals.includes(node.text)) continue;
    if (isDeclarationName(node)) continue;
    out.push(violation("app-restricted-global", file.path, lineOf(file.source, node), `app may not use ${node.text}`));
  }
  return out;
}

function isDeclarationName(node: ts.Identifier): boolean {
  const parent = node.parent;
  return isMemberName(parent, node) || isTypeName(parent, node);
}

function isMemberName(parent: ts.Node, node: ts.Identifier): boolean {
  if (ts.isPropertyAccessExpression(parent) || ts.isPropertyAssignment(parent) || ts.isPropertySignature(parent)) {
    return parent.name === node;
  }
  return false;
}

function isTypeName(parent: ts.Node, node: ts.Identifier): boolean {
  if (ts.isQualifiedName(parent)) return parent.left === node || parent.right === node;
  return ts.isTypeReferenceNode(parent) && parent.typeName === node;
}

export function doubleAssertionRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    if (register.rules.protocolProductionOnly && !isProduction(file.path)) continue;
    for (const node of descendants(file.source)) {
      if (!ts.isAsExpression(node)) continue;
      if (ts.isAsExpression(node.expression) && node.expression.type.kind === ts.SyntaxKind.UnknownKeyword) {
        out.push(
          violation(
            "double-assertion",
            file.path,
            lineOf(file.source, node),
            "a production double assertion bypasses decoding",
          ),
        );
      }
    }
  }
  return out;
}
