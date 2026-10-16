import ts from "typescript";
import type { Register, Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { SourceFile } from "./sources.ts";
import { descendants, identifiersIn, isExported, isProduction, lineOf } from "./sources.ts";
import { signatureTypes } from "./codeRules.ts";

const staticStyleKeys = new Set([
  "color",
  "background",
  "backgroundColor",
  "border",
  "borderColor",
  "borderRadius",
  "boxShadow",
  "margin",
  "marginTop",
  "marginRight",
  "marginBottom",
  "marginLeft",
  "padding",
  "paddingTop",
  "paddingRight",
  "paddingBottom",
  "paddingLeft",
  "gap",
  "font",
  "fontFamily",
  "fontSize",
  "fontWeight",
  "lineHeight",
  "letterSpacing",
  "textAlign",
  "display",
  "position",
  "top",
  "left",
  "right",
  "bottom",
  "width",
  "height",
  "flexDirection",
  "justifyContent",
  "alignItems",
]);

const entrypointStatementKinds = new Set<ts.SyntaxKind>([
  ts.SyntaxKind.ImportDeclaration,
  ts.SyntaxKind.ExportDeclaration,
  ts.SyntaxKind.ExportAssignment,
  ts.SyntaxKind.FunctionDeclaration,
  ts.SyntaxKind.TypeAliasDeclaration,
  ts.SyntaxKind.InterfaceDeclaration,
  ts.SyntaxKind.VariableStatement,
]);

export function inlineStyleRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    if (register.rules.protocolProductionOnly && !isProduction(file.path)) continue;
    for (const node of descendants(file.source)) {
      const object = styleObject(node);
      if (object !== null) out.push(...staticStyleViolations(file, object));
    }
  }
  return out;
}

function styleObject(node: ts.Node): ts.ObjectLiteralExpression | null {
  if (!ts.isJsxAttribute(node) || node.name.getText() !== "style") return null;
  const initializer = node.initializer;
  if (initializer === undefined || !ts.isJsxExpression(initializer)) return null;
  const expression = initializer.expression;
  return expression !== undefined && ts.isObjectLiteralExpression(expression) ? expression : null;
}

function staticStyleViolations(file: SourceFile, object: ts.ObjectLiteralExpression): Violation[] {
  const out: Violation[] = [];
  for (const property of object.properties) {
    if (!ts.isPropertyAssignment(property)) continue;
    const name = property.name.getText().replace(/["']/g, "");
    if (!staticStyleKeys.has(name)) continue;
    if (!isLiteralValue(property.initializer)) continue;
    out.push(
      violation(
        "inline-style-static",
        file.path,
        lineOf(file.source, property),
        `${name} is a static style declaration; it belongs in the owning stylesheet`,
      ),
    );
  }
  return out;
}

function isLiteralValue(node: ts.Expression): boolean {
  return ts.isStringLiteral(node) || ts.isNumericLiteral(node) || ts.isNoSubstitutionTemplateLiteral(node);
}

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
  const out: Violation[] = internalReexports(file);
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
