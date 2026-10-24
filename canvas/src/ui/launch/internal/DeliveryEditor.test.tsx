import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { cannedFailure } from "../../../test/fakeBridge";
import { launchBridge } from "../../../test/launchFixture";
import { failure as failed } from "../../../test/outcome";
import { change, press, renderPanel, start } from "../../../test/planEditorHarness";

describe("delivery conditions", () => {
  it("offers what the simulator lists, describes the choice and types its settings from the catalog", async () => {
    const { fake } = await renderPanel();
    press("Add delivery condition");
    expect(screen.getByLabelText("Delivery condition 1 kind")).toHaveFocus();
    expect(screen.getAllByRole("option").map((option) => option.textContent)).toEqual(
      expect.arrayContaining(["drop", "producer flap", "reorder"]),
    );
    change("Delivery condition 1 kind", "drop");
    expect(screen.getByText(/Loses records in transport/)).toBeInTheDocument();
    const rate = screen.getByLabelText("Delivery condition 1 rate");
    expect(rate).toHaveAttribute("placeholder", "0 to 1");
    change("Delivery condition 1 rate", "0.25");
    change("Delivery condition 1 starts at minute", "3");
    change("Delivery condition 1 ends at minute", "9");
    const args = await start(fake);
    expect(args?.perturbations).toEqual([{ kind: "drop", params: { rate: 0.25 }, at_s: 180, until_s: 540 }]);
  });

  it("renders an enumerated setting as a choice and asks for an end where the kind needs one", async () => {
    const { fake } = await renderPanel();
    press("Add delivery condition");
    change("Delivery condition 1 kind", "producer_flap");
    expect(screen.getByLabelText("Delivery condition 1 ends at minute")).toHaveAttribute("placeholder", "needed");
    expect(screen.getByText(/It needs an end minute\./)).toBeInTheDocument();
    const mode = screen.getByLabelText("Delivery condition 1 mode");
    expect(
      within(mode)
        .getAllByRole("option")
        .map((option) => option.textContent),
    ).toEqual(["default", "hold", "lose"]);
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    press("Start the run");
    const end = screen.getByLabelText("Delivery condition 1 ends at minute");
    expect(end).toHaveFocus();
    expect(end).toHaveAccessibleDescription(/^Delivery condition 1 needs an end: Silences the producer/);
    expect(fake.jobCalls).toEqual([]);
    change("Delivery condition 1 ends at minute", "5");
    change("Delivery condition 1 mode", "hold");
    change("Delivery condition 1 burst", "2");
    const args = await start(fake);
    expect(args?.perturbations).toEqual([
      { kind: "producer_flap", params: { mode: "hold", burst: 2 }, at_s: 0, until_s: 300 },
    ]);
  });

  it("names a setting outside the range the catalog declares and focuses it", async () => {
    await renderPanel();
    press("Add delivery condition");
    change("Delivery condition 1 kind", "drop");
    change("Delivery condition 1 rate", "1.5");
    fireEvent.click(screen.getByRole("radio", { name: /As fast as possible/ }));
    press("Start the run");
    const rate = screen.getByLabelText("Delivery condition 1 rate");
    expect(rate).toHaveFocus();
    expect(rate).toHaveAccessibleDescription("Delivery condition 1: rate is a number from 0 to 1.");
  });

  it("clears the settings when the kind changes, and removes a row with the focus returning to Add", async () => {
    await renderPanel();
    press("Add delivery condition");
    change("Delivery condition 1 kind", "drop");
    change("Delivery condition 1 rate", "0.5");
    change("Delivery condition 1 kind", "reorder");
    expect(screen.queryByLabelText("Delivery condition 1 rate")).toBeNull();
    press("Remove delivery condition 1");
    expect(screen.getByRole("button", { name: "Add delivery condition" })).toHaveFocus();
    expect(screen.queryByLabelText("Delivery condition 1 kind")).toBeNull();
  });

  it("says when the list could not be read and offers no way to add", async () => {
    const fake = launchBridge();
    fake.actions.set("catalog.perturbations", failed(cannedFailure("boom", 500, "the director is away")));
    await renderPanel(fake);
    expect(screen.getByRole("alert")).toHaveTextContent(
      "The delivery conditions could not be read: the director is away",
    );
    expect(screen.getByRole("button", { name: "Add delivery condition" })).toBeDisabled();
  });
});
