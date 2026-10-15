import { act, fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { SessionStore } from "../../app/sessionStore";
import { EditorStore } from "../../app/editorStore";
import { PanelStore } from "../../app/panelStore";
import { PlantMaker } from "../../app/plantMaker";
import { LaunchStore } from "../../app/launchStore";
import { NavigationStore } from "../../app/navigationStore";
import { PlantStore } from "../../app/plantStore";
import { ReviewStore } from "../../app/reviewStore";
import { ThemeStore } from "../../app/themeStore";
import { WorkspaceStore } from "../../app/workspaceStore";
import { failed } from "../../model/bridge";
import { FakeBridge, failure } from "../../test/fakeBridge";
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
  const stores = {
    themes: themes(),
    session: new SessionStore(bridge),
    workspace: new WorkspaceStore(bridge),
    navigation: new NavigationStore(location),
    plants: new PlantStore(bridge),
    review: new ReviewStore(bridge),
    launch: new LaunchStore(bridge),
    editor: new EditorStore(bridge),
    panels: new PanelStore({ read: () => null, write: () => undefined }),
    maker: new PlantMaker(bridge),
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
