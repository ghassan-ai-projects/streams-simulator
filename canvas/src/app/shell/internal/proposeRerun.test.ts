import { describe, expect, it } from "vitest";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { check } from "../../../test/sessionHarness";
import { launchHarness } from "../../../test/launchFixture";
import { memoryLocation } from "../../../test/location";
import { report } from "../../../test/runFixture";
import { reviewState } from "../../../test/reviewFixture";
import { NavigationStore } from "./navigationStore";
import { proposeRerun } from "./proposeRerun";

const plan = { seed: 1, durationS: 60, explore: true, faults: [], perturbations: [], calls: [] };

function setup() {
  const location = memoryLocation();
  const { launch } = launchHarness(new FakeBridge());
  return { location, launch, navigation: new NavigationStore(location) };
}

describe("proposing another run", () => {
  it("proposes the run's plan for its plant, naming the run, and opens the plant's launch panel", () => {
    const { location, launch, navigation } = setup();
    proposeRerun(launch, navigation, reviewState({}), "w-4", "domains/mini.domain.json");
    expect(launch.proposal("mini")).toMatchObject({ from: "w-4", plan: { seed: 7 } });
    expect(location.hash).toContain("launch=1");
    expect(navigation.snapshot()).toMatchObject({ screen: "plant", domain: "domains/mini.domain.json", launch: true });
  });

  it("proposes fresh settings, marked as such, for a run that did not finish", () => {
    const { launch, navigation } = setup();
    const open = { ...report, ended: false, complete: false, mode: "explore" };
    proposeRerun(
      launch,
      navigation,
      reviewState({ report: { phase: "ready", value: open } }),
      "w-5",
      "domains/mini.domain.json",
    );
    expect(launch.proposal("mini")).toMatchObject({ from: "w-5", fresh: true, plan: { explore: true, faults: [] } });
  });

  it("proposes nothing and stays where it is before the report is ready", () => {
    const { location, launch, navigation } = setup();
    const before = location.hash;
    proposeRerun(launch, navigation, reviewState({ report: { phase: "loading" } }), "w-4", "domains/mini.domain.json");
    expect(launch.proposal("mini")).toBeNull();
    expect(location.hash).toBe(before);
  });

  it("sets up another connected session for a connected run: no sockets, the recorded list kept, the mode kept", () => {
    const { location, launch, navigation } = setup();
    const digest = `sha256:${"a".repeat(64)}`;
    const state = reviewState({ connection: "saved", sessionDomain: "mini", sessionPolicies: [digest] });
    proposeRerun(launch, navigation, state, "c-4", "domains/mini.domain.json");
    expect(launch.proposal("mini")).toBeNull();
    expect(launch.connect.draftFor("mini")).toMatchObject({
      hostSocket: "",
      policies: digest,
      policiesFrom: "previous",
    });
    expect(launch.connect.kindFor("mini")).toBe("session");
    expect(location.hash).toContain("launch=1");
  });

  it("keeps no list for a connected run that recorded none", () => {
    const { launch, navigation } = setup();
    proposeRerun(launch, navigation, reviewState({ connection: "saved" }), "c-4", "domains/mini.domain.json");
    expect(launch.connect.draftFor("mini")).toMatchObject({ policies: "", policiesFrom: "operator" });
  });

  it("sets up fresh session settings with the unrecovered flag for a connected run that did not finish", () => {
    const { location, launch, navigation } = setup();
    const digest = `sha256:${"a".repeat(64)}`;
    const open = { ...report, ended: false, complete: false };
    const state = reviewState({
      connection: "saved",
      sessionPolicies: [digest],
      report: { phase: "ready", value: open },
    });
    proposeRerun(launch, navigation, state, "c-4", "domains/mini.domain.json");
    expect(launch.proposal("mini")).toBeNull();
    expect(launch.connect.draftFor("mini")).toMatchObject({ hostSocket: "", policies: "", seed: "1" });
    expect(launch.connect.kindFor("mini")).toBe("session");
    expect(launch.connect.unrecovered("mini")).toBe(true);
    expect(location.hash).toContain("launch=1");
  });

  it("clears an earlier session failure and an earlier launch failure so neither comes back", async () => {
    const fake = new FakeBridge();
    fake.actions.set("session.check", success(check));
    fake.actions.set("session.start", failure(cannedFailure("plantwire_failed", 500, "host absent")));
    fake.actions.set("job.start", failure(cannedFailure("job_active", 409, "busy")));
    const { launch } = launchHarness(fake);
    await launch.connect.start("mini", { ...launch.connect.draftFor("mini"), hostSocket: "/x" });
    await launch.start({ domain: "mini", title: "Mini", path: "p" }, plan);
    expect(launch.connect.snapshot().phase).toBe("failed");
    expect(launch.snapshot().phase).toBe("failed");
    proposeRerun(launch, new NavigationStore(memoryLocation()), reviewState({}), "w-4", "domains/mini.domain.json");
    expect(launch.connect.snapshot()).toEqual({ phase: "idle" });
    expect(launch.snapshot()).toEqual({ phase: "idle" });
  });
});
