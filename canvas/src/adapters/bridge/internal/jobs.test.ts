import { describe, expect, it } from "vitest";
import { bridgeTransport } from "../../../test/bridgeTransport";
import { respond } from "../../../test/fakeWire";
import { bridgeAdapter } from "..";
import { jobList, jobStatus } from "./jobs";

const wire = {
  world_id: "w-9",
  run_id: "r-9",
  domain: "factory",
  seed: 3,
  state: "running",
  start_ns: 100e9,
  reached_ns: 160e9,
  requested_end_ns: 1900e9,
  rate: 60,
  explore: true,
  calls: [
    { at_s: 20, entity_id: "CH-1", effector: "chiller__start", accepted: true, effect_eta_ns: 5e9 },
    { at_s: 30, entity_id: "CH-1", effector: "chiller__stop", accepted: false, reason: "interlock_refused" },
  ],
};

const status = {
  worldId: "w-9",
  runId: "r-9",
  domain: "factory",
  seed: 3,
  state: "running",
  startNs: 100e9,
  reachedNs: 160e9,
  requestedEndNs: 1900e9,
  rate: 60,
  explore: true,
  failure: "",
  calls: [
    { entity: "CH-1", effector: "chiller__start", atS: 20, accepted: true, reason: "", effectEtaNs: 5e9 },
    {
      entity: "CH-1",
      effector: "chiller__stop",
      atS: 30,
      accepted: false,
      reason: "interlock_refused",
      effectEtaNs: null,
    },
  ],
};

function answering(result: unknown) {
  return bridgeTransport(() => respond(200, { result }));
}

function lastCall(t: ReturnType<typeof bridgeTransport>): { url: unknown; body: unknown } {
  const [url, init] = t.fetch.mock.calls.at(-1) ?? [];
  const raw = init?.body;
  return { url, body: typeof raw === "string" ? (JSON.parse(raw) as unknown) : undefined };
}

describe("job gateway", () => {
  it("starts a job with the plan in the director's words and reads its status", async () => {
    const t = answering(wire);
    const outcome = await bridgeAdapter(t).launch.startJob({
      domain: "factory",
      seed: 3,
      explore: true,
      durationS: 1800,
      rate: 60,
      faults: [{ entity: "CT-1", fault: "fan", atS: 10 }],
      calls: [{ entity: "CH-1", effector: "chiller__start", atS: 20, args: { chiller_id: "CH-1" } }],
    });
    expect(outcome).toEqual({ ok: true, value: status });
    expect(lastCall(t)).toEqual({
      url: "/api/actions/job.start",
      body: {
        args: {
          domain: "factory",
          seed: 3,
          explore: true,
          duration_s: 1800,
          rate: 60,
          faults: [{ entity_id: "CT-1", fault: "fan", at_s: 10 }],
          calls: [{ entity_id: "CH-1", effector: "chiller__start", at_s: 20, args: { chiller_id: "CH-1" } }],
        },
      },
    });
  });

  it("reads, stops and lists jobs by their world", async () => {
    const reading = answering(wire);
    expect(await bridgeAdapter(reading).launch.jobStatus("w-9")).toEqual({ ok: true, value: status });
    expect(lastCall(reading)).toEqual({ url: "/api/actions/job.status", body: { args: { world_id: "w-9" } } });
    const stopping = answering({ ...wire, state: "stopping" });
    expect(await bridgeAdapter(stopping).launch.stopJob("w-9")).toMatchObject({
      ok: true,
      value: { state: "stopping" },
    });
    expect(lastCall(stopping)).toEqual({ url: "/api/actions/job.stop", body: { args: { world_id: "w-9" } } });
    const listing = answering({ jobs: [wire, { ...wire, world_id: "w-8", state: "completed" }] });
    const listed = await bridgeAdapter(listing).launch.listJobs();
    expect(listed.ok && listed.value.map((job) => job.worldId)).toEqual(["w-9", "w-8"]);
    expect(lastCall(listing)).toEqual({ url: "/api/actions/job.list", body: { args: {} } });
  });

  it("says which job step the binary refused", async () => {
    const refused = bridgeTransport(() => respond(422, { error: { code: "job_active", message: "a job is running" } }));
    expect(
      await bridgeAdapter(refused).launch.startJob({ ...status, durationS: 60, faults: [], calls: [] }),
    ).toMatchObject({
      ok: false,
      failure: {
        kind: "rejected",
        step: "start",
        diagnostic: { displayCode: "job_active", message: "a job is running" },
      },
    });
    expect(await bridgeAdapter(refused).launch.jobStatus("w-1")).toMatchObject({ failure: { step: "status" } });
    expect(await bridgeAdapter(refused).launch.stopJob("w-1")).toMatchObject({ failure: { step: "stop" } });
    expect(await bridgeAdapter(refused).launch.listJobs()).toMatchObject({ failure: { step: "list" } });
  });

  it("refuses an answer it cannot read as a job", async () => {
    const outcome = await bridgeAdapter(answering({ ...wire, state: "paused" })).launch.jobStatus("w-9");
    expect(outcome).toMatchObject({ ok: false, failure: { kind: "unreadable" } });
  });
});

