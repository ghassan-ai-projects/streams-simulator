import ts from "typescript";
import type { Register, Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { SourceFile } from "./sources.ts";
import { descendants, isExported, isProduction, lineOf } from "./sources.ts";

export function appUnknownRules(files: readonly SourceFile[]): Violation[] {
  const out: Violation[] = [];
  for (const file of files) {
    if (file.module !== "app" || !isProduction(file.path)) continue;
    for (const statement of file.source.statements) {
      if (!isExported(statement)) continue;
      for (const node of signatureTypes(statement)) {
        if (!mentionsUnknown(node)) continue;
        out.push(
          violation(
            "app-unknown-api",
            file.path,
            lineOf(file.source, node),
            "an exported application API exposes unknown",
          ),
        );
      }
    }
  }
  return out;
}

export function signatureTypes(statement: ts.Statement): ts.Node[] {
  if (ts.isFunctionDeclaration(statement)) return callableTypes(statement);
  if (ts.isInterfaceDeclaration(statement) || ts.isClassDeclaration(statement)) {
    return statement.members.flatMap((member) => memberTypes(member));
  }
  if (ts.isTypeAliasDeclaration(statement)) return [statement.type];
  if (ts.isVariableStatement(statement)) {
    return statement.declarationList.declarations.flatMap((declaration) => declaredTypes(declaration));
  }
  return [];
}

function declaredTypes(declaration: ts.VariableDeclaration): ts.Node[] {
  const annotated = declaration.type === undefined ? [] : [declaration.type];
  const value = declaration.initializer;
  if (value === undefined) return annotated;
  if (ts.isArrowFunction(value) || ts.isFunctionExpression(value)) return [...annotated, ...callableTypes(value)];
  return annotated;
}

function memberTypes(member: ts.TypeElement | ts.ClassElement): ts.Node[] {
  if (isHiddenMember(member)) return [];
  return memberSignature(member);
}

function isHiddenMember(member: ts.TypeElement | ts.ClassElement): boolean {
  if (!ts.canHaveModifiers(member)) return false;
  const modifiers = ts.getModifiers(member) ?? [];
  return modifiers.some(
    (modifier) => modifier.kind === ts.SyntaxKind.PrivateKeyword || modifier.kind === ts.SyntaxKind.ProtectedKeyword,
  );
}

function memberSignature(member: ts.TypeElement | ts.ClassElement): ts.Node[] {
  if (ts.isPropertyDeclaration(member) || ts.isPropertySignature(member))
    return member.type === undefined ? [] : [member.type];
  if (ts.isMethodDeclaration(member) || ts.isMethodSignature(member) || ts.isConstructorDeclaration(member))
    return callableTypes(member);
  return [];
}

function callableTypes(node: ts.SignatureDeclaration): ts.Node[] {
  const params = node.parameters.flatMap((parameter) => (parameter.type === undefined ? [] : [parameter.type]));
  return node.type === undefined ? params : [...params, node.type];
}

function mentionsUnknown(node: ts.Node): boolean {
  if (node.kind === ts.SyntaxKind.UnknownKeyword) return true;
  return node.getChildren().some(mentionsUnknown);
}

export function booleanParameterRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  const allowed = new Set(register.rules.booleanFactNames);
  for (const file of files) {
    if (!isProduction(file.path)) continue;
    for (const statement of file.source.statements) {
      for (const parameter of exportedParameterLists(statement)) {
        out.push(...booleanViolations(file, parameter, allowed));
      }
    }
  }
  return out;
}

function exportedParameterLists(statement: ts.Statement): ts.ParameterDeclaration[] {
  if (ts.isFunctionDeclaration(statement) && isExported(statement)) return [...statement.parameters];
  if (ts.isClassDeclaration(statement) && isExported(statement)) {
    return statement.members.flatMap((member) =>
      ts.isMethodDeclaration(member) || ts.isConstructorDeclaration(member) ? [...member.parameters] : [],
    );
  }
  if (ts.isVariableStatement(statement) && isExported(statement)) {
    return statement.declarationList.declarations.flatMap((declaration) => {
      const value = declaration.initializer;
      if (value === undefined) return [];
      return ts.isArrowFunction(value) || ts.isFunctionExpression(value) ? [...value.parameters] : [];
    });
  }
  return [];
}

function booleanViolations(
  file: SourceFile,
  parameter: ts.ParameterDeclaration,
  allowed: ReadonlySet<string>,
): Violation[] {
  if (!isBooleanParameter(parameter)) return [];
  const name = parameter.name.getText();
  if (allowed.has(name)) return [];
  return [
    violation(
      "boolean-flag-param",
      file.path,
      lineOf(file.source, parameter),
      `${name} selects behaviour; use a named option`,
    ),
  ];
}

function isNoiseIdentifier(node: ts.Identifier): boolean {
  return (
    (ts.isPropertyAccessExpression(node.parent) && node.parent.name === node) || ts.isTypeReferenceNode(node.parent)
  );
}

export function appClockRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  const clocks = new Set(register.rules.clockGlobals);
  for (const file of files) {
    if (file.module !== "app" || !isProduction(file.path)) continue;
    for (const node of descendants(file.source)) {
      if (!ts.isIdentifier(node) || !clocks.has(node.text)) continue;
      if (isNoiseIdentifier(node)) continue;
      out.push(
        violation(
          "app-clock-timer",
          file.path,
          lineOf(file.source, node),
          `app may not read ${node.text}; take it from a port`,
        ),
      );
    }
  }
  return out;
}

function isBooleanParameter(parameter: ts.ParameterDeclaration): boolean {
  if (parameter.type !== undefined) return parameter.type.kind === ts.SyntaxKind.BooleanKeyword;
  const value = parameter.initializer;
  return value !== undefined && (value.kind === ts.SyntaxKind.TrueKeyword || value.kind === ts.SyntaxKind.FalseKeyword);
}

export function testTimerRules(files: readonly SourceFile[], register: Register): Violation[] {
  const out: Violation[] = [];
  const allowed = new Set(register.rules.fakeTimerTests);
  for (const file of files) {
    if (!file.isTest || allowed.has(file.path)) continue;
    for (const node of descendants(file.source)) {
      if (!ts.isCallExpression(node)) continue;
      const name = node.expression.getText();
      if (name !== "setTimeout" && name !== "setInterval" && !/\.(setTimeout|setInterval)$/.test(name)) continue;
      out.push(
        violation(
          "test-timer",
          file.path,
          lineOf(file.source, node),
          "a test may not use a real timer; use a deferred driver",
        ),
      );
    }
  }
  return out;
}
