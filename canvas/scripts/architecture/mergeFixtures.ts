import { mergeRegister } from "./mergeRegister.ts";

interface Case {
  readonly name: string;
  readonly base: unknown;
  readonly patch: unknown;
  readonly expected: unknown;
}

const cases: readonly Case[] = [
  {
    name: "objects merge recursively",
    base: { a: { b: 1, c: 2 } },
    patch: { a: { c: 3 } },
    expected: { a: { b: 1, c: 3 } },
  },
  { name: "null deletes a key", base: { a: 1, b: 2 }, patch: { a: null }, expected: { b: 2 } },
  { name: "null on a missing key is a no-op", base: { a: 1 }, patch: { z: null }, expected: { a: 1 } },
  { name: "arrays replace", base: { a: [1, 2, 3] }, patch: { a: [4] }, expected: { a: [4] } },
  { name: "scalars replace objects", base: { a: { b: 1 } }, patch: { a: 5 }, expected: { a: 5 } },
  { name: "objects replace scalars", base: { a: 5 }, patch: { a: { b: 1 } }, expected: { a: { b: 1 } } },
  { name: "new keys are added", base: { a: 1 }, patch: { b: { c: null, d: 2 } }, expected: { a: 1, b: { d: 2 } } },
  { name: "an empty patch keeps the base", base: { a: [1] }, patch: {}, expected: { a: [1] } },
  { name: "a non-object patch replaces the base", base: { a: 1 }, patch: [1], expected: [1] },
];

export function runMergeFixtures(): string[] {
  const failures: string[] = [];
  for (const entry of cases) {
    const snapshot = JSON.stringify(entry.base);
    const merged = JSON.stringify(mergeRegister(entry.base, entry.patch));
    if (merged !== JSON.stringify(entry.expected)) failures.push(`mergeRegister ${entry.name}: got ${merged}`);
    if (JSON.stringify(entry.base) !== snapshot) failures.push(`mergeRegister ${entry.name}: mutated the base`);
  }
  return failures;
}
