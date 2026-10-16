import ts from "typescript";
import type { Register, Violation } from "./model.ts";
import { violation } from "./model.ts";
import type { SourceFile } from "./sources.ts";
import { descendants, isProduction, lineOf } from "./sources.ts";

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
