import { act, fireEvent, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { failed } from "../../model/bridge";
import { failure } from "../../test/fakeBridge";
import { editing } from "../../test/editorHarness";

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
});
