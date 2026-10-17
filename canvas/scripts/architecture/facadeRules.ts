import ts from "typescript";
import type { Register, Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { SourceFile } from "./sources.ts";
import { identifiersIn, isExported, lineOf } from "./sources.ts";
import { signatureTypes } from "./codeRules.ts";

const entrypointStatementKinds = new Set<ts.SyntaxKind>([
  ts.SyntaxKind.ImportDeclaration,
  ts.SyntaxKind.ExportDeclaration,
  ts.SyntaxKind.ExportAssignment,
  ts.SyntaxKind.FunctionDeclaration,
  ts.SyntaxKind.TypeAliasDeclaration,
  ts.SyntaxKind.InterfaceDeclaration,
  ts.SyntaxKind.VariableStatement,
]);
export function entrypointRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  const entrypoints = new Set(Object.values(register.modules).map((spec) => spec.entrypoint));
  for (const file of files) {
    if (!entrypoints.has(file.path) || !/index\.tsx?$/.test(file.path)) continue;
    out.push(...entrypointShape(file), ...internalTypeViolations(file));
  }
  return out;
}

function internalTypeViolations(file: SourceFile): Violation[] {
  const internal = internalNames(file);
  const out: Violation[] = [...internalReexports(file), ...internalTypeAliasExports(file)];
  if (internal.size === 0) return out;
  for (const statement of file.source.statements) {
    if (!isExported(statement)) continue;
    for (const type of signatureTypes(statement)) {
      for (const identifier of identifiersIn(type)) {
        if (!internal.has(identifier.text)) continue;
        out.push(
          violation(
            "facade-internal-type",
            file.path,
            lineOf(file.source, identifier),
            `${identifier.text} is internal to this module and may not appear in a public signature`,
          ),
        );
      }
    }
  }
  return out;
}

function internalReexports(file: SourceFile): Violation[] {
  const out: Violation[] = [];
  for (const statement of file.source.statements) {
    if (!ts.isExportDeclaration(statement) || statement.moduleSpecifier === undefined) continue;
    if (!ts.isStringLiteral(statement.moduleSpecifier)) continue;
    if (!statement.moduleSpecifier.text.includes("internal")) continue;
    if (valueReexport(statement)) continue;
    out.push(
      violation(
        "facade-internal-type",
        file.path,
        lineOf(file.source, statement),
        "a facade may not re-export its internals",
      ),
    );
  }
  return out;
}

function internalTypeAliasExports(file: SourceFile): Violation[] {
  const imported = internalTypeImports(file);
  const out: Violation[] = [];
  for (const statement of file.source.statements) {
    if (!ts.isExportDeclaration(statement) || statement.moduleSpecifier !== undefined) continue;
    const clause = statement.exportClause;
    if (clause === undefined || !ts.isNamedExports(clause)) continue;
    for (const specifier of clause.elements) {
      const local = specifier.propertyName?.text ?? specifier.name.text;
      if (!imported.has(local)) continue;
      out.push(
        violation(
          "facade-internal-type",
          file.path,
          lineOf(file.source, specifier),
          `${local} is imported from an internal module and may not be re-exported as a public type`,
        ),
      );
    }
  }
  return out;
}

function internalTypeImports(file: SourceFile): ReadonlySet<string> {
  const names = new Set<string>();
  for (const statement of file.source.statements) {
    if (!ts.isImportDeclaration(statement) || !ts.isStringLiteral(statement.moduleSpecifier)) continue;
    if (!statement.moduleSpecifier.text.includes("internal")) continue;
    addTypeImportNames(names, statement.importClause);
  }
  return names;
}

function addTypeImportNames(names: Set<string>, clause: ts.ImportClause | undefined): void {
  if (clause === undefined) return;
  if (clause.phaseModifier === ts.SyntaxKind.TypeKeyword && clause.name !== undefined) names.add(clause.name.text);
  const bindings = clause.namedBindings;
  if (bindings === undefined || !ts.isNamedImports(bindings)) return;
  for (const element of bindings.elements) {
    if (clause.phaseModifier === ts.SyntaxKind.TypeKeyword || element.isTypeOnly) {
      names.add(element.name.text);
    }
  }
}

function valueReexport(statement: ts.ExportDeclaration): boolean {
  if (statement.isTypeOnly || statement.exportClause === undefined || !ts.isNamedExports(statement.exportClause)) {
    return false;
  }
  return statement.exportClause.elements.every((specifier) => !specifier.isTypeOnly);
}

