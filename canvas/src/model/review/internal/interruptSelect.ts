import { coverAround } from "./interruptCover";
import { allDevices, sameScope, scopeKey, scopeLabel } from "./interruptScope";
import { modeLine } from "./interruptWords";
import type { ControlScope, CoveredDevice, InterruptMode, InterruptState } from "./interrupts";
import type { SessionLive } from "./sessionLive";
import { deviceOf } from "./sessionLive";

export interface SeenState {
  readonly run: string;
  readonly state: InterruptState;
  readonly seenKey: string;
  readonly older: readonly InterruptState[];
}

export interface ScopeChoice {
  readonly key: string;
  readonly scope: ControlScope;
  readonly label: string;
  readonly present: boolean;
}

export interface ScopeStatus {
  readonly held: boolean;
  readonly latched: boolean;
  readonly approving: boolean;
  readonly coveredBy: string;
}

export interface ActiveScope {
  readonly key: string;
  readonly label: string;
  readonly mode: InterruptMode;
  readonly latched: boolean;
}

function sameState(a: InterruptState | null, b: InterruptState | null): boolean {
  return JSON.stringify(a) === JSON.stringify(b);
}

function olderThanReply(seen: SeenState, incoming: InterruptState | null): boolean {
  return !sameState(incoming, seen.state) && seen.older.some((older) => sameState(older, incoming));
}

export function newestInterrupts(
  newest: SessionLive | null,
  seen: SeenState | null,
  run: string | null,
): InterruptState | null {
  if (seen?.run !== run) return newest?.interrupts ?? null;
  if (newest === null || newest.key === seen.seenKey) return seen.state;
  return olderThanReply(seen, newest.interrupts) ? seen.state : newest.interrupts;
}

export function withShownInterrupts(
  live: SessionLive | null,
  newest: SessionLive | null,
  shown: InterruptState | null,
): SessionLive | null {
  if (live === null || live.interrupts === shown) return live;
  if (live.key !== newest?.key) return live;
  return { ...live, interrupts: shown };
}

export function interruptOf(live: SessionLive | null, instance: string): CoveredDevice | null {
  const device = deviceOf(live, instance);
  if (device === null) return null;
  return live?.interrupts?.devices.find((covered) => covered.device === device.device) ?? null;
}

function choiceOf(scope: ControlScope): ScopeChoice {
  return { key: scopeKey(scope), scope, label: scopeLabel(scope), present: true };
}

function sortedUnique(names: readonly string[]): readonly string[] {
  return [...new Set(names.filter((name) => name !== ""))].sort((a, b) => a.localeCompare(b));
}

function sessionChoices(live: SessionLive | null): readonly ScopeChoice[] {
  const devices = live?.devices ?? [];
  const boards = sortedUnique(devices.map((device) => device.board));
  const names = sortedUnique(devices.map((device) => device.device));
  return [
    choiceOf(allDevices),
    ...boards.map((id) => choiceOf({ kind: "board", id })),
    ...names.map((id) => choiceOf({ kind: "device", id })),
  ];
}

export function scopeChoices(live: SessionLive | null, picked: ControlScope): readonly ScopeChoice[] {
  const choices = sessionChoices(live);
  if (choices.some((choice) => sameScope(choice.scope, picked))) return choices;
  const label = `${scopeLabel(picked)} (not in this session)`;
  return [...choices, { key: scopeKey(picked), scope: picked, label, present: false }];
}

export function chosenChoice(choices: readonly ScopeChoice[], picked: ControlScope): ScopeChoice {
  const found = choices.find((choice) => sameScope(choice.scope, picked));
  return found ?? choices[0] ?? choiceOf(allDevices);
}

function coveredNote(state: InterruptState, scope: ControlScope): string {
  const holds = state.holds.map((hold) => hold.scope);
  const heldBy = coverAround(scope, holds, state).wider;
  const hold = state.holds.find((entry) => heldBy !== null && sameScope(entry.scope, heldBy));
  if (heldBy !== null) return `${scopeLabel(heldBy)}: ${modeLine({ mode: "hold", latched: hold?.latched === true })}`;
  const approvedBy = coverAround(scope, state.approve, state).wider;
  return approvedBy === null ? "" : `${scopeLabel(approvedBy)}: ${modeLine({ mode: "approve", latched: false })}`;
}

export function scopeStatus(state: InterruptState | null, scope: ControlScope): ScopeStatus {
  if (state === null) return { held: false, latched: false, approving: false, coveredBy: "" };
  const hold = state.holds.find((entry) => sameScope(entry.scope, scope)) ?? null;
  const approving = state.approve.some((entry) => sameScope(entry, scope));
  return { held: hold !== null, latched: hold?.latched === true, approving, coveredBy: coveredNote(state, scope) };
}

export function activeScopes(state: InterruptState | null): readonly ActiveScope[] {
  if (state === null) return [];
  const holds = state.holds.map((hold): ActiveScope => ({
    key: `hold:${scopeKey(hold.scope)}`,
    label: scopeLabel(hold.scope),
    mode: "hold",
    latched: hold.latched,
  }));
  const approves = state.approve.map((scope): ActiveScope => ({
    key: `approve:${scopeKey(scope)}`,
    label: scopeLabel(scope),
    mode: "approve",
    latched: false,
  }));
  return [...holds, ...approves];
}
