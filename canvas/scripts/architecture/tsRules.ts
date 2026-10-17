import ts from "typescript";
import type { Register, Violation } from "./model.ts";
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

export function umbrellaRules(files: readonly SourceFile[], register: Register): Violation[] {
  const umbrellas = umbrellaModules(register);
  return files
    .filter((file) => umbrellas.has(file.module))
    .map((file) =>
      violation("umbrella-file", file.path, 1, `${file.module} only groups modules; the file belongs in one`),
    );
}

function umbrellaModules(register: Register): Set<string> {
  const entries = Object.entries(register.modules);
  const groups = entries.filter(([, outer]) =>
    entries.some(([, inner]) => nestsUnder(inner.folders ?? [], outer.folders ?? [])),
  );
  return new Set(groups.map(([name]) => name));
}

function nestsUnder(inner: readonly string[], outer: readonly string[]): boolean {
  return inner.some((folder) => outer.some((own) => folder.startsWith(`${own}/`)));
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
