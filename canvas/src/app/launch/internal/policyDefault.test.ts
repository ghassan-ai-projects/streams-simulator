import { describe, expect, it } from "vitest";
import { firstSessionDraft } from "../../../model/launch";
import { success } from "../../ports";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { failure } from "../../../test/outcome";
import { ConnectStore } from "./connectStore";

const a = `sha256:${"a".repeat(64)}`;
const b = `sha256:${"b".repeat(64)}`;
const file = "settings/default.policies.json";
const check = {
  domain: "mini",
  environment: "simulated",
  instances: [{ instance: "fan-01", origin: "simulated", kind: "fan" }],
  policies: [a, b],
};

function rig(saved?: readonly string[]) {
  const bridge = new FakeBridge();
  bridge.persistWrites();
  bridge.actions.set("session.check", success(check));
  if (saved !== undefined) bridge.addFile(file, { policies: saved }, "e0");
  return { bridge, connect: new ConnectStore(bridge.launch, fakeClock()) };
}

describe("the workspace default of the allowed policy digests", () => {
  it("pre-fills a plant's first draft with the saved list, marked as from the default", async () => {
    const { connect } = rig([b, a]);
    const filled = await connect.prefill("mini");
    expect(filled).toMatchObject({ policies: `${b}\n${a}`, policiesFrom: "default" });
    expect(connect.draftFor("mini")).toBe(filled);
    expect(connect.defaults.snapshot()).toMatchObject({ file: "present", policies: [b, a], save: "idle" });
  });

  it("pre-fills nothing when there is no default, an empty one, or one that cannot be read", async () => {
    expect(await rig().connect.prefill("mini")).toBeNull();
    expect(rig().connect.defaults.snapshot().file).toBe("unread");
    const none = rig();
    await none.connect.prefill("mini");
    expect(none.connect.defaults.snapshot()).toMatchObject({ file: "absent", policies: [] });
    expect(none.connect.draftFor("mini")).toBe(firstSessionDraft);
    expect(await rig([]).connect.prefill("mini")).toBeNull();
    const broken = rig();
    broken.bridge.addFile(file, "{ not json");
    expect(await broken.connect.prefill("mini")).toBeNull();
    expect(broken.connect.defaults.snapshot()).toMatchObject({ file: "unreadable", policies: [] });
    const down = rig([a]);
    down.bridge.files.set(file, failure(cannedFailure("bridge_failed", 500, "boom")));
    expect(await down.connect.prefill("mini")).toBeNull();
    expect(down.connect.defaults.snapshot().file).toBe("unreadable");
  });

  it("never overwrites a draft the plant already holds, an edited one or one the operator added a digest to", async () => {
    const { connect } = rig([a]);
    const typed = { ...firstSessionDraft, hostSocket: "/tmp/h.sock" };
    connect.keepDraft("mini", typed);
    expect(await connect.prefill("mini")).toBeNull();
    expect(connect.draftFor("mini")).toBe(typed);
    const added = connect.addPolicy("other", b);
    expect(await connect.prefill("other")).toBeNull();
    expect(connect.draftFor("other")).toBe(added);
    expect(await connect.prefill("third")).toMatchObject({ policies: a });
  });

  it("does not fill a draft the operator edits while the default is being read", async () => {
    const { bridge, connect } = rig([a]);
    const gate = bridge.holdRead(file);
    const pending = connect.prefill("mini");
    const typed = { ...firstSessionDraft, policies: b, policiesFrom: "operator" as const };
    connect.keepDraft("mini", typed);
    gate[0]?.resolve(success({ path: file, content: JSON.stringify({ policies: [a] }), etag: "e0" }));
    expect(await pending).toBeNull();
    expect(connect.draftFor("mini")).toBe(typed);
  });

  it("adds a digest to the plant's draft once, as the operator's own edit", () => {
    const { connect } = rig();
    const first = connect.addPolicy("mini", a);
    expect(first).toMatchObject({ policies: a, policiesFrom: "operator" });
    expect(connect.addPolicy("mini", a)).toBe(first);
    expect(connect.addPolicy("mini", b).policies).toBe(`${a}\n${b}`);
    expect(connect.draftFor("mini").policies).toBe(`${a}\n${b}`);
  });
});

