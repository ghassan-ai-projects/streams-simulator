import type { Outcome, TwinFailure } from "../../../app/ports";
import type { TwinReport, TwinRequest } from "../../../model/review";
import { twinFailure } from "./failures";
import type { Transport } from "./http";
import { action } from "./http";
import { twinReport } from "./twinReportDecoder";

function twinArgs({ run, domain, tolerance }: TwinRequest): Readonly<Record<string, unknown>> {
  return { run, domain, ...(tolerance.trim() === "" ? {} : { tolerance }) };
}

export function requestTwinReport(
  transport: Transport,
  request: TwinRequest,
): Promise<Outcome<TwinReport, TwinFailure>> {
  return action(transport, "run.twin_report", twinArgs(request), twinReport, twinFailure);
}
