import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import { success } from "../../model/bridge";
import { parsePlant } from "../../model/plant";
import { FakeBridge } from "../../test/fakeBridge";
import { launchHarness } from "../../test/launchFixture";
import { domain } from "../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

const effectors = [
  { name: "start", args_schema: { required: ["chiller_id"], properties: { chiller_id: { type: "string" } } } },
  {
    name: "set_setpoint",
    args_schema: {
      required: ["chiller_id", "setpoint_c"],
      properties: { chiller_id: { type: "string" }, setpoint_c: { type: "number", minimum: 4, maximum: 12 } },
    },
  },
];

const plant = parsePlant({
  ...domain,
  kinds: domain.kinds.map((kind) => (kind.name === "chiller" ? { ...kind, effectors } : kind)),
});

function bridge() {
  const fake = new FakeBridge();
  fake.actions.set("world.create", success({ world_id: "w-7" }));
  fake.actions.set("clock.advance", success({}));
  fake.actions.set("run.end", success({}));
  fake.actions.set("world.destroy", success({}));
  fake.handlers.set("effector.invoke", (args) =>
    args.effector === "start"
      ? success({ accepted: false, reason: "interlock_refused", command_id: "director-1" })
      : success({ accepted: true, effect_eta_ns: 1e9, command_id: "director-2" }),
  );
  return fake;
}

function change(label: string, value: string) {
  fireEvent.change(screen.getByLabelText(label), { target: { value } });
}

function renderPanel(fake: FakeBridge, onFinished = vi.fn<(worldId: string) => void>()) {
  const { launch, scheduler } = launchHarness(fake);
  const view = render(
    <LaunchPanel launch={launch} scheduler={scheduler} plant={plant} path="p.json" onClose={vi.fn()} onFinished={onFinished} />,
  );
  return { view, onFinished };
}

describe("effector calls in the run plan", () => {
  it("adds a call row with the effector's arguments, checks it and says what the plan holds", async () => {
    const { view } = renderPanel(bridge());
    expect(screen.getByText(/^None: nothing is commanded/)).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Add effector call" }));
    expect(screen.getByLabelText("Effector call 1 instance")).toHaveFocus();
    change("Effector call 1 instance", "CH-1");
    change("Effector call 1 effector", "set_setpoint");
    expect((await axe.run(view.container)).violations).toEqual([]);
    change("Effector call 1 setpoint c", "20");
    change("Effector call 1 starts at minute", "5");
    expect(
      screen.getByText(/^Start simulates 30 min of plant time.* with no faults and 1 effector call,/),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    const field = screen.getByLabelText("Effector call 1 setpoint c");
    expect(field).toHaveFocus();
    expect(field).toHaveAccessibleDescription("Effector call 1: setpoint c is a number from 4 to 12.");
    fireEvent.click(screen.getByRole("button", { name: "Remove effector call 1" }));
    expect(screen.getByRole("button", { name: "Add effector call" })).toHaveFocus();
  });

  it("keeps the review closed until the user has read what the plant accepted and refused", async () => {
    const fake = bridge();
    const { onFinished } = renderPanel(fake);
    fireEvent.click(screen.getByRole("button", { name: "Add effector call" }));
    change("Effector call 1 instance", "CH-1");
    change("Effector call 1 effector", "start");
    fireEvent.click(screen.getByRole("button", { name: "Add effector call" }));
    change("Effector call 2 instance", "CH-1");
    change("Effector call 2 effector", "set_setpoint");
    change("Effector call 2 setpoint c", "7");
    change("Effector call 2 starts at minute", "10");
    await act(async () => {
      fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
      fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
      await Promise.resolve();
    });
    const list = await screen.findByRole("list", { name: "Effector call results" });
    expect(list).toHaveTextContent("00:00:00 · CH-1 · start: command not accepted: interlock_refused");
    expect(list).toHaveTextContent("00:10:00 · CH-1 · set setpoint: accepted");
    expect(screen.getByRole("status")).toHaveTextContent("1 of 2 effector calls were not accepted.");
    expect(onFinished).not.toHaveBeenCalled();
    expect(fake.calls.filter((call) => call.name === "effector.invoke")[1]?.args).toEqual({
      world_id: "w-7",
      entity_id: "CH-1",
      effector: "set_setpoint",
      args: { chiller_id: "CH-1", setpoint_c: 7 },
    });
    fireEvent.click(screen.getByRole("button", { name: "Open the review" }));
    expect(onFinished).toHaveBeenCalledWith("w-7");
  });
});
