import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { ConfirmButton } from "./ConfirmButton";

function button(over: Partial<Parameters<typeof ConfirmButton>[0]> = {}) {
  const onConfirm = vi.fn();
  render(
    <ConfirmButton
      words="Revert"
      icon={null}
      title="Discard the changes"
      question="Discard your changes?"
      confirmWords="Discard them"
      needsConfirm={false}
      onConfirm={onConfirm}
      {...over}
    />,
  );
  return onConfirm;
}

describe("a button that asks before it throws work away", () => {
  it("acts at once when there is nothing to lose", () => {
    const onConfirm = button();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
    expect(screen.queryByRole("group")).toBeNull();
  });

  it("asks first when there is, acts on a yes and does not on a no", () => {
    const onConfirm = button({ needsConfirm: true });
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(onConfirm).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Keep editing" }));
    expect(onConfirm).not.toHaveBeenCalled();
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    fireEvent.click(screen.getByRole("button", { name: "Discard them" }));
    expect(onConfirm).toHaveBeenCalledTimes(1);
  });

  it("does nothing while it is disabled", () => {
    const onConfirm = button({ disabled: true, needsConfirm: true });
    fireEvent.click(screen.getByRole("button", { name: "Revert" }));
    expect(onConfirm).not.toHaveBeenCalled();
    expect(screen.queryByRole("group")).toBeNull();
    expect(screen.getByRole("button", { name: "Revert" })).toHaveAttribute("aria-disabled", "true");
  });
});
