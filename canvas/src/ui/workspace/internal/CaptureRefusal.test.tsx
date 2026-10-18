import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { captureResult, refuseCapture, serveCapture } from "../../../test/captureFixture";
import { captureCalls, heading, openCapture, paste } from "../../../test/captureHarness";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { failure as failed } from "../../../test/outcome";

async function refused(keyword = "kind_digest_mismatch") {
  const bridge = new FakeBridge();
  const opened = await openCapture(undefined, bridge);
  refuseCapture(bridge, keyword, "plantwire: the birth is refused");
  paste();
  await screen.findByRole("alert");
  return opened;
}

async function failing(code: string, status: number, message: string) {
  const bridge = new FakeBridge();
  const opened = await openCapture(undefined, bridge);
  bridge.actions.set("kind.capture", failed(cannedFailure(code, status, message)));
  paste();
  await screen.findByRole("alert");
  return opened;
}

describe("a refused announcement", () => {
  it("leads with a plain sentence and the next step, keeps the binary's words in the details, and writes nothing", async () => {
    const { bridge, view } = await refused();
    expect(
      screen.getByRole("heading", { level: 2, name: "The board's announcement was not accepted" }),
    ).toBeInTheDocument();
    const alert = screen.getByRole("alert");
    expect(alert).toHaveTextContent("does not match its own fingerprint");
    expect(alert).toHaveTextContent("Copy the announcement again from the board.");
    expect(alert).not.toHaveTextContent("plantwire");
    expect(alert).toHaveFocus();
    expect(screen.getByText(/plantwire: the birth is refused/)).toBeInTheDocument();
    expect(screen.getByText("does not hold")).toBeInTheDocument();
    expect(screen.getByText("Nothing was saved.")).toBeInTheDocument();
    expect(screen.queryByRole("heading", { name: heading })).toBeNull();
    expect(bridge.writes).toEqual([]);
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("says no announcement was found, for text and for a run alike", async () => {
    await refused("no_birth");
    expect(screen.getByRole("heading", { level: 2, name: "No announcement found" })).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("There is no announcement from a board in what was read.");
  });

  it("says garbage is not valid JSON when the binary says so", async () => {
    await refused("json");
    expect(screen.getByRole("heading", { level: 2, name: "That is not valid JSON" })).toBeInTheDocument();
  });

  it("tells the sealed part of a run from an announcement, without the raw text as a headline", async () => {
    await refused("sealed_stream");
    expect(screen.getByRole("alert")).toHaveTextContent("sealed part of a run");
  });

  it("lists what else was found beside the first cause", async () => {
    const bridge = new FakeBridge();
    await openCapture(undefined, bridge);
    const issues = [
      { message: "none", keyword: "no_birth" },
      { message: "not json", keyword: "json" },
    ];
    bridge.actions.set("kind.capture", failed({ ...cannedFailure("kind_failed", 422, "refused"), issues }));
    paste();
    const also = await screen.findByRole("list", { name: "Also found" });
    expect(within(also).getByText("The file is not valid JSON")).toBeInTheDocument();
    expect(screen.getByRole("heading", { level: 2, name: "No announcement found" })).toBeInTheDocument();
  });

  it("can be tried again, and reviews the announcement once it holds", async () => {
    const { bridge } = await refused("no_birth");
    serveCapture(bridge, captureResult());
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(await screen.findByRole("heading", { level: 2, name: heading })).toHaveFocus();
    expect(captureCalls(bridge)).toHaveLength(2);
  });
});

describe("an announcement that could not be asked for", () => {
  it("says the bridge is unavailable, with the bridge's words in the details", async () => {
    await failing("bridge_failed", 500, "down");
    expect(screen.getByRole("heading", { level: 2, name: "The bridge is unavailable" })).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveFocus();
    expect(screen.getByText("down (bridge_failed)")).toBeInTheDocument();
  });

  it("says a run without a recording holds no announcement", async () => {
    await failing("file_not_found", 404, "no journal");
    expect(screen.getByRole("heading", { level: 2, name: "No announcement found" })).toBeInTheDocument();
    expect(screen.getByRole("alert")).toHaveTextContent("This run has no recorded announcement.");
  });

  it("says what the page cannot understand", async () => {
    await failing("too_large", 413, "too large");
    expect(screen.getByRole("heading", { level: 2, name: "The announcement could not be read" })).toBeInTheDocument();
  });
});
