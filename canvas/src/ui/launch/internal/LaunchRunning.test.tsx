import { act, fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { parsePlant } from "../../../model/plant";
import { elapse, launchBridge, launchHarness } from "../../../test/launchFixture";
import { domain } from "../../../test/plantFixture";
import { LaunchPanel } from "./LaunchPanel";

const path = "domains/mini.domain.json";
describe("launch panel running", () => {
  it("keeps a run going when the panel closes and opens its review when it is back on that plant", async () => {
    const fake = launchBridge();
    const { launch, timers } = launchHarness(fake);
    const onFinished = vi.fn<(worldId: string) => void>();
    const plant = parsePlant(domain);
    const panel = () => (
      <LaunchPanel launch={launch} plant={plant} path={path} onClose={vi.fn()} onFinished={onFinished} />
    );
    const first = render(panel());
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    fireEvent.click(screen.getByRole("button", { name: "Start the run" }));
    await screen.findByText("Running Mini plant as run w-7");
    first.unmount();
    expect(launch.snapshot()).toMatchObject({ phase: "running", stopping: false, origin: { domain: "mini", path } });
    render(panel());
    expect(screen.getByText("Running Mini plant as run w-7")).toBeInTheDocument();
    await act(async () => {
      await elapse(timers, 500);
    });
    expect(screen.getByRole("status")).toHaveTextContent(/of simulated time · .* · no faults\.$/);
    expect(screen.getByRole("status")).not.toHaveTextContent("finished");
    expect(onFinished).not.toHaveBeenCalled();
    act(() => {
      timers.advance(1999);
    });
    expect(onFinished).not.toHaveBeenCalled();
    act(() => {
      timers.advance(1);
    });
    expect(onFinished).toHaveBeenCalledWith("w-7");
  });

  it("tells another plant that a run is in progress and leaves a finished run of another plant alone", async () => {
    const fake = launchBridge();
    const { launch, timers } = launchHarness(fake);
    const elsewhere = { domain: "other", title: "Other plant", path: "domains/other.domain.json" };
    void launch.start(elsewhere, { seed: 1, durationS: 60, explore: true, faults: [], perturbations: [], calls: [] });
    const onFinished = vi.fn<(worldId: string) => void>();
    render(
      <LaunchPanel launch={launch} plant={parsePlant(domain)} path={path} onClose={vi.fn()} onFinished={onFinished} />,
    );
    expect(await screen.findByText("A run of Other plant is in progress.")).toBeInTheDocument();
    await act(async () => {
      await elapse(timers, 500);
    });
    expect(launch.snapshot()).toMatchObject({ phase: "done", origin: { domain: "other" } });
    expect(onFinished).not.toHaveBeenCalled();
    expect(screen.getByRole("button", { name: "Start the run" })).toBeInTheDocument();
  });
});
