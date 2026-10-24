import { screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { parsePlant } from "../../../model/plant";
import { launchBridge } from "../../../test/launchFixture";
import { attributeValues, canary, namedBy, roles } from "../../../test/runningPlant";
import { runningPanel } from "../../../test/runningPanel";
import { domain } from "../../../test/plantFixture";

const names = (bridge: ReturnType<typeof launchBridge>) => bridge.calls.map((call) => call.name);

describe("acting now in an Explore run", () => {
  it("offers a call and a fault, each with its own instance picker, and passes axe", async () => {
    const { view } = await runningPanel();
    const group = screen.getByRole("group", { name: "Act now" });
    expect(group).toBeInTheDocument();
    for (const label of ["Call now instance", "Call now effector", "Fault now instance", "Fault now fault"]) {
      expect(screen.getByLabelText(label)).toBeInTheDocument();
    }
    expect(screen.getByRole("button", { name: "Call now" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Inject now" })).toHaveAttribute("aria-disabled", "true");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("blocks an argument outside its declared bounds before anything is sent, and focuses the field", async () => {
    const { choose, press, bridge, tick } = await runningPanel();
    await tick(500);
    await choose("Call now instance", "CH-1");
    await choose("Call now effector", "set_setpoint");
    await choose("Call now setpoint c", "20");
    await press("Call now");
    expect(names(bridge)).not.toContain("job.invoke_effector");
    const field = screen.getByLabelText("Call now setpoint c");
    expect(field).toHaveFocus();
    expect(field).toHaveAccessibleDescription("The call: setpoint c is a number from 4 to 12.");
    await choose("Call now setpoint c", "7");
    expect(screen.queryByText(/is a number from 4 to 12/)).toBeNull();
    await press("Call now");
    expect(bridge.calls.find((call) => call.name === "job.invoke_effector")?.args).toEqual({
      world_id: "w-7",
      entity_id: "CH-1",
      effector: "set_setpoint",
      args: { chiller_id: "CH-1", setpoint_c: 7 },
    });
  });

  it("asks for the missing instance or effector in words and puts the focus there", async () => {
    const { press } = await runningPanel();
    await press("Call now");
    expect(screen.getByLabelText("Call now instance")).toHaveFocus();
    expect(screen.getByText("The call needs an instance and an effector.")).toBeInTheDocument();
  });

  it("lists the outcome of a call given now, marked as now, with the simulator's reason when refused", async () => {
    const bridge = launchBridge("w-7");
    bridge.jobs.refuse("start", "interlock_refused");
    const { choose, press, tick } = await runningPanel({ bridge });
    await tick(500);
    await choose("Call now instance", "CH-1");
    await choose("Call now effector", "set_setpoint");
    await choose("Call now setpoint c", "7");
    await press("Call now");
    await choose("Call now effector", "start");
    await press("Call now");
    const list = screen.getByRole("list", { name: "Commands given now" });
    expect(list).toHaveTextContent("00:00:30 · CH-1 · set setpoint: accepted · now");
    expect(list).toHaveTextContent("00:00:30 · CH-1 · start: command not accepted: interlock_refused · now");
  });

  it("injects a fault now and lists it, at the time it was injected", async () => {
    const { choose, press, bridge, tick } = await runningPanel();
    await tick(500);
    await choose("Fault now instance", "CH-1");
    await choose("Fault now fault", "chiller__fouling");
    await press("Inject now");
    expect(bridge.calls.find((call) => call.name === "job.inject_fault")?.args).toEqual({
      world_id: "w-7",
      entity_id: "CH-1",
      fault: "chiller__fouling",
    });
    expect(screen.getByRole("list", { name: "Commands given now" })).toHaveTextContent(
      "00:00:30 · CH-1 · fault injected: fouling · now",
    );
  });
});

describe("acting now in an Evaluate run", () => {
  it("offers no fault injection, says why in one line, and still lets the operator call an effector", async () => {
    const { choose, press, bridge, tick } = await runningPanel({ mode: "Evaluate" });
    await tick(500);
    expect(screen.queryByRole("button", { name: "Inject now" })).toBeNull();
    expect(screen.queryByLabelText(/^Fault now/)).toBeNull();
    expect(screen.getByText("Faults can only be injected live in Explore runs.")).toBeInTheDocument();
    await choose("Call now instance", "CH-1");
    await choose("Call now effector", "start");
    await press("Call now");
    expect(names(bridge)).toContain("job.invoke_effector");
    expect(screen.getByRole("list", { name: "Commands given now" })).toHaveTextContent("start: accepted · now");
  });

  it("names no fault of the plant anywhere on the page: text, attributes or accessible names", async () => {
    const hidden = parsePlant({
      ...domain,
      kinds: domain.kinds.map((kind) =>
        kind.name === "chiller" ? { ...kind, faults: [{ id: "chiller__CANARY_TRUTH_9f3a" }] } : kind,
      ),
    });
    const { choose, press, tick, view } = await runningPanel({ mode: "Evaluate", plant: hidden });
    await tick(500);
    await choose("Call now instance", "CH-1");
    await choose("Call now effector", "start");
    await press("Call now");
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(roles.map(namedBy)).toEqual(roles.map(() => 0));
    expect((await axe.run(view.container)).violations).toEqual([]);
  });
});
