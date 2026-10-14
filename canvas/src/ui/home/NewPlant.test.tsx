import { fireEvent, screen, within } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { success } from "../../model/bridge";
import { FakeBridge } from "../../test/fakeBridge";
import { navigation, show } from "../../test/homeHarness";

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
});
