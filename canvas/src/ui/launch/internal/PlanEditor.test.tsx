import { act, fireEvent, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { flushPromises } from "../../../test/deferred";
import { cannedFailure } from "../../../test/fakeBridge";
import { launchBridge } from "../../../test/launchFixture";
import { failure as failed } from "../../../test/outcome";
import { change, press, renderPanel, start } from "../../../test/planEditorHarness";

function at<T>(items: readonly T[], index: number): T {
  const item = items[index];
  if (item === undefined) throw new Error(`no item ${String(index)}`);
  return item;
}

describe("name and purpose", () => {
  it("offers both as optional, with examples and counts, and starts the run with them", async () => {
    const { fake } = await renderPanel();
    const name = screen.getByLabelText("Name");
    expect(name).toHaveAttribute("placeholder", "Fan failure and recovery");
    expect(name).toHaveAttribute("maxlength", "80");
    expect(screen.getByLabelText("Purpose")).toHaveAttribute("maxlength", "500");
    expect(name).toHaveAccessibleDescription(/Optional\. .* 0 of 80/);
    change("Name", "Cooling loss");
    const purpose = "Observe the chiller with the tower fan out";
    change("Purpose", purpose);
    expect(screen.getByText("12 of 80")).toBeInTheDocument();
    expect(screen.getByText(`${String(purpose.length)} of 500`)).toBeInTheDocument();
    const args = await start(fake);
    expect(args).toMatchObject({ name: "Cooling loss", purpose });
  });

  it("sends neither when they are left blank, and says what an unset purpose means", async () => {
    const { fake } = await renderPanel();
    change("Name", "   ");
    expect(screen.getByText(/investigation without a stated question/, { selector: "small" })).toBeInTheDocument();
    const args = await start(fake);
    expect(args).not.toHaveProperty("name");
    expect(args).not.toHaveProperty("purpose");
  });
});

describe("a fault that recovers", () => {
  it("reads as starting and ending, and sends the end", async () => {
    const { fake } = await renderPanel();
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    change("Fault 1 fault", "chiller__fouling");
    change("Fault 1 starts at minute", "5");
    expect(screen.getByLabelText("Fault 1 ends at minute")).toHaveAttribute("placeholder", "never");
    change("Fault 1 ends at minute", "12.5");
    const args = await start(fake);
    expect(args?.faults).toEqual([{ entity_id: "CH-1", fault: "chiller__fouling", at_s: 300, until_s: 750 }]);
  });

  it("refuses an end that does not follow the start, focuses it and names it", async () => {
    const { fake } = await renderPanel();
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    change("Fault 1 fault", "chiller__fouling");
    change("Fault 1 starts at minute", "5");
    change("Fault 1 ends at minute", "5");
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    press("Start the run");
    const end = screen.getByLabelText("Fault 1 ends at minute");
    expect(end).toHaveFocus();
    expect(end).toHaveAccessibleDescription("Fault 1 ends after it starts.");
    expect(screen.getByLabelText("Fault 1 starts at minute")).not.toHaveAttribute("aria-invalid");
    expect(fake.jobCalls).toEqual([]);
  });

  it("groups the faults of an instance by the layer they act on", async () => {
    await renderPanel();
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    const select = screen.getByLabelText("Fault 1 fault");
    const groups = within(select).getAllByRole("group");
    expect(groups.map((group) => group.getAttribute("label"))).toEqual([
      "Process (world)",
      "Measurement (sensor)",
      "World and measurement",
    ]);
    expect(within(at(groups, 1)).getByRole("option", { name: "sensor drift" })).toBeInTheDocument();
  });

  it("keeps the list flat when the binary names no categories", async () => {
    const fake = launchBridge();
    fake.actions.set("catalog.describe", { ok: true, value: { spec: {}, digest: "x" } });
    await renderPanel(fake);
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    expect(within(screen.getByLabelText("Fault 1 fault")).queryAllByRole("group")).toEqual([]);
    expect(screen.getByRole("option", { name: "stuck valve" })).toBeInTheDocument();
  });
});

describe("the plan in time order", () => {
  it("lists every kind of step together, earliest first, and takes the focus to the step's editor", async () => {
    await renderPanel();
    press("Add effector call");
    change("Effector call 1 instance", "CH-1");
    change("Effector call 1 effector", "chiller__set_setpoint");
    change("Effector call 1 starts at minute", "15");
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    change("Fault 1 fault", "chiller__sensor_drift");
    change("Fault 1 starts at minute", "5");
    change("Fault 1 ends at minute", "12");
    press("Add delivery condition");
    change("Delivery condition 1 kind", "reorder");
    change("Delivery condition 1 starts at minute", "3");
    const list = screen.getByRole("list", { name: "Plan in time order" });
    const items = within(list).getAllByRole("listitem");
    expect(items.map((item) => item.textContent)).toEqual([
      "03:00Deliveryreorder starts",
      "05:00Sensor faultCH-1 sensor drift",
      "12:00ClearCH-1 sensor drift ends",
      "15:00CommandCH-1 set setpoint",
    ]);
    fireEvent.click(within(at(items, 2)).getByRole("button"));
    expect(screen.getByLabelText("Fault 1 ends at minute")).toHaveFocus();
    fireEvent.click(within(at(items, 0)).getByRole("button"));
    expect(screen.getByLabelText("Delivery condition 1 kind")).toHaveFocus();
    fireEvent.click(within(at(items, 3)).getByRole("button"));
    expect(screen.getByLabelText("Effector call 1 instance")).toHaveFocus();
  });

  it("says that nothing is planned yet", async () => {
    await renderPanel();
    expect(screen.getByText(/^Nothing yet: add a fault/)).toBeInTheDocument();
    expect(screen.queryByRole("list", { name: "Plan in time order" })).toBeNull();
  });
});

describe("before you start", () => {
  it("reviews the plan and says plainly that nothing is scoring the run", async () => {
    await renderPanel();
    change("Name", "Cooling loss");
    press("Add fault");
    const review = screen.getByRole("region", { name: "Before you start" });
    expect(within(review).getByText("Mini plant")).toBeInTheDocument();
    expect(within(review).getByText("Cooling loss")).toBeInTheDocument();
    expect(within(review).getByText(/^Explore\. The truth is visible/)).toBeInTheDocument();
    expect(within(review).getByText("30 min of simulated time, seed 1")).toBeInTheDocument();
    expect(within(review).getByText("1 fault")).toBeInTheDocument();
    expect(
      within(review).getByText("Unscored investigation: no outside system is scoring this run."),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("radio", { name: /^Evaluate/ }));
    expect(
      within(review).getByText(/^Unscored blind run: no outside system is scoring this run\./),
    ).toBeInTheDocument();
    expect(within(review).getByText(/set up a connected session instead/)).toBeInTheDocument();
  });
});

describe("a plan the binary refuses", () => {
  it("shows the binary's own message and returns to the plan with everything still entered", async () => {
    const fake = launchBridge();
    fake.actions.set(
      "job.start",
      failed(cannedFailure("invalid_args", 422, "job: faults[0]: until_s must be after at_s")),
    );
    await renderPanel(fake);
    change("Name", "Cooling loss");
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    change("Fault 1 fault", "chiller__fouling");
    change("Fault 1 ends at minute", "20");
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    await act(async () => {
      press("Start the run");
      await flushPromises();
    });
    expect(screen.getByRole("alert")).toHaveTextContent("job: faults[0]: until_s must be after at_s. Check the plan");
    fireEvent.click(screen.getByRole("button", { name: "Change the plan" }));
    expect(screen.getByLabelText("Name")).toHaveValue("Cooling loss");
    expect(screen.getByLabelText("Fault 1 ends at minute")).toHaveValue("20");
  });
});

describe("accessibility of the plan editor", () => {
  it("passes axe with every kind of step planned", async () => {
    const { view } = await renderPanel();
    change("Name", "Cooling loss");
    press("Add fault");
    change("Fault 1 instance", "CH-1");
    press("Add delivery condition");
    change("Delivery condition 1 kind", "producer_flap");
    press("Add effector call");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
