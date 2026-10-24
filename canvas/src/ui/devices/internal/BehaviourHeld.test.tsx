import { act, fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { cannedFailure } from "../../../test/fakeBridge";
import type { FakeBridge } from "../../../test/fakeBridge";
import { failure, success } from "../../../test/outcome";
import { choicesPath, coolerDocument, coolerPath, suggestion, twinBridge } from "../../../test/twinFixture";
import { behaviourOf, hash, inCommand, inReading, pickStep } from "../../../test/twinHarness";
import { sampleWorkspace } from "../../../test/workspaceFixture";

const pumpPath = "kinds/pump_monitor.kind.json";
const unfinished = { kind: "pump_monitor", metrics: [{ name: "pressure", data_type: "number" }], commands: [] };
const finished = {
  kind: "pump_monitor",
  metrics: [{ name: "pressure", data_type: "number", unit: "%", eu_range: { min: 0, max: 100 }, resolution: 1 }],
  commands: [],
};

function withPump(): FakeBridge {
  const fake = twinBridge();
  fake.addFile(pumpPath, unfinished, "k3");
  const listing = fake.workspaceOutcome;
  if (!listing.ok) throw new Error("no listing");
  const kinds = [...listing.value.kinds, { path: pumpPath, kind: "pump_monitor", title: "" }];
  fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
  return fake;
}

function metricsWith(change: (metric: Record<string, unknown>) => Record<string, unknown>) {
  return coolerDocument.metrics.map((metric) => change({ ...metric }));
}

function withoutField(name: string, field: string) {
  return metricsWith((metric) =>
    metric.name === name ? Object.fromEntries(Object.entries(metric).filter(([key]) => key !== field)) : metric,
  );
}

function bareBridge(): FakeBridge {
  const fake = twinBridge();
  fake.addFile(coolerPath, { ...coolerDocument, metrics: withoutField("temperature", "eu_range") }, "k1");
  return fake;
}

function chip(name: RegExp) {
  return within(screen.getByRole("navigation", { name: "Device steps" })).getByRole("button", { name });
}

describe("a device that is not ready for a twin", () => {
  it("is named in a plain sentence, can be opened, and joins the twin once it is fixed, with no reload", async () => {
    const fake = withPump();
    await behaviourOf(fake);
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
    fireEvent.click(screen.getByRole("button", { name: "Add pump_monitor to the twin" }));
    const notice = within(await screen.findByRole("region", { name: "Devices that are not ready for a twin" }));
    expect(notice.getByText("Reading 'pressure' needs a unit.")).toBeInTheDocument();
    expect(notice.getByText("Reading 'pressure' needs a resolution.")).toBeInTheDocument();
    expect(notice.getByText(/^pump_monitor is left out of the twin as a whole/)).toBeInTheDocument();
    expect(notice.getByRole("button", { name: "Retry" })).toBeInTheDocument();
    expect(screen.queryByRole("group", { name: "Reading pressure" })).toBeNull();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
    fake.addFile(pumpPath, finished, "k4");
    act(() => {
      fake.changeOnDisk();
    });
    expect(await screen.findByRole("group", { name: "Reading pressure" })).toBeInTheDocument();
    expect(screen.queryByRole("region", { name: "Devices that are not ready for a twin" })).toBeNull();
    fireEvent.click(screen.getByRole("button", { name: "Leave pump_monitor out of the twin" }));
    expect(screen.queryByRole("group", { name: "Reading pressure" })).toBeNull();
  });

  it("has a link that opens it, and a Retry that works without a change of the file", async () => {
    const fake = withPump();
    const { location } = await behaviourOf(fake);
    fireEvent.click(screen.getByRole("button", { name: "Add pump_monitor to the twin" }));
    await screen.findByText("Reading 'pressure' needs a unit.");
    fake.addFile(pumpPath, finished, "k4");
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(await screen.findByRole("group", { name: "Reading pressure" })).toBeInTheDocument();
    fake.addFile(pumpPath, unfinished, "k5");
    fireEvent.click(screen.getByRole("button", { name: "Leave pump_monitor out of the twin" }));
    fireEvent.click(screen.getByRole("button", { name: "Add pump_monitor to the twin" }));
    fireEvent.click(await screen.findByRole("button", { name: "Open pump_monitor" }));
    expect(location.hash).toBe(`#/device?device=${encodeURIComponent(pumpPath)}`);
  });

  it("does not stop another device from having a behaviour", async () => {
    const fake = withPump();
    await behaviourOf(fake);
    expect(screen.getByLabelText("Twin name")).toHaveValue("cooler-sensor-twin");
    expect(screen.queryByRole("region", { name: "Devices that are not ready for a twin" })).toBeNull();
    const asked = fake.calls.filter((call) => call.name === "kind.suggest").map((call) => call.args);
    expect(asked.length).toBeGreaterThan(0);
    asked.forEach((args) => {
      expect(args).toEqual({ kinds: ["cooler"] });
    });
  });
});

describe("a behaviour that could not be opened", () => {
  it("says why, offers a Retry, and opens once the cause is gone", async () => {
    const fake = twinBridge();
    const good = fake.handlers.get("kind.suggest");
    fake.handlers.set("kind.suggest", () => failure(cannedFailure("boom", 500, "the bridge broke")));
    mountApp(fake, hash);
    await screen.findByRole("region", { name: "Editing cooler" });
    pickStep(/^Behaviour/);
    expect(await screen.findByRole("alert")).toHaveTextContent("The behaviour could not be opened: the bridge broke");
    if (good !== undefined) fake.handlers.set("kind.suggest", good);
    fireEvent.click(screen.getByRole("button", { name: "Retry" }));
    expect(await screen.findByRole("group", { name: "Reading temperature" })).toBeInTheDocument();
  });
});

describe("the instruments a reading may use", () => {
  it("greys out an instrument that cannot read the reading, with one plain reason, and enables a converted one", async () => {
    const wide = metricsWith((metric) =>
      metric.name === "duty" ? { ...metric, eu_range: { min: 0, max: 200 } } : metric,
    );
    const fake = twinBridge();
    fake.addFile(coolerPath, { ...coolerDocument, metrics: wide }, "k1");
    await behaviourOf(fake);
    const duty = inReading("duty");
    fireEvent.click(duty.getByRole("button", { name: "Change the instrument of duty" }));
    const humidity = duty.getByRole("radio", { name: /Capacitive humidity sensor/ });
    expect(humidity).toBeDisabled();
    expect(duty.getByText("This reading's range is wider than the instrument measures")).toBeInTheDocument();
    expect(duty.queryByText(/does not fit the reading/)).toBeNull();
    fireEvent.click(duty.getByRole("button", { name: "Close the list for duty" }));
    const temperature = inReading("temperature");
    fireEvent.click(temperature.getByRole("button", { name: "Change the instrument of temperature" }));
    const thermistor = temperature.getByRole("radio", { name: /Thermistor/ });
    expect(thermistor).toBeEnabled();
    expect(temperature.getByText(/measured in degC, converted/)).toBeInTheDocument();
  });

  it("asks for the range of a reading that states none, and leads to its field", async () => {
    await behaviourOf(bareBridge());
    const temperature = inReading("temperature");
    fireEvent.click(temperature.getByRole("button", { name: "Change the instrument of temperature" }));
    expect(temperature.getByRole("radio", { name: /Thermistor/ })).toBeDisabled();
    expect(temperature.getByText("Set a range for this reading in the Interface step")).toBeInTheDocument();
    fireEvent.click(temperature.getByRole("button", { name: "Set it in the Interface step" }));
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
    expect(screen.getByLabelText("Minimum")).toHaveFocus();
  });

  it("mentions the missing range in the interface only for an instrument that was chosen and needs it", async () => {
    const fake = bareBridge();
    const chosen = { instrument: "streamsim/thermistor@1" };
    const kind = { kind: "cooler", instance: "cooler-01", metrics: { temperature: chosen }, commands: {} };
    fake.addFile(choicesPath, { ...suggestion, kinds: [kind] }, "c1");
    mountApp(fake, hash);
    expect(await screen.findByText("Needed for a twin: 1")).toBeInTheDocument();
    const needs = within(screen.getByRole("region", { name: "Needed for a twin" }));
    expect(needs.getByText("Reading temperature · Range")).toBeInTheDocument();
    pickStep(/^Behaviour/);
    const temperature = inReading("temperature");
    expect(
      await temperature.findByText("A range is needed for the instrument chosen for this reading"),
    ).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByText("Set what the twin needs in the Interface step first.")).toBeInTheDocument();
    expect(chip(/^Behaviour/)).toHaveTextContent("Needs attention");
    fireEvent.click(temperature.getByRole("button", { name: "Set it in the Interface step" }));
    fireEvent.change(await screen.findByLabelText("Minimum"), { target: { value: "0" } });
    fireEvent.change(screen.getByLabelText("Maximum"), { target: { value: "100" } });
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(await screen.findByText("Saved and valid.")).toBeInTheDocument();
    expect(await screen.findByText("Ready for a twin")).toBeInTheDocument();
    expect(screen.queryByText(/Needed for a twin/)).toBeNull();
  });
});

