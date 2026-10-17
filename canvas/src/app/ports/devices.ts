import type { Diagnostic, FileContent, Outcome } from "./shared";

interface KindIssue {
  readonly pointer: string;
  readonly message: string;
}

export interface KindChecked {
  readonly kind: string;
  readonly digest: string;
  readonly metrics: number;
  readonly commands: number;
}

export type DeviceFailure =
  | { readonly kind: "missing" | "conflict" | "unavailable" | "unreadable"; readonly diagnostic: Diagnostic }
  | { readonly kind: "refused"; readonly diagnostic: Diagnostic; readonly issues: readonly KindIssue[] };

export interface DevicesGateway {
  readFile(path: string): Promise<Outcome<FileContent, DeviceFailure>>;
  writeFile(
    path: string,
    content: string,
    etag: string,
  ): Promise<Outcome<{ readonly path: string; readonly etag: string }, DeviceFailure>>;
  check(path: string): Promise<Outcome<KindChecked, DeviceFailure>>;
}
