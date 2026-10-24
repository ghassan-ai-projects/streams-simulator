import { fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { stubResizeObserver } from "../../../test/dom";
import type { FakeBridge } from "../../../test/fakeBridge";
import { success } from "../../../test/outcome";
import { twinBridge } from "../../../test/twinFixture";
import { behaviourOf, createTwin, inCommand, inReading, pickStep } from "../../../test/twinHarness";
import { sampleWorkspace } from "../../../test/workspaceFixture";

const indicatorPath = "kinds/indicator.kind.json";
const indicator = {
  kind: "indicator",
  metrics: [
    {
      name: "level_percent",
      data_type: "number",
      unit: "%",
      eu_range: { min: 0, max: 100 },
      resolution: 1,
      scan_class: "fast",
    },
  ],
  commands: [],
};
const setting = "The setting scan_class of reading 'level_percent' is not reproduced by a twin.";

function withIndicator(): FakeBridge {
  const fake = twinBridge();
  fake.addFile(indicatorPath, indicator, "k5");
  const listing = fake.workspaceOutcome;
  if (!listing.ok) throw new Error("no listing");
  const kinds = [...listing.value.kinds, { path: indicatorPath, kind: "indicator", title: "" }];
  fake.workspaceOutcome = success({ ...sampleWorkspace, kinds });
  return fake;
}

async function addIndicator() {
  const app = await behaviourOf(withIndicator());
  fireEvent.click(screen.getByRole("button", { name: "Add indicator to the twin" }));
  const notice = within(await screen.findByRole("region", { name: "Devices that are not ready for a twin" }));
  return { app, notice };
}

describe("a device with a setting a twin does not reproduce", () => {
  it("names the setting, the reading and the consequence, and keeps the binary's words under Details", async () => {
    stubResizeObserver();
    const { app, notice } = await addIndicator();
    expect(notice.getByText(setting)).toBeInTheDocument();
    expect(notice.getByText(/^indicator is left out of the twin as a whole/)).toBeInTheDocument();
    const details = notice.getByText("Details").closest("details");
    expect(details).not.toHaveAttribute("open");
    expect(
      within(details as HTMLElement).getByText("is not part of the interface a twin reproduces"),
    ).toBeInTheDocument();
    expect((await axe.run(app.view.container)).violations).toEqual([]);
  });

  it("leaves the device out of the summary of what goes into the twin, with its reason", async () => {
    await addIndicator();
    const scope = within(screen.getByRole("region", { name: "What goes into this twin" }));
    expect(scope.getByText("Included").nextElementSibling).toHaveTextContent("cooler");
    expect(scope.getByText("Excluded").nextElementSibling).toHaveTextContent(
      "indicator: the setting scan_class of reading 'level_percent' is not reproduced by a twin",
    );
    expect(scope.getByText("Blocked").nextElementSibling).toHaveTextContent("None");
  });

  it("creates a partial twin that says what it excludes in the status line and in the check", async () => {
    stubResizeObserver();
    await addIndicator();
    createTwin();
    const excludes = `This twin excludes indicator (${setting.slice(0, -1).replace("The", "the")}).`;
    expect(await screen.findByText(`Partial twin created. ${excludes}`)).toBeInTheDocument();
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    expect(await check.findByText("Partial twin")).toBeInTheDocument();
    expect(check.getAllByText(excludes).length).toBeGreaterThan(0);
  });
});

describe("the choices the binary suggested", () => {
  it("says how many are still the suggestion and labels each one, without blocking Create", async () => {
    await behaviourOf();
    const note = screen.getByRole("region", { name: "Suggested choices" });
    expect(note).toHaveTextContent(
      "3 choices are still the suggested defaults: review the instrument and response of each reading.",
    );
    expect(screen.getByRole("button", { name: "Create twin" })).toHaveAttribute("aria-disabled", "false");
    expect(inReading("temperature").getByText("Suggested")).toBeInTheDocument();
    expect(inCommand("set_duty").getByText("Suggested")).toBeInTheDocument();
  });

  it("calls a choice the person's own once changed, and counts what is left", async () => {
    await behaviourOf();
    const temperature = inReading("temperature");
    fireEvent.click(temperature.getByRole("button", { name: "Change the instrument of temperature" }));
    fireEvent.click(temperature.getByRole("radio", { name: /Thermistor/ }));
    fireEvent.click(temperature.getByRole("button", { name: "Use this instrument" }));
    expect(temperature.getByText("Your choice")).toBeInTheDocument();
    expect(screen.getByRole("region", { name: "Suggested choices" })).toHaveTextContent("2 choices are");
    expect(temperature.getByText(/On record: degC, -40 to 125/)).toBeInTheDocument();
  });

  it("goes away when the person keeps the suggestions on purpose", async () => {
    await behaviourOf();
    fireEvent.click(screen.getByRole("button", { name: "Keep the suggestions" }));
    expect(screen.queryByRole("region", { name: "Suggested choices" })).toBeNull();
    expect(inCommand("set_duty").getByText("Your choice")).toBeInTheDocument();
  });
});

describe("what the twin reproduces", () => {
  it("is written in the twin details, with the stand-in warning", async () => {
    await behaviourOf();
    const details = within(screen.getByRole("complementary", { name: "Twin details" }));
    expect(details.getByRole("heading", { name: "What this twin reproduces" })).toBeInTheDocument();
    expect(details.getByRole("heading", { name: "What it does not reproduce" })).toBeInTheDocument();
    expect(details.getByText(/Device firmware behaviour/)).toBeInTheDocument();
    expect(details.getByText(/simple stand-in until it is compared with recorded field data/)).toBeInTheDocument();
  });
});

describe("the four claims of the check", () => {
  it("shows interface, plausibility, calibration and hardware validation as separate rows", async () => {
    stubResizeObserver();
    const { view } = await behaviourOf();
    createTwin();
    await screen.findByText(/^Twin created\./);
    pickStep(/^Check/);
    const check = within(await screen.findByRole("region", { name: "Check of cooler" }));
    expect(await check.findByText("Interface: the twin matches its device")).toBeInTheDocument();
    expect(check.getByText("Plausibility: readings stay plausible in a nominal day")).toBeInTheDocument();
    expect(check.getByText("Calibration: not performed")).toBeInTheDocument();
    expect(check.getByText("Hardware validation: not run")).toBeInTheDocument();
    expect(check.getByText("Not performed")).toBeInTheDocument();
    expect(
      check.getByText("Interface check passed; plausibility check passed; not calibrated; not compared with a board."),
    ).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
