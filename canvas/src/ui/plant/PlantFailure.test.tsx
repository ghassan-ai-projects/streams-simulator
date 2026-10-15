import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { failure } from "../../test/fakeBridge";
import { navigation } from "../../test/homeHarness";
import { PlantFailure } from "./PlantFailure";

describe("a plant that cannot be opened", () => {
  it("says the plant is not in the workspace, gives a way back, and keeps the binary's words in details", () => {
    const nav = navigation();
    render(<PlantFailure failure={failure("not_found", 404, "canvasbridge: file not found")} navigation={nav.store} />);
    expect(screen.getByRole("alert")).toHaveTextContent("This plant is not in the workspace any more.");
    expect(screen.getByText("canvasbridge: file not found (not_found)")).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Back to the workspace" }));
    expect(nav.store.snapshot()).toEqual({ screen: "home" });
  });

  it("says the plant could not be opened for any other failure", () => {
    render(<PlantFailure failure={failure("unreadable", 500, "bad json")} navigation={navigation().store} />);
    expect(screen.getByRole("alert")).toHaveTextContent("This plant could not be opened.");
  });
});
