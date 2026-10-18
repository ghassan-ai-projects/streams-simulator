import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { announced, captureResult, refuseCapture } from "../../../test/captureFixture";
import { captureCalls, heading, openCapture, paste } from "../../../test/captureHarness";
import { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { success } from "../../../test/outcome";
import { sampleWorkspace } from "../../../test/workspaceFixture";

const mebibyte = 1024 * 1024;

function chooseFile(file: File) {
  fireEvent.change(screen.getByLabelText("Or load it from a file"), { target: { files: [file] } });
}

describe("where the announcement comes from", () => {
  it("loads a file into the text, so it can be read before it is sent", async () => {
    const { bridge } = await openCapture();
    chooseFile(new File(['{"type":"NBIRTH"}'], "birth.json", { type: "application/json" }));
    await waitFor(() => {
      expect(screen.getByLabelText("Announcement text")).toHaveValue('{"type":"NBIRTH"}');
    });
    expect(captureCalls(bridge)).toEqual([]);
    fireEvent.click(screen.getByRole("button", { name: "Read the announcement" }));
    expect(await screen.findByRole("heading", { level: 2, name: heading })).toBeInTheDocument();
    expect(captureCalls(bridge)[0]).toEqual({ name: "kind.capture", args: { content: '{"type":"NBIRTH"}' } });
  });

  it("does nothing when no file was chosen", async () => {
    await openCapture();
    fireEvent.change(screen.getByLabelText("Or load it from a file"), { target: { files: [] } });
    expect(screen.getByLabelText("Announcement text")).toHaveValue("");
  });

  it("refuses a file over 1 MiB in plain words, without reading it, and keeps Read off", async () => {
    await openCapture();
    const big = new File(["{}"], "journal.jsonl");
    Object.defineProperty(big, "size", { value: mebibyte + 1 });
    chooseFile(big);
    expect(await screen.findByRole("alert")).toHaveTextContent("That is larger than 1 MiB.");
    expect(screen.getByLabelText("Announcement text")).toHaveValue("");
    expect(screen.getByRole("button", { name: "Read the announcement" })).toHaveAttribute("aria-disabled", "true");
  });

  it("refuses pasted text over 1 MiB, and takes the refusal back when the text is cut", async () => {
    const { bridge } = await openCapture();
    const text = screen.getByLabelText("Announcement text");
    fireEvent.change(text, { target: { value: "x".repeat(mebibyte + 1) } });
    expect(screen.getByRole("alert")).toHaveTextContent("larger than 1 MiB");
    expect(text).toHaveAttribute("aria-invalid", "true");
    fireEvent.click(screen.getByRole("button", { name: "Read the announcement" }));
    expect(captureCalls(bridge)).toEqual([]);
    fireEvent.change(text, { target: { value: "{}" } });
    expect(screen.queryByRole("alert")).toBeNull();
    expect(text).not.toHaveAttribute("aria-invalid");
  });

  it("says a file that cannot be read could not be read", async () => {
    await openCapture();
    const broken = new File(["{}"], "birth.json");
    Object.defineProperty(broken, "text", { value: () => Promise.reject(new Error("unreadable")) });
    chooseFile(broken);
    expect(await screen.findByRole("alert")).toHaveTextContent("That file could not be read.");
  });

  it("clears the file's refusal when text is typed instead", async () => {
    await openCapture();
    const big = new File(["{}"], "journal.jsonl");
    Object.defineProperty(big, "size", { value: mebibyte + 1 });
    chooseFile(big);
    await screen.findByRole("alert");
    fireEvent.change(screen.getByLabelText("Announcement text"), { target: { value: "{}" } });
    expect(screen.queryByRole("alert")).toBeNull();
  });
});

describe("reading a recorded run", () => {
  it("offers the runs of the workspace and reads the one that is picked", async () => {
    const { bridge } = await openCapture();
    const run = screen.getByLabelText("Recorded run");
    expect(within(run).getAllByRole("option")).toHaveLength(2);
    fireEvent.change(run, { target: { value: "w-1" } });
    fireEvent.click(screen.getByRole("button", { name: "Read this run" }));
    expect(await screen.findByRole("heading", { level: 2, name: heading })).toHaveFocus();
    expect(captureCalls(bridge)).toEqual([{ name: "kind.capture", args: { run: "w-1" } }]);
  });

  it("reads the newest run when none was picked", async () => {
    const { bridge } = await openCapture();
    fireEvent.click(screen.getByRole("button", { name: "Read this run" }));
    await screen.findByRole("heading", { level: 2, name: heading });
    expect(captureCalls(bridge)[0]).toEqual({ name: "kind.capture", args: { run: "w-2" } });
  });

  it("shows the binary's sentence when the run holds no announcement", async () => {
    const { bridge } = await openCapture();
    refuseCapture(bridge, "no_birth", "plantwire: no NBIRTH record found");
    fireEvent.click(screen.getByRole("button", { name: "Read this run" }));
    expect(await screen.findByRole("alert")).toHaveTextContent("plantwire: no NBIRTH record found");
    expect(screen.getByText("No announcement from a board was found")).toBeInTheDocument();
  });

  it("says there is no run to read when the workspace has none", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs: [] });
    await openCapture(undefined, bridge);
    expect(screen.getByText(/No recorded runs yet/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Read this run" })).toBeNull();
  });

  it("cannot read while a read is under way", async () => {
    const bridge = new FakeBridge();
    bridge.handlers.set("kind.capture", () => new Promise<Canned<unknown>>(() => undefined));
    await openCapture(captureResult([announced()]), bridge);
    paste();
    expect(await screen.findByRole("status")).toHaveTextContent("Reading the announcement…");
    fireEvent.click(screen.getByRole("button", { name: "Read this run" }));
    fireEvent.click(screen.getByRole("button", { name: "Read the announcement" }));
    expect(captureCalls(bridge)).toHaveLength(1);
  });
});

describe("leaving the screen", () => {
  it("returns to the Devices list on Escape, with the focus on the button that opened it", async () => {
    await openCapture();
    fireEvent.keyDown(screen.getByLabelText("Announcement text"), { key: "Escape" });
    const devices = within(screen.getByRole("region", { name: "Devices" }));
    expect(devices.getByRole("button", { name: "Capture from a board" })).toHaveFocus();
    expect(screen.queryByRole("heading", { level: 1, name: "Capture from a board" })).toBeNull();
  });

  it("returns on the Workspace crumb, and opens again with a clean screen", async () => {
    await openCapture();
    paste();
    await screen.findByRole("heading", { level: 2, name: heading });
    fireEvent.click(screen.getByRole("button", { name: "Workspace" }));
    const devices = within(screen.getByRole("region", { name: "Devices" }));
    fireEvent.click(devices.getByRole("button", { name: "Capture from a board" }));
    expect(screen.getByLabelText("Announcement text")).toHaveValue("");
    expect(screen.queryByRole("heading", { level: 2, name: heading })).toBeNull();
  });

  it("is also on the page of an empty workspace", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ ...sampleWorkspace, domains: [], compositions: [], runs: [] });
    await openCapture(undefined, bridge);
    expect(screen.getByRole("heading", { level: 1, name: "Capture from a board" })).toBeInTheDocument();
  });
});
