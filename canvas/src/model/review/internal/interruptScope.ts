import type { ControlScope } from "./interrupts";

export const allDevices: ControlScope = { kind: "all", id: "" };

export function scopeWords(scope: ControlScope): string {
  if (scope.kind === "all") return "all devices";
  return `${scope.kind} ${scope.id}`;
}

export function scopeLabel(scope: ControlScope): string {
  const words = scopeWords(scope);
  return `${words.charAt(0).toUpperCase()}${words.slice(1)}`;
}

export function scopeKey(scope: ControlScope): string {
  return `${scope.kind}:${scope.id}`;
}

export function sameScope(a: ControlScope, b: ControlScope): boolean {
  return scopeKey(a) === scopeKey(b);
}
