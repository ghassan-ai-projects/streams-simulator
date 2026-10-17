import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { createSession } from "../../../app/session";
import { EditorStore } from "../../../app/editor";
import { PanelStore } from "../../../app/shell";
import { LaunchStore } from "../../../app/launch";
import { PlantMaker } from "../../../app/plant";
import { NavigationStore } from "../../../app/shell";
import { PlantStore } from "../../../app/plant";
import { ReviewStore } from "../../../app/review";
import { ThemeStore } from "../../../app/shell";
import { WorkspaceStore } from "../../../app/workspace";
import { failure as failed } from "../../../test/outcome";
import { FakeBridge, failure } from "../../../test/fakeBridge";
import { fakeClock } from "../../../test/fakeClock";
import { App } from "./App";

function themes() {
  let value: string | null = null;
  return new ThemeStore(
    { read: () => value, write: (_k, v) => (value = v) },
    { prefersDark: () => false, subscribe: () => () => undefined },
  );
}

function renderApp(bridge: FakeBridge) {
  const location = { read: () => "#/", write: () => undefined, subscribe: () => () => undefined };
  const timers = fakeClock();
  const stores = {
    themes: themes(),
    session: createSession(bridge.session),
    workspace: new WorkspaceStore(bridge.workspace),
    navigation: new NavigationStore(location),
    plants: new PlantStore(bridge.plant),
    review: new ReviewStore(bridge.review, timers, timers),
    launch: new LaunchStore(bridge.launch, timers, timers),
    scheduler: timers,
    editor: new EditorStore(bridge.editor),
    panels: new PanelStore({ read: () => null, write: () => undefined }),
    maker: new PlantMaker(bridge.plant),
    showDetails: () => undefined,
  };
  return render(<App stores={stores} />);
}

describe("App", () => {
  it("asks an unpaired browser for the pairing code", async () => {
    renderApp(new FakeBridge());
    expect(await screen.findByRole("heading", { name: "Pair this browser" })).toBeInTheDocument();
  });

  it("opens the workspace once paired, with no accessibility violation", async () => {
    const bridge = new FakeBridge();
    bridge.paired = true;
    const { container } = renderApp(bridge);
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
    expect((await axe.run(container)).violations).toEqual([]);
  });

  it("explains an unreachable bridge", async () => {
    const bridge = new FakeBridge();
    bridge.sessionOutcome = failed(failure("unreachable", 0, "The bridge did not answer."));
    renderApp(bridge);
    expect(await screen.findByRole("heading", { name: "The bridge is not reachable" })).toBeInTheDocument();
    expect(screen.getByText(/The bridge did not answer\./)).toBeInTheDocument();
  });

  it("cycles the theme from the top bar and marks the document", async () => {
    const bridge = new FakeBridge();
    bridge.paired = true;
    renderApp(bridge);
    await screen.findByRole("heading", { name: "lab-workspace" });
    fireEvent.click(screen.getByRole("button", { name: "Theme: system" }));
    expect(screen.getByRole("button", { name: "Theme: light" })).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe("light");
  });

  it("goes from pairing to the workspace", async () => {
    const bridge = new FakeBridge();
    renderApp(bridge);
    const field = await screen.findByLabelText("Pairing code");
    fireEvent.change(field, { target: { value: "aaaa-bbbb-cccc" } });
    await act(async () => {
      fireEvent.click(screen.getByRole("button", { name: "Pair" }));
      await Promise.resolve();
    });
    expect(await screen.findByRole("heading", { name: "lab-workspace" })).toBeInTheDocument();
  });
});
