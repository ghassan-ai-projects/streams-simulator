import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { launchBridge, launchHarness } from "../../../test/launchFixture";
import { domain } from "../../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

function plantWith(unit: Record<string, string>) {
  const effectors = [
    {
      name: "set_setpoint",
      args_schema: {
        required: ["chiller_id", "setpoint_c"],
        properties: {
          chiller_id: { type: "string" },
          setpoint_c: { type: "number", minimum: 4, maximum: 12, ...unit },
        },
      },
    },
  ];
  return parsePlant({
    ...domain,
    kinds: domain.kinds.map((kind) => (kind.name === "chiller" ? { ...kind, effectors } : kind)),
  });
}

function renderWith(unit: Record<string, string>) {
  const { launch } = launchHarness(launchBridge());
  render(<LaunchPanel launch={launch} plant={plantWith(unit)} path="p.json" onClose={vi.fn()} onFinished={vi.fn()} />);
  fireEvent.click(screen.getByRole("button", { name: "Add effector call" }));
  fireEvent.change(screen.getByLabelText("Effector call 1 instance"), { target: { value: "CH-1" } });
  fireEvent.change(screen.getByLabelText("Effector call 1 effector"), { target: { value: "set_setpoint" } });
  fireEvent.change(screen.getByLabelText("Effector call 1 setpoint c"), { target: { value: "7" } });
}

describe("the unit an effector argument declares, in the run plan", () => {
  it("follows the argument's name in the call row and after its value in the agenda", () => {
    renderWith({ unit: "degC" });
    expect(screen.getByText("setpoint c (°C)")).toBeInTheDocument();
    expect(screen.getByText("CH-1 set setpoint (setpoint c 7 °C)")).toBeInTheDocument();
    expect(screen.getByLabelText("Effector call 1 setpoint c")).toBeInTheDocument();
  });

  it("shows none when the argument declares none", () => {
    renderWith({});
    expect(screen.getByText("setpoint c")).toBeInTheDocument();
    expect(screen.getByText("CH-1 set setpoint (setpoint c 7)")).toBeInTheDocument();
  });
});
