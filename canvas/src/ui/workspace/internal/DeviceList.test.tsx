import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { createDeviceDesk } from "../../../app/devices";
import { serveCapture } from "../../../test/captureFixture";
import { deferred } from "../../../test/deferred";
import { failure as failed, success } from "../../../test/outcome";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import type { Canned } from "../../../test/fakeFailures";
import { kindCheck } from "../../../test/kindCheckFake";
import { sampleWorkspace } from "../../../test/workspaceFixture";
import { navigation, show } from "../../../test/workspaceHarness";

const fan = { path: "kinds/fan.kind.json", kind: "fan", title: "" };
const pump = { path: "kinds/pump.kind.json", kind: "", title: "" };

function withDevices() {
  const bridge = new FakeBridge();
  bridge.workspaceOutcome = success({ ...sampleWorkspace, kinds: [fan, pump] });
  bridge.addFile(fan.path, {
    kind: "fan",
    metrics: [{ name: "duty" }, { name: "mode" }],
    commands: [{ name: "stop" }],
  });
  bridge.addFile(pump.path, "not json");
  bridge.handlers.set("kind.check", kindCheck);
  return bridge;
}

function devices() {
  return within(screen.getByRole("region", { name: "Devices" }));
}

describe("the devices section", () => {
  it("says what a device is and offers to create one when there are none", async () => {
    const view = await show(new FakeBridge());
    expect(devices().getByText(/No devices yet/)).toBeInTheDocument();
    expect(devices().getByRole("button", { name: "Create device" })).toHaveAttribute("aria-expanded", "false");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("lists each device with what it reports and accepts, and opens it in the device editor", async () => {
    const nav = navigation();
    await show(withDevices(), nav);
    expect(await devices().findByText("2 readings · 1 command")).toBeInTheDocument();
    expect(devices().getByText("pump")).toBeInTheDocument();
    expect(devices().getByText("Could not be read")).toBeInTheDocument();
    expect(devices().getAllByRole("listitem")).toHaveLength(2);
    fireEvent.click(devices().getByRole("button", { name: "Edit fan" }));
    expect(nav.location.hash).toBe("#/device?device=kinds%2Ffan.kind.json");
  });

  it("says whether each device is valid and has a twin, from what the workspace already lists", async () => {
    const bridge = withDevices();
    const listed = bridge.workspaceOutcome;
    if (!listed.ok) throw new Error("no workspace");
    const kinds = [fan, pump, { path: "kinds/heater.kind.json", kind: "heater", title: "" }];
    const made = { path: "domains/heater-twin.domain.json", id: "heater-twin", title: "Heater", instances: 1 };
    bridge.workspaceOutcome = success({ ...listed.value, kinds, domains: [...listed.value.domains, made] });
    bridge.addFile("kinds/heater.kind.json", { kind: "heater", metrics: [], commands: [] });
    bridge.addFile("twins/heater.twin.json", { id: "heater-twin", kinds: [] });
    await show(bridge);
    const row = (name: string) => within(devices().getByText(name).closest("li") as HTMLElement);
    expect(await row("heater").findByText("Twin ready")).toBeInTheDocument();
    expect(row("fan").getByText("Needs a twin")).toBeInTheDocument();
    expect(row("pump").getByText("Not valid")).toBeInTheDocument();
  });

  it("shows the devices while their files are still being read", async () => {
    const bridge = withDevices();
    bridge.holdRead(fan.path);
    await show(bridge);
    expect(devices().getAllByText("Reading the file…")).toHaveLength(2);
  });

  it("is on the empty workspace page too", async () => {
    const bridge = new FakeBridge();
    bridge.workspaceOutcome = success({ ...sampleWorkspace, domains: [], compositions: [], runs: [] });
    await show(bridge);
    expect(screen.getByRole("heading", { name: /has no plants yet/ })).toBeInTheDocument();
    expect(devices().getByRole("button", { name: "Create device" })).toBeInTheDocument();
  });
});

describe("capturing from a board", () => {
  it("goes to the capture screen with a clean review, however much was read before", async () => {
    const bridge = withDevices();
    serveCapture(bridge);
    const desk = createDeviceDesk(bridge.devices);
    await desk.capture().read({ content: "{}" });
    const nav = navigation();
    await show(bridge, nav, undefined, desk);
    fireEvent.click(devices().getByRole("button", { name: "Capture from a board" }));
    expect(nav.location.hash).toBe("#/capture");
    expect(desk.capture().snapshot()).toEqual({ phase: "idle" });
  });

  it("gives the focus back to the button when the capture screen was left", async () => {
    const bridge = withDevices();
    const desk = createDeviceDesk(bridge.devices);
    desk.capture().leave();
    await show(bridge, navigation(), undefined, desk);
    expect(devices().getByRole("button", { name: "Capture from a board" })).toHaveFocus();
  });

  it("leaves the focus alone when the workspace is simply opened", async () => {
    await show(withDevices());
    expect(devices().getByRole("button", { name: "Capture from a board" })).not.toHaveFocus();
  });
});

describe("creating a device", () => {
  async function opened(bridge = withDevices(), nav = navigation()) {
    await show(bridge, nav);
    fireEvent.click(devices().getByRole("button", { name: "Create device" }));
    return { bridge, nav, form: screen.getByRole("form", { name: "New device" }) };
  }

  it("opens a form, explains a name it cannot use, and keeps Create off until the name fits", async () => {
    const { form } = await opened();
    expect(within(form).getByText("Give the device a name.")).toBeInTheDocument();
    const name = within(form).getByLabelText("Name");
    expect(name).toHaveFocus();
    fireEvent.change(name, { target: { value: "FAN" } });
    expect(within(form).getByText("A device named FAN already exists.")).toBeInTheDocument();
    expect(within(form).getByRole("button", { name: "Create and open the device" })).toHaveAttribute(
      "aria-disabled",
      "true",
    );
  });

  it("leaves the pattern of a name to the binary, which refuses it in plain words before anything is written", async () => {
    const { bridge, form } = await opened();
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "bad name" } });
    expect(within(form).getByRole("button", { name: "Create and open the device" })).toHaveAttribute(
      "aria-disabled",
      "false",
    );
    fireEvent.submit(form);
    expect(await within(form).findByRole("alert")).toHaveTextContent("Use letters, digits and . _ : - only");
    expect(bridge.writes).toHaveLength(0);
  });

  it("submits once however often it is submitted, and says it is working", async () => {
    const { bridge, nav, form } = await opened();
    const gate = deferred<Canned<unknown>>();
    bridge.handlers.set("kind.check", () => gate.promise);
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "heater" } });
    fireEvent.submit(form);
    fireEvent.submit(form);
    expect(within(form).getByRole("status")).toHaveTextContent("Creating the device…");
    gate.resolve(kindCheck({ content: '{"kind":"heater"}', twin: false }));
    await vi.waitFor(() => {
      expect(nav.location.hash).toBe("#/device?device=kinds%2Fheater.kind.json");
    });
    expect(bridge.writes).toHaveLength(1);
    expect(bridge.calls.filter((call) => call.name === "kind.check")).toHaveLength(1);
  });

  it("writes an empty device and opens it in the editor", async () => {
    const { bridge, nav, form } = await opened();
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "heater" } });
    fireEvent.submit(form);
    await vi.waitFor(() => {
      expect(nav.location.hash).toBe("#/device?device=kinds%2Fheater.kind.json");
    });
    expect(bridge.writes[0]).toMatchObject({ path: "kinds/heater.kind.json", etag: "" });
  });

  it("does not submit a name that does not fit", async () => {
    const { bridge, form } = await opened();
    fireEvent.submit(form);
    expect(bridge.writes).toHaveLength(0);
  });

  it("says why the device could not be created", async () => {
    const { bridge, form } = await opened();
    bridge.writeOutcomes.set("kinds/heater.kind.json", failed(cannedFailure("conflict", 409, "exists")));
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "heater" } });
    fireEvent.submit(form);
    expect(await within(form).findByRole("alert")).toHaveTextContent("A device named heater already exists.");
  });

  it("closes on Escape or Cancel and returns to the button", async () => {
    const { form } = await opened();
    fireEvent.keyDown(form, { key: "Escape" });
    expect(screen.queryByRole("form", { name: "New device" })).toBeNull();
    expect(devices().getByRole("button", { name: "Create device" })).toHaveFocus();
    fireEvent.click(devices().getByRole("button", { name: "Create device" }));
    fireEvent.click(screen.getByRole("button", { name: "Cancel" }));
    expect(screen.queryByRole("form", { name: "New device" })).toBeNull();
  });
});
