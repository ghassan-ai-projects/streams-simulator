import { describe, expect, it, vi } from "vitest";
import { deferred } from "../../../test/deferred";
import { cannedFailure, FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { failure, success } from "../../../test/outcome";
import { assumptionsWire, diffWire, lineageWire } from "../../../test/provenanceFixture";
import { fetchedFor } from "./fetchStore";
import { createProvenanceSession } from "./provenanceSession";
import { createWorkspaceStore } from "./workspaceStore";

function session(bridge = new FakeBridge()) {
  const workspace = createWorkspaceStore(bridge.workspace);
  return { bridge, workspace, provenance: createProvenanceSession(bridge.provenance, workspace) };
}

const lineageCalls = (bridge: FakeBridge) => bridge.calls.filter((call) => call.name === "domain.lineage");

describe("the lineage of the workspace", () => {
  it("is read once each time the workspace is read, not before", async () => {
    const { bridge, workspace, provenance } = session();
    bridge.actions.set("domain.lineage", success(lineageWire));
    expect(provenance.lineage.snapshot()).toBeNull();
    await workspace.open();
    await vi.waitFor(() => {
      expect(provenance.lineage.snapshot()?.plants).toHaveLength(5);
    });
    bridge.changeOnDisk();
    await vi.waitFor(() => {
      expect(lineageCalls(bridge)).toHaveLength(2);
    });
  });

  it("does not read it while the workspace is still loading or when it failed", async () => {
    const { bridge, workspace } = session();
    bridge.workspaceOutcome = failure(cannedFailure("bridge_failed", 500, "disk full"));
    await workspace.open();
    expect(lineageCalls(bridge)).toHaveLength(0);
  });

  it("goes back to nothing when the bridge cannot give it", async () => {
    const { bridge, provenance } = session();
    bridge.actions.set("domain.lineage", success(lineageWire));
    await provenance.lineage.refresh();
    expect(provenance.lineage.snapshot()).not.toBeNull();
    bridge.actions.set("domain.lineage", failure(cannedFailure("domain_failed", 500, "broken")));
    await provenance.lineage.refresh();
    expect(provenance.lineage.snapshot()).toBeNull();
  });

  it("asks once more, and only once, when changes arrive while it is being read", async () => {
    const { bridge, provenance } = session();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("domain.lineage", () => gate.promise);
    const first = provenance.lineage.refresh();
    void provenance.lineage.refresh();
    void provenance.lineage.refresh();
    bridge.handlers.set("domain.lineage", () => success(lineageWire));
    gate.resolve(success(lineageWire));
    await first;
    expect(lineageCalls(bridge)).toHaveLength(2);
  });
});

describe("the assumptions and the diff", () => {
  it("loads the assumptions of a domain under a key, and forgets them when cleared", async () => {
    const { bridge, provenance } = session();
    bridge.actions.set("domain.assumptions", success(assumptionsWire));
    const load = provenance.assumptions.load("domains/a.domain.json", "twins/a.twin.json");
    expect(provenance.assumptions.snapshot()).toEqual({
      phase: "loading",
      key: "domains/a.domain.json|twins/a.twin.json",
    });
    await load;
    expect(provenance.assumptions.snapshot()).toMatchObject({
      phase: "ready",
      value: { domain: { id: "arduino-thermal-twin" } },
    });
    provenance.assumptions.clear();
    expect(provenance.assumptions.snapshot()).toEqual({ phase: "idle" });
  });

  it("keeps the answer of the last question when an earlier one comes back later", async () => {
    const { bridge, provenance } = session();
    const slow = deferred<Canned<unknown>>();
    bridge.handlers.set("domain.diff", (args) => (args.a === "w-1" ? slow.promise : success(diffWire)));
    const first = provenance.diff.load("w-1", "domains/a.domain.json");
    await provenance.diff.load("w-2", "domains/a.domain.json");
    slow.resolve(failure(cannedFailure("late", 500, "too late")));
    await first;
    expect(provenance.diff.snapshot()).toMatchObject({ phase: "ready", key: "w-2|domains/a.domain.json" });
  });

  it("keeps the binary's refusal", async () => {
    const { bridge, provenance } = session();
    bridge.actions.set("domain.diff", failure(cannedFailure("run_not_found", 422, "no run w-9")));
    await provenance.diff.load("w-9", "domains/a.domain.json");
    expect(provenance.diff.snapshot()).toMatchObject({ phase: "failed", failure: { kind: "refused" } });
  });

  it("shows a state only to the question it answers", () => {
    const state = { phase: "loading", key: "a|b" } as const;
    expect(fetchedFor(state, "a|b")).toBe(state);
    expect(fetchedFor(state, "a|c")).toEqual({ phase: "idle" });
    expect(fetchedFor({ phase: "idle" }, "a|b")).toEqual({ phase: "idle" });
  });
});

describe("composing an out-of-date plant again", () => {
  it("composes, reads the lineage again and returns the summary", async () => {
    const { bridge, provenance } = session();
    bridge.actions.set("domain.lineage", success(lineageWire));
    bridge.actions.set("domain.compose", success({ channels: 1, faults: 2, effectors: 3, digest: "d" }));
    const composed = await provenance.recompose("compositions/a.composition.json", "domains/a.domain.json");
    expect(composed).toEqual({ ok: true, value: { channels: 1, faults: 2, effectors: 3, digest: "d" } });
    await vi.waitFor(() => {
      expect(lineageCalls(bridge)).toHaveLength(1);
    });
  });

  it("leaves the lineage alone when it could not compose", async () => {
    const { bridge, provenance } = session();
    bridge.actions.set("domain.compose", failure(cannedFailure("compose_failed", 422, "bad preset")));
    const composed = await provenance.recompose("c", "d");
    expect(composed.ok).toBe(false);
    expect(lineageCalls(bridge)).toHaveLength(0);
  });
});
