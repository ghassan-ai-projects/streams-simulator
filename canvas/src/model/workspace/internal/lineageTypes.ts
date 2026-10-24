export type SourceStatus = "in_sync" | "out_of_sync" | "does_not_compose" | "does_not_scaffold" | "kind_files_missing";

type RouteCode =
  | "edit_composition"
  | "recompose"
  | "open_device_behavior"
  | "rescaffold_twin"
  | "create_from_example"
  | "copy_domain_file";

type EditReason = "composition" | "composition_out_of_sync" | "twin" | "twin_out_of_sync" | "example" | "hand_authored";

export interface LineageSource {
  readonly file: string;
  readonly id: string;
  readonly version: string;
  readonly status: SourceStatus;
  readonly reason: string;
}

export interface LineageRun {
  readonly id: string;
  readonly mode: string;
  readonly recorded: boolean;
  readonly domainVersion: string;
  readonly domainDigest: string;
  readonly sameAsCurrent: boolean;
}

export interface PlantKind {
  readonly kind: string;
  readonly found: boolean;
  readonly file: string;
}

export interface Editable {
  readonly verdict: "editable" | "read_only";
  readonly reason: EditReason;
  readonly route: { readonly code: RouteCode; readonly sentence: string };
}

export interface LineagePlant {
  readonly id: string;
  readonly version: string;
  readonly title: string;
  readonly digest: string;
  readonly file: string;
  readonly composition: LineageSource | null;
  readonly twin: LineageSource | null;
  readonly kinds: readonly PlantKind[];
  readonly runs: readonly LineageRun[];
  readonly editable: Editable;
}

interface TwinInclusion {
  readonly twin: string;
  readonly file: string;
  readonly kinds: number;
}

export interface DeviceLineage {
  readonly kind: string;
  readonly file: string;
  readonly twinsIncluding: readonly TwinInclusion[];
  readonly hasOwnTwin: boolean;
  readonly implementedBy: readonly string[];
}

export interface Unreadable {
  readonly file: string;
  readonly code: string;
  readonly message: string;
}

export interface Lineage {
  readonly plants: readonly LineagePlant[];
  readonly devices: readonly DeviceLineage[];
  readonly unreadable: readonly Unreadable[];
}
