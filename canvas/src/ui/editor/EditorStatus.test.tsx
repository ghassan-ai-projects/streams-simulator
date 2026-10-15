import { act, fireEvent, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { failed } from "../../model/bridge";
import { failure } from "../../test/fakeBridge";
import { current, editing } from "../../test/editorHarness";

describe("plant editor status", () => {
  it("says which list holds the issues and shows it in Build with one click", async () => {
    const { fake, editor } = await editing();
    act(() => {
      editor.select("CH-1");
    });
    act(() => {
      editor.edit((doc) => ({ ...doc, title: "Lab two" }));
    });
    fake.actions.set(
      "domain.compose",
      failed({
        ...failure("composition_invalid", 422, "invalid"),
        issues: [{ pointer: "/connections/0/to", message: 'instance "CT-1" has no port "heat_in"' }],
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    expect(
      await screen.findByText("Needs fixing: 1 issue, listed in Build. Your changes are saved."),
    ).toBeInTheDocument();
    fireEvent.click(screen.getByRole("button", { name: "Show issues" }));
    expect(screen.getByRole("heading", { name: "Build" })).toHaveFocus();
    expect(screen.getByRole("region", { name: "Issues from the binary" })).toHaveTextContent('has no port "heat_in"');
  });

  it("offers one remedy for a profile weight the plant no longer declares, and removes only that weight", async () => {
    const { fake, editor } = await editing();
    act(() => {
      editor.edit((doc) => ({
        ...doc,
        profiles: [{ name: "sensor_pathology", fault_weights: { gone__fault: 1, kept__fault: 2 } }],
      }));
    });
    fake.actions.set(
      "domain.compose",
      failed({
        ...failure("composition_invalid", 422, "invalid"),
        issues: [
          {
            pointer: "/profiles/0/fault_weights/gone__fault",
            message: 'profile "sensor_pathology" weights undeclared fault "gone__fault"',
          },
        ],
      }),
    );
    fireEvent.click(screen.getByRole("button", { name: "Save" }));
    fireEvent.click(await screen.findByRole("button", { name: "Remove this weight" }));
    expect(screen.getByText(/has a weight for gone__fault, which no instance declares/)).toBeInTheDocument();
    expect(current(editor).profiles).toEqual([{ name: "sensor_pathology", fault_weights: { kept__fault: 2 } }]);
  });
});
