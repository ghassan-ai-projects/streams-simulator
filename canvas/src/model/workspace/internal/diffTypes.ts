export interface DiffSide {
  readonly id: string;
  readonly version: string;
  readonly title: string;
  readonly digest: string;
  readonly source: { readonly kind: "file" | "run"; readonly label: string };
}

export interface ModelChange {
  readonly code: string;
  readonly subject: string;
  readonly section: string;
  readonly field: string;
  readonly before: unknown;
  readonly after: unknown;
  readonly sentence: string;
}

export interface InstanceSet {
  readonly count: number;
  readonly names: readonly string[];
  readonly idTemplates: readonly string[];
}

export interface DiffRelation {
  readonly type: string;
  readonly fromKind: string;
  readonly toKind: string;
  readonly links: readonly string[];
}

export interface ChangedKind {
  readonly kind: string;
  readonly status: "changed" | "added" | "removed";
  readonly before: InstanceSet;
  readonly after: InstanceSet;
  readonly relations: readonly DiffRelation[];
  readonly changes: readonly ModelChange[];
}

export interface DiffIdentity {
  readonly sameId: boolean;
  readonly sameVersion: boolean;
  readonly sameDigest: boolean;
  readonly versionKeptDigestChanged: boolean;
}

export interface ModelDiff {
  readonly a: DiffSide;
  readonly b: DiffSide;
  readonly identity: DiffIdentity;
  readonly kinds: readonly ChangedKind[];
  readonly domainChanges: readonly ModelChange[];
  readonly changes: number;
  readonly unchanged: readonly (readonly [string, number])[];
}