describe("the command that sets a reading", () => {
  it("says what a setpoint needs when no reading is one, and leads to the Interface step", async () => {
    const fake = twinBridge();
    const plain = metricsWith((metric) => ({ ...metric, writable: false }));
    fake.addFile(coolerPath, { ...coolerDocument, metrics: plain }, "k1");
    await behaviourOf(fake);
    const command = inCommand("set_duty");
    expect(
      command.getAllByText("Mark a reading as a setpoint and give it a range in the Interface step").length,
    ).toBeGreaterThan(0);
    fireEvent.click(command.getByRole("button", { name: "Set it in the Interface step" }));
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
    expect(screen.getByLabelText("Setpoint")).toHaveFocus();
  });

  it("chooses the only reading for the only number argument", async () => {
    const fake = twinBridge();
    const one = {
      name: "set_duty",
      risk_class: "medium",
      args_schema: { type: "object", properties: { duty: { type: "number", minimum: 0, maximum: 100 } } },
    };
    fake.addFile(coolerPath, { ...coolerDocument, commands: [one] }, "k1");
    await behaviourOf(fake);
    const command = inCommand("set_duty");
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "none" } });
    fireEvent.change(command.getByLabelText("Response"), { target: { value: "direct" } });
    expect(command.getByLabelText("Reading set by the duty value")).toHaveDisplayValue("duty");
    expect(command.queryByText("Choose the reading this command sets.")).toBeNull();
  });
});

describe("the badge of the behaviour step", () => {
  it("is never Ready while the banner says to save the interface first", async () => {
    await behaviourOf();
    pickStep(/^Interface/);
    fireEvent.change(screen.getByLabelText("Name"), { target: { value: "cooler2" } });
    pickStep(/^Behaviour/);
    expect(await screen.findByText(/Save the interface first/)).toBeInTheDocument();
    expect(chip(/^Behaviour/)).toHaveTextContent("Needs attention");
    expect(chip(/^Behaviour/)).not.toHaveTextContent("Ready");
  });

  it("shows the reason beside a disabled Create twin, and links it to the button", async () => {
    await behaviourOf();
    fireEvent.change(screen.getByLabelText("Twin name"), { target: { value: "" } });
    const create = screen.getByRole("button", { name: "Create twin" });
    expect(create).toHaveAttribute("aria-disabled", "true");
    expect(create).toHaveAccessibleDescription("Finish the choices marked in the form first.");
  });
});
