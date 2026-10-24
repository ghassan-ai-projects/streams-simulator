import type { JobStatus } from "../../../model/launch";
import { jobFailureWords, lostWords, pacedRate, plannedS, reachedS, startFailureWords } from "../../../model/launch";
import type { LaunchFailure } from "../../ports";
import type { LaunchState } from "./launchState";
import type { TrackedRun } from "./trackedRun";

type Running = Extract<LaunchState, { phase: "running" }>;

export function startingState(run: TrackedRun): Running {
  const { origin, plan, startedMs } = run;
  return {
    phase: "running",
    origin,
    worldId: null,
    atS: 0,
    durationS: plan.durationS,
    plan,
    startedMs,
    atMs: startedMs,
    stopping: run.stopping,
    rate: null,
    calls: [],
  };
}

function runningState(run: TrackedRun, status: JobStatus, atMs: number): Running {
  const { origin, plan, startedMs, calls } = run;
  return {
    phase: "running",
    origin,
    worldId: status.worldId,
    atS: reachedS(status),
    durationS: plannedS(status),
    plan,
    startedMs,
    atMs,
    stopping: run.stopping || status.state === "stopping",
    rate: pacedRate(status),
    calls,
  };
}

function failedJob(run: TrackedRun, status: JobStatus): LaunchState {
  const message = status.failure === "" ? "The job failed and gave no reason." : status.failure;
  return {
    phase: "failed",
    origin: run.origin,
    failure: { kind: "rejected", step: "run", diagnostic: { displayCode: "job_failed", message } },
    words: jobFailureWords(status.failure),
    kept: status.worldId,
  };
}

export function followedState(run: TrackedRun, status: JobStatus, atMs: number): LaunchState {
  if (status.state === "failed") return failedJob(run, status);
  if (status.state !== "completed" && status.state !== "stopped") return runningState(run, status, atMs);
  const { origin, plan, calls } = run;
  return {
    phase: "done",
    origin,
    worldId: status.worldId,
    early: status.state === "stopped",
    plan,
    reachedS: reachedS(status),
    calls,
  };
}

export function refusedStart(run: TrackedRun, failure: LaunchFailure): LaunchState {
  const words = startFailureWords(failure.diagnostic.message);
  return { phase: "failed", origin: run.origin, failure, words, kept: null };
}

export function lostState(run: TrackedRun, failure: LaunchFailure): LaunchState {
  const words = failure.kind === "unavailable" ? lostWords : jobFailureWords(failure.diagnostic.message);
  return { phase: "failed", origin: run.origin, failure, words, kept: run.worldId };
}
