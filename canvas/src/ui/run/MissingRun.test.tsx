import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { navigation } from "../../test/homeHarness";
import { MissingRun } from "./MissingRun";

describe("a run that is not in the workspace", () => {
  it("names the run and gives a way back to the workspace", () => {
    const nav = navigation();
    render(<MissingRun worldId="w-99" navigation={nav.store} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Run w-99 is not in this workspace.");
    fireEvent.click(screen.getByRole("button", { name: "Back to the workspace" }));
    expect(nav.store.snapshot()).toEqual({ screen: "home" });
  });
});
