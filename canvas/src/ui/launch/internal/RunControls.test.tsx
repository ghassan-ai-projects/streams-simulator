import { screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { failure as failed } from "../../../test/outcome";
import { cannedFailure } from "../../../test/fakeBridge";
import { launchBridge } from "../../../test/launchFixture";
import { runningPanel } from "../../../test/runningPanel";

const clock = () => screen.getByText(/ of 00:30:00 simulated/).textContent;

describe("the controls of a running run", () => {
  it("offers pause, a step that waits for the pause, and the speed the run goes at", async () => {
    const { view } = await runningPanel();
    expect(screen.getByRole("group", { name: "Controls" })).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pause" })).toHaveAttribute("aria-disabled", "false");
    expect(screen.getByRole("button", { name: "Step one chunk" })).toHaveAttribute("aria-disabled", "true");
    const speed = screen.getByRole<HTMLSelectElement>("combobox", { name: "Speed" });
    expect(speed.value).toBe("60");
    expect([...speed.options].map((option) => option.textContent)).toEqual([
      "1 s per s",
      "10 s per s",
      "60 s per s",
      "300 s per s",
      "3600 s per s",
      "As fast as possible",
    ]);
    expect(screen.getByRole("button", { name: "Stop now" })).toBeInTheDocument();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("pauses where the clock stands, says so, lets the user step, and resumes", async () => {
    const { tick, press, view } = await runningPanel();
    await tick(1000);
    expect(clock()).toBe("Running · 00:01:00 of 00:30:00 simulated · 60 s per s");
    await press("Pause");
    expect(clock()).toBe("Paused · 00:01:00 of 00:30:00 simulated");
    expect(screen.getByText("Paused Mini plant as run w-7")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Resume" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Pause" })).toBeNull();
    await tick(3000);
    expect(clock()).toBe("Paused · 00:01:00 of 00:30:00 simulated");
    expect(screen.getByRole("button", { name: "Step one chunk" })).toHaveAttribute("aria-disabled", "false");
    await press("Step one chunk");
    expect(clock()).toBe("Paused · 00:02:00 of 00:30:00 simulated");
    expect((await axe.run(view.container)).violations).toEqual([]);
    await press("Resume");
    await tick(1000);
    expect(clock()).toBe("Running · 00:03:00 of 00:30:00 simulated · 60 s per s");
    expect(screen.getByText("Running Mini plant as run w-7")).toBeInTheDocument();
  });

  it("changes the speed from the select and shows the live rate in the clock line and the select", async () => {
    const { choose, tick } = await runningPanel();
    await tick(500);
    await choose("Speed", "10");
    expect(screen.getByRole<HTMLSelectElement>("combobox", { name: "Speed" }).value).toBe("10");
    expect(clock()).toMatch(/· 10 s per s$/);
    await choose("Speed", "300");
    expect(screen.getByRole<HTMLSelectElement>("combobox", { name: "Speed" }).value).toBe("300");
    expect(clock()).toMatch(/· 300 s per s$/);
  });

  it("keeps a click on Step from doing anything while the run is not paused", async () => {
    const { press, bridge } = await runningPanel();
    await press("Step one chunk");
    expect(bridge.calls.map((call) => call.name)).not.toContain("job.step");
  });

  it("shows the director's refusal under the controls, as an alert, and clears it on the next try", async () => {
    const bridge = launchBridge("w-7");
    const { press } = await runningPanel({ bridge });
    bridge.actions.set("job.pause", failed(cannedFailure("job_not_running", 409, "the job is not running or paused")));
    await press("Pause");
    expect(screen.getByRole("alert")).toHaveTextContent(
      "The simulator did not do that: the job is not running or paused.",
    );
    bridge.actions.delete("job.pause");
    await press("Pause");
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("is operable from the keyboard: the controls are buttons and a select in the tab order", async () => {
    await runningPanel();
    const pause = screen.getByRole("button", { name: "Pause" });
    pause.focus();
    expect(pause).toHaveFocus();
    const select = screen.getByRole("combobox", { name: "Speed" });
    select.focus();
    expect(select).toHaveFocus();
    expect(pause).not.toHaveAttribute("tabindex");
    expect(select).not.toHaveAttribute("tabindex");
  });

  it("shows pending words while a control call is out and disables every control", async () => {
    const bridge = launchBridge("w-7");
    const { press } = await runningPanel({ bridge });
    const gate = bridge.holdAction("job.pause");
    const pressing = press("Pause");
    await Promise.resolve();
    expect(screen.getByText("Pausing…")).toBeInTheDocument();
    expect(screen.getByRole("button", { name: "Pause" })).toHaveAttribute("aria-disabled", "true");
    expect(screen.getByRole("combobox", { name: "Speed" })).toHaveAttribute("aria-disabled", "true");
    gate.release();
    await pressing;
    expect(screen.queryByText("Pausing…")).toBeNull();
  });
});
