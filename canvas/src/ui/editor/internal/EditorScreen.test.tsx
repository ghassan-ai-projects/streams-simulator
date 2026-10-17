import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { failure as failed } from "../../../test/outcome";
import { mountApp } from "../../../test/appHarness";
import { stubFlowDom } from "../../../test/dom";
import { cannedFailure } from "../../../test/fakeBridge";
import { current, labBridge, path } from "../../../test/editorHarness";

beforeEach(() => {
  stubFlowDom();
});

const editHash = (composition = path) => `#/edit?composition=${encodeURIComponent(composition)}`;

async function editor(bridge = labBridge()) {
  const app = mountApp(bridge, editHash());
  await screen.findByRole("region", { name: "Editing Lab" });
  return { ...app, bridge };
}

function rename(app: Awaited<ReturnType<typeof editor>>, title = "Lab two") {
  act(() => {
    app.stores.editor.edit((doc) => ({ ...doc, title }));
  });
}

function shortcut(key: string, init: KeyboardEventInit = {}, target: Element = document.body) {
  return fireEvent.keyDown(target, { key, ctrlKey: true, ...init });
}

describe("the editor screen", () => {
  it("opens the composition the route names, with its equipment drawn and listed", async () => {
    await editor();
    expect(screen.getByRole("heading", { name: "Lab" })).toHaveAttribute("aria-current", "location");
    expect(screen.getByRole("group", { name: /^CH-1/ })).toBeInTheDocument();
    expect(
      within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /CT-1/ }),
    ).toBeInTheDocument();
  });

  it("waits with a status while the composition is read", async () => {
    const bridge = labBridge();
    bridge.holdRead(path);
    mountApp(bridge, editHash());
    expect(await screen.findByText("Opening the composition…")).toBeInTheDocument();
  });

  it("says why a composition cannot be opened", async () => {
    const bridge = labBridge();
    bridge.files.delete(path);
    mountApp(bridge, editHash());
    expect(await screen.findByRole("alert")).toHaveTextContent("The composition could not be opened:");
  });

  it("invites the first piece of equipment when the composition has none", async () => {
    const bridge = labBridge();
    bridge.addFile(path, { id: "lab", title: "Lab", areas: [{ name: "Cooling" }], equipment: [], connections: [] });
    await editor(bridge);
    expect(screen.getByRole("heading", { name: "No equipment yet" })).toBeInTheDocument();
    expect(screen.queryByRole("button", { name: "Zoom in" })).toBeNull();
  });

  it("selects an instance from the outline and opens its editor", async () => {
    await editor();
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /CH-1/ }));
    expect(screen.getByRole("heading", { name: "CH-1" })).toBeInTheDocument();
  });

  it("selects an instance clicked on the graph and clears the choice on the empty pane", async () => {
    const app = await editor();
    fireEvent.click(screen.getByRole("group", { name: /^CT-1/ }));
    expect(app.stores.editor.snapshot()).toMatchObject({ selected: "CT-1" });
    const pane = app.view.container.querySelector(".react-flow__pane");
    if (pane === null) throw new Error("no pane");
    fireEvent.click(pane);
    expect(app.stores.editor.snapshot()).toMatchObject({ selected: null });
  });

  it("warns before the tab closes only while there are unsaved changes", async () => {
    const app = await editor();
    expect(window.dispatchEvent(new Event("beforeunload", { cancelable: true }))).toBe(true);
    rename(app);
    expect(window.dispatchEvent(new Event("beforeunload", { cancelable: true }))).toBe(false);
  });

  it("stops warning once the editor is left", async () => {
    const app = await editor();
    rename(app);
    app.view.unmount();
    expect(window.dispatchEvent(new Event("beforeunload", { cancelable: true }))).toBe(true);
  });

  it("asks before leaving for the workspace with unsaved changes", async () => {
    const app = await editor();
    rename(app);
    fireEvent.click(
      within(screen.getByRole("navigation", { name: "Breadcrumb" })).getByRole("button", { name: "Workspace" }),
    );
    expect(screen.getByRole("heading", { name: "Leave with unsaved changes?" })).toBeInTheDocument();
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "edit" });
  });

  it("leaves a clean editor without asking", async () => {
    const app = await editor();
    fireEvent.click(
      within(screen.getByRole("navigation", { name: "Breadcrumb" })).getByRole("button", { name: "Workspace" }),
    );
    expect(app.stores.navigation.snapshot()).toEqual({ screen: "workspace" });
  });
});

describe("the editor keyboard", () => {
  it("saves with Ctrl+S only when there is something to save", async () => {
    const app = await editor();
    expect(shortcut("s")).toBe(false);
    expect(app.bridge.writes).toHaveLength(0);
    rename(app);
    shortcut("s");
    await screen.findByText("Ready to run. Saved.");
    expect(app.bridge.writes).toHaveLength(1);
  });

  it("undoes with Ctrl+Z, redoes with Ctrl+Shift+Z and Ctrl+Y, and also with the Command key", async () => {
    const app = await editor();
    rename(app);
    shortcut("z");
    expect(current(app.stores.editor).title).toBe("Lab");
    shortcut("z", { shiftKey: true });
    expect(current(app.stores.editor).title).toBe("Lab two");
    shortcut("z", { ctrlKey: false, metaKey: true });
    expect(current(app.stores.editor).title).toBe("Lab");
    shortcut("y");
    expect(current(app.stores.editor).title).toBe("Lab two");
  });

  it("leaves undo to a text field and ignores unrelated or Alt-modified chords", async () => {
    const app = await editor();
    rename(app);
    fireEvent.click(within(screen.getByRole("navigation", { name: "Outline" })).getByRole("button", { name: /CH-1/ }));
    const field = screen.getByLabelText("Name");
    expect(shortcut("z", {}, field)).toBe(true);
    expect(shortcut("q")).toBe(true);
    expect(shortcut("z", { altKey: true })).toBe(true);
    expect(current(app.stores.editor).title).toBe("Lab two");
  });

  it("does nothing for plain letters", async () => {
    const app = await editor();
    rename(app);
    fireEvent.keyDown(document.body, { key: "z" });
    expect(current(app.stores.editor).title).toBe("Lab two");
  });
});

describe("the editor bar", () => {
  it("shows a save conflict and reloads from disk", async () => {
    const app = await editor();
    rename(app);
    app.bridge.writeOutcomes.set(path, failed(cannedFailure("conflict", 409, "changed on disk")));
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    const reload = await screen.findByRole("button", { name: "Reload from disk" });
    app.bridge.writeOutcomes.delete(path);
    fireEvent.click(reload);
    await screen.findByText("No changes");
    expect(current(app.stores.editor).title).toBe("Lab");
  });

  it("opens the composed plant from Open plant", async () => {
    const app = await editor();
    fireEvent.click(await screen.findByRole("button", { name: "Open plant" }));
    expect(app.stores.navigation.snapshot()).toMatchObject({ screen: "plant", hall: null });
  });
});
