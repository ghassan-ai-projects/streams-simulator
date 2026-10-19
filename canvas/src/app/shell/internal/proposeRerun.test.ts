import { describe, expect, it } from "vitest";
import { FakeBridge } from "../../../test/fakeBridge";
import { launchHarness } from "../../../test/launchFixture";
import { memoryLocation } from "../../../test/location";
import { reviewState } from "../../../test/reviewFixture";
import { NavigationStore } from "./navigationStore";
import { proposeRerun } from "./proposeRerun";

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
});
