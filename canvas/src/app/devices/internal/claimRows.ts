import {
  calibrationWords,
  claimSummary,
  hardwareNote,
  hardwareWords,
  issueText,
  standInWords,
} from "../../../model/devices";
import type { Claims } from "../../../model/devices";
import { counted } from "../../../model/workspace";
import type { DeviceFailure, LintReport } from "../../ports";
import type { CheckRow } from "./checkRow";
import type { Built, TwinDraft } from "./twinDraft";
import { builtOf, verdictOf } from "./twinSelectors";

const shown = 6;

type Claim = Pick<CheckRow, "id" | "state" | "status" | "text">;

function claim(head: Claim, notes: readonly string[] = [], details: readonly string[] = []): CheckRow {
  return { ...head, notes, details };
}

export function refusal(failure: DeviceFailure): {
  readonly notes: readonly string[];
  readonly details: readonly string[];
} {
  if (failure.kind !== "refused" || failure.issues.length === 0) {
    return { notes: [failure.diagnostic.message], details: [] };
  }
  return { notes: failure.issues.map(issueText), details: failure.issues.map((issue) => issue.message) };
}

export function interfaceRow(built: Built | null): CheckRow {
  const id = "implements";
  if (built === null) return claim({ id, state: "idle", status: "Not checked", text: "Interface: not checked yet" });
  const { implemented } = built;
  if (implemented.kind === "checking") {
    const text = "Interface: checking that the twin matches its device…";
    return claim({ id, state: "pending", status: "Checking", text });
  }
  if (implemented.kind === "passed") {
    return claim({ id, state: "ok", status: "Passed", text: "Interface: the twin matches its device" });
  }
  const { notes, details } = refusal(implemented.failure);
  const text = "Interface: the twin does not match its device";
  return claim({ id, state: "bad", status: "Failed", text }, notes, details);
}

function findings(report: LintReport): readonly string[] {
  const lines = report.findings.slice(0, shown).map((finding) => `${finding.entity}: ${finding.detail}`);
  const more = report.findings.length - shown;
  return more > 0 ? [...lines, `…and ${String(more)} more`] : lines;
}

function plausibleRow(report: LintReport): CheckRow {
  const id = "lint";
  const notes = findings(report);
  if (report.errors > 0) {
    const text = `Plausibility: the check found ${counted(report.errors, "error")}`;
    return claim({ id, state: "bad", status: "Failed", text }, notes);
  }
  const text = "Plausibility: readings stay plausible in a nominal day";
  if (report.warnings > 0) {
    return claim({ id, state: "warn", status: `Passed with ${counted(report.warnings, "warning")}`, text }, notes);
  }
  return claim({ id, state: "ok", status: "Passed", text }, notes);
}

export function plausibilityRow(built: Built | null): CheckRow {
  const id = "lint";
  if (built === null) return claim({ id, state: "idle", status: "Not checked", text: "Plausibility: not checked yet" });
  if (built.lint.kind === "unavailable") {
    return claim({ id, state: "idle", status: "Not run", text: "Plausibility: the check could not be run" });
  }
  if (built.lint.kind === "running") {
    const text = "Plausibility: checking that readings stay plausible in a nominal day…";
    return claim({ id, state: "pending", status: "Checking", text });
  }
  return plausibleRow(built.lint.report);
}

export function calibrationRow(): CheckRow {
  const head: Claim = { id: "calibration", state: "neutral", status: "Not performed", text: calibrationWords };
  return claim(head, [standInWords]);
}

export function hardwareRow(): CheckRow {
  return claim({ id: "hardware", state: "neutral", status: "Not run", text: hardwareWords }, [hardwareNote]);
}

function interfaceClaim({ implemented }: Built): Claims["interface"] {
  if (implemented.kind === "passed") return "passed";
  return implemented.kind === "failed" ? "failed" : "running";
}

function plausibilityClaim({ lint }: Built): Pick<Claims, "plausibility" | "warnings" | "errors"> {
  if (lint.kind === "unavailable") return { plausibility: "unavailable", warnings: 0, errors: 0 };
  if (lint.kind === "running") return { plausibility: "running", warnings: 0, errors: 0 };
  const { errors, warnings } = lint.report;
  const plausibility = errors > 0 ? "failed" : warnings > 0 ? "warnings" : "passed";
  return { plausibility, warnings, errors };
}

export function claimsOf(built: Built): Claims {
  return { interface: interfaceClaim(built), ...plausibilityClaim(built) };
}

export function checkSummary(draft: TwinDraft): string | null {
  const built = builtOf(draft);
  if (built === null || verdictOf(built) === "pending") return null;
  return claimSummary(claimsOf(built));
}
