export interface Definition {
  readonly path: string;
  readonly id: string;
  readonly title: string;
  readonly instances?: number;
}

export interface RunEntry {
  readonly worldId: string;
  readonly ended: boolean;
  readonly complete: boolean;
  readonly incompleteReason: string;
  readonly explore: boolean;
  readonly seed: number | null;
  readonly plant: string;
  readonly durationS: number;
  readonly recordedAt: string;
  readonly faults?: number;
}

export interface Workspace {
  readonly name: string;
  readonly domains: readonly Definition[];
  readonly compositions: readonly Definition[];
  readonly scenarios: readonly Definition[];
  readonly runs: readonly RunEntry[];
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
