import type { Outcome } from "../app/ports";

interface CannedIssue {
  readonly pointer?: string;
  readonly message: string;
  readonly keyword?: string;
  readonly member?: string;
}

interface CannedFailure {
  readonly status: number;
  readonly code: string;
  readonly message: string;
  readonly issues: readonly CannedIssue[];
  readonly result?: unknown;
}

export type Canned<T> = Outcome<T, CannedFailure>;

export function cannedFailure(code: string, status = 401, message = "refused"): CannedFailure {
  return { status, code, message, issues: [] };
}
