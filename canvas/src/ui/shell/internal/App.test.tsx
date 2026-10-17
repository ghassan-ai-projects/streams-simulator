import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { createSession } from "../../../app/session";
import { PanelStore } from "../../../app/shell";
import { LaunchStore } from "../../../app/launch";
import { PlantMaker } from "../../../app/plant";
import { NavigationStore } from "../../../app/shell";
import { ReviewStore } from "../../../app/review";
import { ThemeStore } from "../../../app/shell";
import { WorkspaceStore } from "../../../app/workspace";
import { failure as failed, success } from "../../../test/outcome";
import { FakeBridge, cannedFailure } from "../../../test/fakeBridge";
import { memoryLocation } from "../../../test/location";
import { fakeClock } from "../../../test/fakeClock";
import { App } from "./App";
import { editorStore, plantStore } from "../../../test/stores";

function themes() {
  let value: string | null = null;
  return new ThemeStore(
    { read: () => value, write: (_k, v) => (value = v) },
    { prefersDark: () => false, subscribe: () => () => undefined },
  );
}

function renderApp(bridge: FakeBridge) {
  const location = memoryLocation();
  const timers = fakeClock();
  const stores = {
    themes: themes(),
    session: createSession(bridge.session),
    workspace: new WorkspaceStore(bridge.workspace),
    navigation: new NavigationStore(location),
    plants: plantStore(bridge.plant),
    review: new ReviewStore(bridge.review, timers, timers),
    launch: new LaunchStore(bridge.launch, timers, timers),
    scheduler: timers,
    editor: editorStore(bridge.editor),
    panels: new PanelStore({ read: () => null, write: () => undefined }),
    maker: new PlantMaker(bridge.plant),
    showDetails: () => undefined,
  };
  return render(<App stores={stores} />);
}

describe("app shell", () => {
  it("opens the workspace once paired, with no accessibility violation", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    const { container } = renderApp(bridge);
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it("explains an unreachable bridge", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = failed(cannedFailure("unreachable", 0, "The bridge did not answer."));
    renderApp(bridge);
    expect(await screen.findByRole("heading", { name: "The bridge is not reachable" })).toBeInTheDocument();
    expect(screen.getByText(/The bridge did not answer\./)).toBeInTheDocument();
  });

  it("cycles the theme from the top bar and marks the document", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = success(true);
    renderApp(bridge);
    await screen.findByRole("heading", { name: "lab-workspace" });
    fireEvent.click(screen.getByRole("button", { name: "Theme: system" }));
    expect(screen.getByRole("button", { name: "Theme: light" })).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe("light");
  });

  it("asks an unpaired browser for the pairing code, then goes to the workspace once it pairs", async () => {
    const bridge = new FakeBridge();
    renderApp(bridge);
    expect(await screen.findByRole("heading", { name: "Pair this browser" })).toBeInTheDocument();
    const field = screen.getByLabelText("Pairing code");
    fireEvent.change(field, { target: { value: "aaaa-bbbb-cccc" } });
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Pair" }));
      await Promise.resolve();
    });
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
  });
});