function internalNames(file: SourceFile): ReadonlySet<string> {
  const names = new Set<string>();
  for (const statement of file.source.statements) {
    if (!ts.isImportDeclaration(statement) || !ts.isStringLiteral(statement.moduleSpecifier)) continue;
    if (!statement.moduleSpecifier.text.includes("internal")) continue;
    const clause = statement.importClause;
    if (clause?.name !== undefined) names.add(clause.name.text);
    for (const name of namedImports(clause)) names.add(name);
  }
  return names;
}

function namedImports(clause: ts.ImportClause | undefined): readonly string[] {
  const bindings = clause?.namedBindings;
  if (bindings === undefined || !ts.isNamedImports(bindings)) return [];
  return bindings.elements.map((element) => element.name.text);
}

function entrypointShape(file: SourceFile): Violation[] {
  const out: Violation[] = [];
  for (const statement of file.source.statements) {
    out.push(
      ...wildcardExport(file, statement),
      ...namespaceExport(file, statement),
      ...facadeWorkflow(file, statement),
      ...controllerExport(file, statement),
      ...unexpectedStatement(file, statement),
    );
  }
  return out;
}

function wildcardExport(file: SourceFile, statement: ts.Statement): Violation[] {
  if (!ts.isExportDeclaration(statement) || statement.exportClause !== undefined) return [];
  return [violation("entrypoint-shape", file.path, lineOf(file.source, statement), "a facade may not use export *")];
}

function namespaceExport(file: SourceFile, statement: ts.Statement): Violation[] {
  if (!ts.isExportDeclaration(statement) || statement.exportClause === undefined) return [];
  if (!ts.isNamespaceExport(statement.exportClause)) return [];
  return [
    violation("entrypoint-shape", file.path, lineOf(file.source, statement), "a facade may not re-export a namespace"),
  ];
}

function facadeWorkflow(file: SourceFile, statement: ts.Statement): Violation[] {
  const body = exportedBody(statement);
  if (body === null || !holdsWorkflow(body)) return [];
  return [
    violation(
      "entrypoint-shape",
      file.path,
      lineOf(file.source, statement),
      "a facade function holds workflow, not construction",
    ),
  ];
}

function exportedBody(statement: ts.Statement): ts.ConciseBody | null {
  if (ts.isFunctionDeclaration(statement) && isExported(statement)) return statement.body ?? null;
  if (!ts.isVariableStatement(statement) || !isExported(statement)) return null;
  const values = statement.declarationList.declarations.map((declaration) => declaration.initializer);
  const callable = values.find(
    (value) => value !== undefined && (ts.isArrowFunction(value) || ts.isFunctionExpression(value)),
  );
  return callable !== undefined && (ts.isArrowFunction(callable) || ts.isFunctionExpression(callable))
    ? callable.body
    : null;
}

function controllerExport(file: SourceFile, statement: ts.Statement): Violation[] {
  if (!ts.isClassDeclaration(statement) || !isExported(statement)) return [];
  if (!/(Controller|Store)$/.test(statement.name?.text ?? "")) return [];
  return [
    violation(
      "entrypoint-shape",
      file.path,
      lineOf(file.source, statement),
      "a facade may not export a concrete controller",
    ),
  ];
}

function unexpectedStatement(file: SourceFile, statement: ts.Statement): Violation[] {
  if (entrypointStatementKinds.has(statement.kind)) return [];
  return [
    violation(
      "entrypoint-shape",
      file.path,
      lineOf(file.source, statement),
      "a facade holds declarations, not workflow",
    ),
  ];
}

function holdsWorkflow(body: ts.ConciseBody): boolean {
  if (!ts.isBlock(body)) return ts.isConditionalExpression(body) || holdsWorkflowNodes(body);
  return body.statements.some((statement) => holdsWorkflowNodes(statement));
}

function holdsWorkflowNodes(node: ts.Node): boolean {
  if (isControlFlow(node)) return true;
  let found = false;
  node.forEachChild((child) => {
    found ||= holdsWorkflowNodes(child);
  });
  return found;
}

function isControlFlow(node: ts.Node): boolean {
  return (
    ts.isIfStatement(node) ||
    ts.isForStatement(node) ||
    ts.isForOfStatement(node) ||
    ts.isWhileStatement(node) ||
    ts.isDoStatement(node) ||
    ts.isTryStatement(node) ||
    ts.isSwitchStatement(node)
  );
}
