import { describe, expect, it } from "vitest";
import {
  announced,
  captureResult,
  differs,
  same,
  savedPump,
  savedPumpPath,
  unreadable,
} from "../../../test/captureFixture";
import { board, opened, text } from "../../../test/captureSessionHarness";
import { decisionCounts } from "./captureSelectors";

describe("replacing a saved device", () => {
  it("never overwrites silently: saving a device that differs from its file does nothing", async () => {
    const { bridge, session } = opened(captureResult([differs()]), true);
    await session.read(text);
    await session.save(0);
    expect(bridge.writes).toEqual([]);
  });

  it("sends the etag it read when the review was built, not the one the file has by now", async () => {
    const { bridge, session } = opened(captureResult([differs()]), true);
    await session.read(text);
    bridge.addFile(savedPumpPath, { ...savedPump, changed: true }, "k9");
    await session.replace(0);
    expect(bridge.writes).toEqual([
      { path: savedPumpPath, etag: "k3", content: expect.stringContaining('"kind": "pump"') as string },
    ]);
    expect(bridge.reads).toEqual([savedPumpPath]);
  });

  it("writes to the file the binary named, even when it is not the file of that name", async () => {
    const other = {
      target: "kinds/Pump.kind.json",
      file: "kinds/Pump.kind.json",
      collides_with: "kinds/Pump.kind.json",
    };
    const { bridge, session } = opened(captureResult([differs("/kind", other)]));
    bridge.addFile("kinds/Pump.kind.json", { kind: "Pump" }, "k5");
    await session.read(text);
    await session.replace(0);
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/Pump.kind.json", etag: "k5" });
  });

  it("really fails when the file changed after the comparison, writes nothing over it, and fails again on a retry", async () => {
    const { bridge, session } = opened(captureResult([differs()]), true);
    bridge.persistWrites();
    await session.read(text);
    bridge.addFile(savedPumpPath, { ...savedPump, changed: true }, "k4");
    await session.replace(0);
    expect(board(session.snapshot()).cards[0]?.decision).toMatchObject({
      kind: "failed",
      failure: {
        kind: "conflict",
        diagnostic: { message: expect.stringContaining("changed since it was compared") as string },
      },
    });
    await session.replace(0);
    expect(bridge.writes.map((write) => write.etag)).toEqual(["k3", "k3"]);
    expect(bridge.files.get(savedPumpPath)).toMatchObject({ ok: true, value: { etag: "k4" } });
  });

  it("replaces when the file is as it was compared", async () => {
    const { bridge, session } = opened(captureResult([differs()]), true);
    bridge.persistWrites();
    await session.read(text);
    await session.replace(0);
    expect(board(session.snapshot()).phase).toBe("saved");
    expect(bridge.files.get(savedPumpPath)).toMatchObject({ ok: true, value: { etag: "e1" } });
  });

  it("cannot replace a file it could not read, and writes nothing", async () => {
    const { bridge, session } = opened(captureResult([differs()]));
    await session.read(text);
    await session.replace(0);
    expect(bridge.writes).toEqual([]);
  });

  it("only replaces a device that differs", async () => {
    const { bridge, session } = opened();
    await session.read(text);
    await session.replace(0);
    expect(bridge.writes).toEqual([]);
  });
});

describe("a device the same as its file, one that cannot be compared, and a device skipped", () => {
  it("offers nothing to save when the board announced what is already saved, and nothing to skip", async () => {
    const { bridge, session } = opened(captureResult([same()]));
    await session.read(text);
    await session.save(0);
    await session.replace(0);
    session.skip(0);
    expect(bridge.writes).toEqual([]);
    expect(board(session.snapshot())).toMatchObject({ phase: "reviewing", cards: [{ decision: { kind: "open" } }] });
  });

  it("offers nothing to save when the saved file cannot be compared, but it can be skipped", async () => {
    const { bridge, session } = opened(captureResult([unreadable()]));
    await session.read(text);
    await session.save(0);
    await session.replace(0);
    session.skip(0);
    expect(bridge.writes).toEqual([]);
    expect(board(session.snapshot()).phase).toBe("saved");
  });

  it("skips a device, takes the skip back, and is done when every device is saved or skipped", async () => {
    const fan = { kind: "fan", metrics: [], commands: [] };
    const { session } = opened(captureResult([announced(), announced(fan)]));
    await session.read(text);
    session.skip(1);
    expect(board(session.snapshot()).phase).toBe("reviewing");
    session.unskip(1);
    expect(board(session.snapshot()).cards[1]?.decision).toEqual({ kind: "open" });
    session.skip(1);
    await session.save(0);
    const state = board(session.snapshot());
    expect(state.phase).toBe("saved");
    expect(decisionCounts(state.cards)).toEqual({ saved: 1, skipped: 1 });
    session.skip(0);
    session.unskip(0);
    expect(decisionCounts(board(session.snapshot()).cards)).toEqual({ saved: 1, skipped: 1 });
  });

  it("is not done when the board announced nothing, or only what is saved", async () => {
    const empty = opened(captureResult([]));
    await empty.session.read(text);
    expect(board(empty.session.snapshot())).toMatchObject({ phase: "reviewing", cards: [] });
    const saved = opened(captureResult([same()]));
    await saved.session.read(text);
    expect(board(saved.session.snapshot()).phase).toBe("reviewing");
  });
});
