export interface ShowingMarks {
  readonly faults: ReadonlyMap<string, readonly string[]>;
  readonly effectors: ReadonlyMap<string, string>;
}
