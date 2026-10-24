import { fireEvent, render, screen, within } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it, vi } from "vitest";
import type { Held, TwinSession } from "../../../app/devices";
import { createNavigationStore } from "../../../app/shell";
import { heldIssues } from "../../../model/devices";
import { memoryLocation } from "../../../test/location";
import { kindOf } from "../../../test/twinFixture";
import { HeldNotice } from "./HeldNotice";

const kind = kindOf({
  kind: "indicator",
  metrics: [{ name: "level_percent", data_type: "number", unit: "%", scan_class: "fast" }],
  commands: [],
});
const unsupported = {
  pointer: "/metrics/0/scan_class",
  message: "is not part of the interface a twin reproduces",
  keyword: "twin_member_unsupported",
  member: "scan_class",
};
const path = "kinds/indicator.kind.json";
const held: Held = {
  path,
  name: "indicator",
  words: "indicator is not ready for a twin",
  issues: heldIssues(kind, [unsupported]),
  adding: false,
};

function setup(entry: Held, lead: string) {
  const reveal = vi.fn();
  const twin = { retry: vi.fn() } as unknown as TwinSession;
  const navigation = createNavigationStore(memoryLocation());
  const view = render(<HeldNotice held={[entry]} twin={twin} navigation={navigation} lead={lead} reveal={reveal} />);
  return { reveal, view };
}

describe("a device that blocks or is left out of the twin", () => {
  it("says a member of the twin blocks Create until it is fixed, and names the setting", () => {
    setup(held, "");
    expect(screen.getByText(/cannot be created until it is fixed or left out/)).toBeInTheDocument();
    expect(
      screen.getByText("The setting scan_class of reading 'level_percent' is not reproduced by a twin."),
    ).toBeInTheDocument();
  });

  it("leads from the issue to the exact field when the device is the one being edited, by keyboard too", async () => {
    const { reveal, view } = setup(held, path);
    const show = within(view.container).getByRole("button", { name: "Show scan_class in the Interface step" });
    show.focus();
    expect(show).toHaveFocus();
    fireEvent.click(show);
    expect(reveal).toHaveBeenCalledWith({ part: "metric", key: "m1" }, "scan_class");
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("offers no jump to a field of another device, only the device itself", () => {
    setup({ ...held, adding: true }, "kinds/cooler.kind.json");
    expect(screen.queryByRole("button", { name: /Show scan_class/ })).toBeNull();
    expect(screen.getByRole("button", { name: "Open indicator" })).toBeInTheDocument();
    expect(screen.getByText(/^indicator is left out of the twin as a whole/)).toBeInTheDocument();
  });

  it("keeps the binary's own words under Details, closed at first", () => {
    setup(held, "");
    const details = screen.getByText("Details").closest("details");
    expect(details).not.toHaveAttribute("open");
    expect(
      within(details as HTMLElement).getByText("is not part of the interface a twin reproduces"),
    ).toBeInTheDocument();
  });
});
