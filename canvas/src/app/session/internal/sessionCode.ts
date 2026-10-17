import { PAIRING_CODE_LENGTH } from "../contract";

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
