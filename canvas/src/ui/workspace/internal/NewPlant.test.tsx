import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { failure as failed, success } from "../../../test/outcome";
import { FakeBridge, failure } from "../../../test/fakeBridge";
import { navigation, show } from "../../../test/homeHarness";

describe("new plant id", () => {
  it("replaces the derived id when typing in Id, and stops deriving it once typed", async () => {
    const bridge = new FakeBridge();
    const base = "compositions/factory-utilities.composition.json";
    const factory = { id: "factory-utilities", equipment: [{ use: "lib/chiller@1", names: ["CH-1"] }] };
    bridge.files.set(base, success({ path: base, content: JSON.stringify(factory), etag: "b" }));
    await show(bridge, navigation());
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    const form = screen.getByRole("form", { name: "New plant" });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Water pumping station" } });
    const id = within(form).getByLabelText<HTMLInputElement>("Id");
    expect(id).toHaveValue("water-pumping-station");
    fireEvent.focus(id);
    expect([id.selectionStart, id.selectionEnd]).toEqual([0, id.value.length]);
    fireEvent.change(id, { target: { value: "pump" } });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Water pumping hall" } });
    expect(id).toHaveValue("pump");
  });

  it("shows the binary's id rule under the Id field on leaving it, and clears the message when the id changes", async () => {
    const bridge = new FakeBridge();
    bridge.actions.set(
      "component.new",
      failed({
        ...failure("composition_invalid", 422, "invalid"),
        issues: [{ pointer: "/id", message: '"Bad Id!" is not a plant id: 3 to 64 lower-case letters' }],
      }),
    );
    await show(bridge, navigation());
    fireEvent.click(screen.getByRole("button", { name: "New plant" }));
    const form = screen.getByRole("form", { name: "New plant" });
    fireEvent.change(within(form).getByLabelText("Name"), { target: { value: "Bad plant" } });
    const id = within(form).getByLabelText("Id");
    fireEvent.change(id, { target: { value: "Bad Id!" } });
    expect(screen.queryByText(/is not a plant id/)).toBeNull();
    fireEvent.blur(id);
    expect(await screen.findByText(/is not a plant id/)).toBeInTheDocument();
    expect(id).toHaveAccessibleDescription(/is not a plant id/);
    expect(screen.getAllByText(/is not a plant id/)).toHaveLength(1);
    expect(id).toHaveAttribute("aria-invalid", "true");
    expect(screen.queryByText(/starting with a letter/)).toBeNull();
    fireEvent.change(id, { target: { value: "bad-plant" } });
    expect(screen.queryByText(/is not a plant id/)).toBeNull();
    expect(id).not.toHaveAttribute("aria-invalid");
    expect(screen.getByText(/starting with a letter/)).toBeInTheDocument();
  });
});
