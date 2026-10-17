import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import type { PlantFailure as Failure } from "../../../app/plant";
import { navigation } from "../../../test/workspaceHarness";
import { PlantFailure } from "./PlantFailure";

describe("a plant that cannot be opened", () => {
  it("says the plant is not in the workspace, gives a way back, and keeps the binary's words in details", () => {
    const nav = navigation();
    const missing: Failure = {
      kind: "missing",
      diagnostic: { message: "canvasbridge: file not found", displayCode: "not_found" },
    };
    render(<PlantFailure failure={missing} navigation={nav.store} />);
    expect(screen.getByRole("alert")).toHaveTextContent("This plant is not in the workspace any more.");
    expect(screen.getByText("canvasbridge: file not found (not_found)")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Back to the workspace" }));
    expect(nav.store.snapshot()).toEqual({ screen: "home" });
  });

  it("says the plant could not be opened for any other failure", () => {
    const unreadable: Failure = {
      kind: "unreadable",
      diagnostic: { message: "bad json", displayCode: "bridge_failed" },
    };
    render(<PlantFailure failure={unreadable} navigation={navigation().store} />);
    expect(screen.getByRole("alert")).toHaveTextContent("This plant could not be opened.");
  });
});