describe("job status payload", () => {
  it("reads a failed job's reason and a call without a reason or an eta", () => {
    const failed = jobStatus({ ...wire, state: "failed", failure: "job failed: boom", calls: [] });
    expect(failed).toMatchObject({ state: "failed", failure: "job failed: boom", calls: [] });
    const bare = jobStatus({ ...wire, calls: [{ at_s: 1, entity_id: "e", effector: "f", accepted: true }] });
    expect(bare?.calls).toEqual([
      { entity: "e", effector: "f", atS: 1, accepted: true, reason: "", effectEtaNs: null },
    ]);
  });

  it.each([
    ["not an object", null],
    ["an empty world id", { ...wire, world_id: "" }],
    ["no run id", { ...wire, run_id: undefined }],
    ["an empty domain", { ...wire, domain: "" }],
    ["an unknown state", { ...wire, state: "paused" }],
    ["a text seed", { ...wire, seed: "3" }],
    ["a missing start", { ...wire, start_ns: undefined }],
    ["an infinite reach", { ...wire, reached_ns: Infinity }],
    ["no requested end", { ...wire, requested_end_ns: "soon" }],
    ["a negative rate", { ...wire, rate: -1 }],
    ["a text rate", { ...wire, rate: "fast" }],
    ["a text explore", { ...wire, explore: "yes" }],
    ["calls missing", { ...wire, calls: undefined }],
    ["calls null", { ...wire, calls: null }],
    ["a call that is not an object", { ...wire, calls: [3] }],
    ["a call without an entity", { ...wire, calls: [{ at_s: 1, effector: "f", accepted: true }] }],
    ["a call without an effector", { ...wire, calls: [{ at_s: 1, entity_id: "e", accepted: true }] }],
    ["a call with a text time", { ...wire, calls: [{ at_s: "1", entity_id: "e", effector: "f", accepted: true }] }],
    ["a call with a text verdict", { ...wire, calls: [{ at_s: 1, entity_id: "e", effector: "f", accepted: "yes" }] }],
    [
      "a call with a numeric reason",
      { ...wire, calls: [{ at_s: 1, entity_id: "e", effector: "f", accepted: false, reason: 4 }] },
    ],
    [
      "a call with a text eta",
      { ...wire, calls: [{ at_s: 1, entity_id: "e", effector: "f", accepted: true, effect_eta_ns: "5" }] },
    ],
    ["a numeric failure", { ...wire, failure: 7 }],
  ])("refuses a job with %s", (_name, bad) => {
    expect(jobStatus(bad)).toBeNull();
  });

  it("refuses a listing that is not a list of jobs", () => {
    expect(jobList({ jobs: [] })).toEqual([]);
    expect(jobList({})).toBeNull();
    expect(jobList({ jobs: [wire, { ...wire, state: "paused" }] })).toBeNull();
    expect(jobList(null)).toBeNull();
  });
});