describe("saving the default", () => {
  async function checked(saved?: readonly string[]) {
    const made = rig(saved);
    await made.connect.prefill("mini");
    await made.connect.check("mini", { ...firstSessionDraft, hostSocket: "/tmp/h.sock" });
    return made;
  }

  it("saves the list the binary echoed after a passing check, not what was typed", async () => {
    const { bridge, connect } = await checked([a]);
    await connect.saveDefault("mini");
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.writes[0]).toMatchObject({ path: file, etag: "e0" });
    expect(JSON.parse(bridge.writes[0]?.content ?? "")).toEqual({ policies: [a, b] });
    expect(connect.defaults.snapshot()).toMatchObject({ file: "present", policies: [a, b], save: "saved" });
  });

  it("creates the default from the empty etag when none exists, and can save again from the new etag", async () => {
    const { bridge, connect } = await checked();
    await connect.saveDefault("mini");
    expect(bridge.writes[0]?.etag).toBe("");
    await connect.saveDefault("mini");
    expect(bridge.writes[1]?.etag).toBe("e1");
    expect(connect.defaults.snapshot().save).toBe("saved");
  });

  it("saves nothing before a check, for another plant, or after the draft changed", async () => {
    const made = rig([a]);
    await made.connect.saveDefault("mini");
    const { bridge, connect } = await checked([a]);
    await connect.saveDefault("other");
    connect.keepDraft("mini", { ...firstSessionDraft, policies: b });
    await connect.saveDefault("mini");
    expect(made.bridge.writes).toEqual([]);
    expect(bridge.writes).toEqual([]);
  });

  it("reports a default that changed elsewhere as a conflict, saves nothing, and keeps the new etag for a retry", async () => {
    const { bridge, connect } = await checked([a]);
    bridge.addFile(file, { policies: [b] }, "elsewhere");
    await connect.saveDefault("mini");
    expect(connect.defaults.snapshot()).toMatchObject({ save: "conflict", file: "present", policies: [b] });
    expect(bridge.writes).toHaveLength(1);
    await connect.saveDefault("mini");
    expect(bridge.writes[1]?.etag).toBe("elsewhere");
    expect(connect.defaults.snapshot().save).toBe("saved");
  });

  it("reports a failed write with the bridge's words and keeps the earlier default", async () => {
    const { bridge, connect } = await checked([a]);
    bridge.writeOutcomes.set(file, failure(cannedFailure("bridge_failed", 500, "disk full")));
    await connect.saveDefault("mini");
    expect(connect.defaults.snapshot()).toMatchObject({ save: "failed", problem: "disk full", policies: [a] });
  });

  it("does not guess an etag when the default could not be read, and reads it first when the earlier read failed", async () => {
    const made = rig([a]);
    made.bridge.files.set(file, failure(cannedFailure("bridge_failed", 500, "boom")));
    await made.connect.prefill("mini");
    await made.connect.check("mini", firstSessionDraft);
    await made.connect.saveDefault("mini");
    expect(made.connect.defaults.snapshot()).toMatchObject({ save: "failed" });
    expect(made.connect.defaults.snapshot().problem).toContain("could not be read");
    expect(made.bridge.writes).toEqual([]);
    made.bridge.addFile(file, { policies: [a] }, "e5");
    await made.connect.saveDefault("mini");
    expect(made.bridge.writes[0]?.etag).toBe("e5");
  });

  it("ignores a second save while one is in flight", async () => {
    const { bridge, connect } = await checked([a]);
    const gate = bridge.holdWrite(1);
    const first = connect.saveDefault("mini");
    await connect.saveDefault("mini");
    expect(connect.defaults.snapshot().save).toBe("saving");
    gate[0]?.resolve(success({ path: file, etag: "w1" }));
    await first;
    expect(bridge.writes).toHaveLength(1);
  });

  it("forgets the result of a save when the draft changes or a new check starts, so it never outlives its list", async () => {
    const { connect } = await checked([a]);
    await connect.saveDefault("mini");
    expect(connect.defaults.snapshot().save).toBe("saved");
    connect.keepDraft("mini", { ...firstSessionDraft, policies: b });
    expect(connect.defaults.snapshot()).toMatchObject({ save: "idle", problem: "" });
    await connect.check("mini", firstSessionDraft);
    await connect.saveDefault("mini");
    expect(connect.defaults.snapshot().save).toBe("saved");
    await connect.check("mini", firstSessionDraft);
    expect(connect.defaults.snapshot().save).toBe("idle");
  });

  it("keeps the saving state while a save is in flight when the draft changes", async () => {
    const { bridge, connect } = await checked([a]);
    const gate = bridge.holdWrite(1);
    const saving = connect.saveDefault("mini");
    connect.keepDraft("mini", { ...firstSessionDraft, policies: b });
    expect(connect.defaults.snapshot().save).toBe("saving");
    gate[0]?.resolve(success({ path: file, etag: "w1" }));
    await saving;
  });
});

describe("the mode and the draft kept for the next run", () => {
  it("remembers the run kind chosen for each plant, to completion until one is chosen", () => {
    const { connect } = rig();
    expect(connect.kindFor("mini")).toBe("complete");
    connect.chooseKind("mini", "session");
    expect(connect.kindFor("mini")).toBe("session");
    expect(connect.kindFor("other")).toBe("complete");
  });

  it("sets up the next session with the recorded list and no sockets, in the session mode", () => {
    const { connect } = rig();
    connect.setUpNext("mini", [a, b]);
    expect(connect.draftFor("mini")).toMatchObject({
      hostSocket: "",
      policies: `${a}\n${b}`,
      policiesFrom: "previous",
    });
    expect(connect.kindFor("mini")).toBe("session");
  });
});
