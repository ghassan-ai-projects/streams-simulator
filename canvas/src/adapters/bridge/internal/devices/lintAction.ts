import type { Transport } from "../transport/http";
import { action } from "../transport/http";
import type { FailureMapper } from "../transport/http";
import { lintReport } from "./twinDecoders";

export function lintAction<F>(transport: Transport, mapFailure: FailureMapper<F>) {
  return (path: string) =>
    action(transport, "domain.lint", { path }, lintReport, mapFailure, (problem) =>
      problem.code === "domain_lint" ? lintReport(problem.result) : null,
    );
}
