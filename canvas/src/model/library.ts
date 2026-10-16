export interface LibraryEntry {
  readonly component: string;
  readonly ref: string;
  readonly title: string;
  readonly words: string;
}

interface Preset {
  readonly name: string;
  readonly title: string;
}

export interface Presets {
  readonly options: readonly Preset[];
  readonly preferred: string;
  readonly description: string;
}

export function searchLibrary(entries: readonly LibraryEntry[], query: string): readonly LibraryEntry[] {
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
  return entries.filter((entry) => terms.every((term) => entry.words.includes(term)));
}

export interface ExamplePlant {
  readonly id: string;
  readonly title: string;
  readonly description: string;
}
