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

type Json = Readonly<Record<string, unknown>>;

function record(value: unknown): Json {
  return typeof value === "object" && value !== null && !Array.isArray(value) ? (value as Json) : {};
}

function text(value: unknown): string {
  return typeof value === "string" ? value : "";
}

function tags(value: unknown): readonly string[] {
  return Array.isArray(value) ? value.map(text) : [];
}

export function searchLibrary(entries: readonly LibraryEntry[], query: string): readonly LibraryEntry[] {
  const terms = query.toLowerCase().split(/\s+/).filter(Boolean);
  return entries.filter((entry) => terms.every((term) => entry.words.includes(term)));
}

export function parseLibrary(listed: unknown): readonly LibraryEntry[] {
  const components = record(listed).components;
  return (Array.isArray(components) ? components : [])
    .map(record)
    .filter((entry) => entry.kind === "equipment" && text(entry.id) !== "")
    .map((entry) => ({
      component: text(entry.id),
      ref: `${text(entry.id)}@${typeof entry.version === "number" ? String(entry.version) : "1"}`,
      title: text(entry.title) || text(entry.id),
      words: [text(entry.title), text(entry.id), text(entry.class), ...tags(entry.tags)].join(" ").toLowerCase(),
    }))
    .sort((a, b) => a.title.localeCompare(b.title));
}

export function parsePresets(shown: unknown): Presets {
  const document = record(record(shown).document);
  const options = Object.entries(record(document.presets)).map(([name, raw]) => ({
    name,
    title: text(record(raw).title) || name,
  }));
  return { options, preferred: text(document.default_preset), description: text(document.description) };
}

export interface ExamplePlant {
  readonly id: string;
  readonly title: string;
  readonly description: string;
}

export function parseExamples(listed: unknown): readonly ExamplePlant[] {
  const examples = record(listed).examples;
  return (Array.isArray(examples) ? examples : [])
    .map(record)
    .filter((entry) => text(entry.id) !== "")
    .map((entry) => ({
      id: text(entry.id),
      title: text(entry.title) || text(entry.id),
      description: text(entry.description),
    }));
}
