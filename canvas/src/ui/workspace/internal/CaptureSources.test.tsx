import { fireEvent, screen, waitFor, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { announced, captureResult, refuseCapture } from "../../../test/captureFixture";
import { captureCalls, heading, openCapture, paste } from "../../../test/captureHarness";
import { deferred } from "../../../test/deferred";
import { FakeBridge } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { success } from "../../../test/outcome";
import { sampleWorkspace } from "../../../test/workspaceFixture";

const mebibyte = 1024 * 1024;

function chooseFile(file: File) {
  fireEvent.change(screen.getByLabelText("Or load it from a file"), { target: { files: [file] } });
}

function bigFile() {
  const big = new File(["{}"], "journal.jsonl");
  Object.defineProperty(big, "size", { value: mebibyte + 1 });
  return big;
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

  it("lets the same file be chosen again, because the choice is cleared once it is taken", async () => {
    await openCapture();
    const input = screen.getByLabelText<HTMLInputElement>("Or load it from a file");
    chooseFile(new File(["first"], "birth.json"));
    expect(input.value).toBe("");
    await waitFor(() => {
      expect(screen.getByLabelText("Announcement text")).toHaveValue("first");
    });
  });

  it("does nothing when no file was chosen", async () => {
    await openCapture();
    fireEvent.change(screen.getByLabelText("Or load it from a file"), { target: { files: [] } });
    expect(screen.getByLabelText("Announcement text")).toHaveValue("");
  });

  it("keeps the last choice when two files are chosen in a row and the first is read last", async () => {
    await openCapture();
    const slow = deferred<string>();
    const first = new File(["a"], "a.json");
    Object.defineProperty(first, "text", { value: () => slow.promise });
    chooseFile(first);
    chooseFile(new File(["second"], "b.json"));
    await waitFor(() => {
      expect(screen.getByLabelText("Announcement text")).toHaveValue("second");
    });
    slow.resolve("first");
    await Promise.resolve();
    expect(screen.getByLabelText("Announcement text")).toHaveValue("second");
  });

  it("does not let a file that is read late overwrite what was typed meanwhile", async () => {
    await openCapture();
    const slow = deferred<string>();
    const file = new File(["a"], "a.json");
    Object.defineProperty(file, "text", { value: () => slow.promise });
    chooseFile(file);
    fireEvent.change(screen.getByLabelText("Announcement text"), { target: { value: "typed" } });
    slow.resolve("late");
    await Promise.resolve();
    await Promise.resolve();
    expect(screen.getByLabelText("Announcement text")).toHaveValue("typed");
  });

  it("refuses a file over 1 MiB in plain words, without reading it, and keeps Read off", async () => {
    await openCapture();
    chooseFile(bigFile());
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
    chooseFile(bigFile());
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
    expect(screen.getByText("Read from the run w-1.")).toBeInTheDocument();
  });

  it("reads the newest run when none was picked", async () => {
    const { bridge } = await openCapture();
    fireEvent.click(screen.getByRole("button", { name: "Read this run" }));
    await screen.findByRole("heading", { level: 2, name: heading });
    expect(captureCalls(bridge)[0]).toEqual({ name: "kind.capture", args: { run: "w-2" } });
  });

  it("shows the binary's words behind a plain sentence when the run holds no announcement", async () => {
    const { bridge } = await openCapture();
    refuseCapture(bridge, "no_birth", "plantwire: no NBIRTH record found");
    fireEvent.click(screen.getByRole("button", { name: "Read this run" }));
    expect(await screen.findByRole("heading", { level: 2, name: "No announcement found" })).toBeInTheDocument();
    expect(screen.getByText(/plantwire: no NBIRTH record found/)).toBeInTheDocument();
  });

  it("says there is no run to read when the workspace has none", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ ...sampleWorkspace, runs: [] });
    await openCapture(undefined, bridge);
    expect(screen.getByText(/No recorded runs yet/)).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Read this run" })).toBeNull();
  });

  it("cannot read twice while a read is under way", async () => {
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
  it("goes back to the workspace on Escape, remembering to give the focus back to the button", async () => {
    const { nav, desk } = await openCapture();
    fireEvent.keyDown(screen.getByLabelText("Announcement text"), { key: "Escape" });
    expect(nav.location.hash).toBe("#/");
    expect(desk.capture().comingBack()).toBe(true);
  });

  it("goes back on the Workspace crumb too, and the review is kept for the way back", async () => {
    const { nav, desk } = await openCapture();
    paste();
    await screen.findByRole("heading", { level: 2, name: heading });
    fireEvent.click(screen.getByRole("button", { name: "Workspace" }));
    expect(nav.location.hash).toBe("#/");
    expect(desk.capture().snapshot().phase).toBe("reviewing");
  });

  it("ignores Escape while a read is under way", async () => {
    const bridge = new FakeBridge();
    bridge.handlers.set("kind.capture", () => new Promise<Canned<unknown>>(() => undefined));
    const { desk } = await openCapture(captureResult([announced()]), bridge);
    paste();
    await screen.findByText("Reading the announcement…");
    fireEvent.keyDown(document.body, { key: "Escape" });
    expect(desk.capture().comingBack()).toBe(false);
  });
});
