export interface Definition {
  readonly path: string;
  readonly id: string;
  readonly title: string;
}

export interface RunEntry {
  readonly worldId: string;
  readonly ended: boolean;
  readonly explore: boolean;
}

export interface Workspace {
  readonly name: string;
  readonly domains: readonly Definition[];
  readonly compositions: readonly Definition[];
  readonly scenarios: readonly Definition[];
  readonly runs: readonly RunEntry[];
}

interface Issue {
  readonly pointer: string;
  readonly message: string;
}

export interface BridgeFailure {
  readonly status: number;
  readonly code: string;
  readonly message: string;
  readonly issues: readonly Issue[];
  readonly result?: unknown;
}

export interface FileContent {
  readonly path: string;
  readonly content: string;
  readonly etag: string;
}

export type Outcome<T> =
  { readonly ok: true; readonly value: T } | { readonly ok: false; readonly failure: BridgeFailure };

export function success<T>(value: T): Outcome<T> {
  return { ok: true, value };
}

export function failed<T>(failure: BridgeFailure): Outcome<T> {
  return { ok: false, failure };
}

type Json = Record<string, unknown>;

function isRecord(value: unknown): value is Json {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function list(value: unknown): readonly unknown[] {
  return Array.isArray(value) ? value : [];
}

function definition(value: unknown): Definition {
  const entry = isRecord(value) ? value : {};
  return { path: text(entry.path), id: text(entry.id), title: text(entry.title) };
}

function runEntry(value: unknown): RunEntry {
  const entry = isRecord(value) ? value : {};
  return { worldId: text(entry.world_id), ended: entry.ended === true, explore: entry.explore === true };
}

export function parseWorkspace(value: unknown): Workspace {
  const body = isRecord(value) ? value : {};
  return {
    name: text(body.name),
    domains: list(body.domains).map(definition),
    compositions: list(body.compositions).map(definition),
    scenarios: list(body.scenarios).map(definition),
    runs: list(body.runs).map(runEntry),
  };
}

function issue(value: unknown): Issue {
  const entry = isRecord(value) ? value : {};
  return { pointer: text(entry.pointer), message: text(entry.message) };
}

export function parseFailure(status: number, value: unknown): BridgeFailure {
  const body = isRecord(value) && isRecord(value.error) ? value.error : {};
  const code = text(body.code) || "bridge_failed";
  const message = text(body.message) || `The bridge answered ${String(status)}.`;
  return { status, code, message, issues: list(body.issues).map(issue), result: body.result ?? null };
}

export function parseFile(value: unknown): FileContent {
  const body = isRecord(value) ? value : {};
  return { path: text(body.path), content: text(body.content), etag: text(body.etag) };
}

export function parsePaired(value: unknown): boolean {
  return isRecord(value) && value.paired === true;
}

export const PAIRING_CODE_LENGTH = 12;
const CODE_ALPHABET = /^[A-Z2-7]*$/;

export function normalisePairingCode(raw: string): string {
  return raw
    .toUpperCase()
    .replace(/[^A-Z2-7]/g, "")
    .slice(0, PAIRING_CODE_LENGTH);
}

export function pairingCodeComplete(code: string): boolean {
  return code.length === PAIRING_CODE_LENGTH && CODE_ALPHABET.test(code);
}

export function pairingMessage(failure: BridgeFailure): string {
  switch (failure.code) {
    case "wrong_code":
      return "That code did not match. Check the code in your terminal and try again.";
    case "pairing_locked":
      return "Pairing is locked after five wrong codes. Restart streamsim canvas for a new code.";
    case "pairing_used":
      return "This code was already used. Restart streamsim canvas to pair another browser.";
    default:
      return failure.message;
  }
}

export function pairingBlocked(failure: BridgeFailure | null): boolean {
  return failure !== null && (failure.code === "pairing_locked" || failure.code === "pairing_used");
}
