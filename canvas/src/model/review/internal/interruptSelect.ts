import { modeLine, scopeKey, scopeLabel } from "./interruptWords";
import type { ControlScope, CoveredDevice, InterruptMode, InterruptState } from "./interrupts";
import type { SessionLive } from "./sessionLive";
import { deviceOf } from "./sessionLive";

export interface SeenState {
  readonly state: InterruptState;
  readonly seenKey: string;
}

export interface ScopeChoice {
  readonly key: string;
  readonly scope: ControlScope;
  readonly label: string;
}

export interface ScopeStatus {
  readonly held: boolean;
  readonly latched: boolean;
  readonly approving: boolean;
}

export interface ActiveScope {
  readonly key: string;
  readonly text: string;
  readonly mode: InterruptMode;
  readonly latched: boolean;
}

export const allDevices: ControlScope = { kind: "all", id: "" };

export function newestInterrupts(newest: SessionLive | null, seen: SeenState | null): InterruptState | null {
  if (seen !== null && (newest === null || newest.key === seen.seenKey)) return seen.state;
  return newest?.interrupts ?? null;
}

export function interruptOf(live: SessionLive | null, instance: string): CoveredDevice | null {
  const device = deviceOf(live, instance);
  if (device === null) return null;
  return live?.interrupts?.devices.find((covered) => covered.device === device.device) ?? null;
}

function choiceOf(scope: ControlScope): ScopeChoice {
  return { key: scopeKey(scope), scope, label: scopeLabel(scope) };
}

function sortedUnique(names: readonly string[]): readonly string[] {
  return [...new Set(names.filter((name) => name !== ""))].sort((a, b) => a.localeCompare(b));
}

export function chosenScope(choices: readonly ScopeChoice[], key: string): ControlScope {
  return choices.find((choice) => choice.key === key)?.scope ?? allDevices;
}

export function scopeChoices(live: SessionLive | null): readonly ScopeChoice[] {
  const devices = live?.devices ?? [];
  const boards = sortedUnique(devices.map((device) => device.board));
  const names = sortedUnique(devices.map((device) => device.device));
  return [
    choiceOf(allDevices),
    ...boards.map((id) => choiceOf({ kind: "board", id })),
    ...names.map((id) => choiceOf({ kind: "device", id })),
  ];
}

export function scopeStatus(state: InterruptState | null, scope: ControlScope): ScopeStatus {
  const key = scopeKey(scope);
  const hold = state?.holds.find((entry) => scopeKey(entry.scope) === key) ?? null;
  const approving = (state?.approve ?? []).some((entry) => scopeKey(entry) === key);
  return { held: hold !== null, latched: hold?.latched === true, approving };
}

export function activeScopes(state: InterruptState | null): readonly ActiveScope[] {
  if (state === null) return [];
  const holds = state.holds.map((hold): ActiveScope => ({
    key: `hold:${scopeKey(hold.scope)}`,
    text: `${scopeLabel(hold.scope)}: ${modeLine({ mode: "hold", latched: hold.latched })}`,
    mode: "hold",
    latched: hold.latched,
  }));
  const approves = state.approve.map((scope): ActiveScope => ({
    key: `approve:${scopeKey(scope)}`,
    text: `${scopeLabel(scope)}: ${modeLine({ mode: "approve", latched: false })}`,
    mode: "approve",
    latched: false,
  }));
  return [...holds, ...approves];
}
