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

export interface PlantEntry {
  readonly id: string;
  readonly title: string;
  readonly domain: string | null;
  readonly composition: string | null;
  readonly instances: number;
  readonly runs: number;
  readonly lastRun: string;
}

export interface RunRow {
  readonly run: RunEntry;
  readonly plantTitle: string;
}
