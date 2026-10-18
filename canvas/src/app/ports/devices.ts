import type { KindEntry } from "../../model/workspace";
import type { Diagnostic, FileContent, Outcome } from "./shared";

interface KindIssue {
  readonly pointer: string;
  readonly message: string;
  readonly keyword?: string;
  readonly member?: string;
}

export interface KindChecked {
  readonly kind: string;
  readonly digest: string;
  readonly metrics: number;
  readonly commands: number;
}

interface InstrumentFit {
  readonly ok: boolean;
  readonly code: "unit_incompatible" | "range_unstated" | "range_narrower" | "reading_not_number" | "";
  readonly member: string;
}

export interface LibraryInstrument {
  readonly ref: string;
  readonly title: string;
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
  readonly match: "exact" | "convertible" | "";
  readonly fit: InstrumentFit | null;
}

export interface InstrumentQuery {
  readonly unit: string;
  readonly min: number | null;
  readonly max: number | null;
}

interface DeviceListing {
  readonly kinds: readonly KindEntry[];
  readonly domains: readonly string[];
}

type SuggestedChoices = Readonly<Record<string, unknown>>;

export interface TwinMade {
  readonly id: string;
  readonly digest: string;
  readonly channels: number;
  readonly faults: number;
  readonly effectors: number;
  readonly implements: readonly string[];
}

export interface DomainChecked {
  readonly id: string;
  readonly digest: string;
}

interface LintFinding {
  readonly severity: string;
  readonly entity: string;
  readonly detail: string;
}

export interface LintReport {
  readonly errors: number;
  readonly warnings: number;
  readonly findings: readonly LintFinding[];
}

export type DeviceFailure =
  | { readonly kind: "missing" | "conflict" | "unavailable" | "unreadable"; readonly diagnostic: Diagnostic }
  | { readonly kind: "refused"; readonly diagnostic: Diagnostic; readonly issues: readonly KindIssue[] };

type Written = Outcome<{ readonly path: string; readonly etag: string }, DeviceFailure>;

export interface DevicesGateway {
  readFile(path: string): Promise<Outcome<FileContent, DeviceFailure>>;
  writeFile(path: string, content: string, etag: string): Promise<Written>;
  check(path: string): Promise<Outcome<KindChecked, DeviceFailure>>;
  checkContent(content: string, twin: boolean): Promise<Outcome<KindChecked, DeviceFailure>>;
  listing(): Promise<Outcome<DeviceListing, DeviceFailure>>;
  instruments(query: InstrumentQuery): Promise<Outcome<readonly LibraryInstrument[], DeviceFailure>>;
  suggest(kinds: readonly string[]): Promise<Outcome<SuggestedChoices, DeviceFailure>>;
  twin(input: {
    readonly kinds: readonly string[];
    readonly choices: string;
    readonly out: string;
    readonly etag: string;
  }): Promise<Outcome<TwinMade, DeviceFailure>>;
  validateDomain(path: string): Promise<Outcome<DomainChecked, DeviceFailure>>;
  lint(path: string): Promise<Outcome<LintReport, DeviceFailure>>;
}
