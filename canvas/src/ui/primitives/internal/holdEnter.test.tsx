import { fireEvent, render, screen } from "@testing-library/react";
import { describe, expect, it, vi } from "vitest";
import { holdEnter } from "./holdEnter";

describe("holding Enter on a form", () => {
  it("stops Enter on a select from submitting and lets every other control submit", () => {
    const submit = vi.fn((event: { preventDefault: () => void }) => {
      event.preventDefault();
    });
    render(
      <form onKeyDown={holdEnter} onSubmit={submit}>
        <select aria-label="Kind">
          <option>a</option>
        </select>
        <input aria-label="Name" />
      </form>,
    );
    expect(fireEvent.keyDown(screen.getByLabelText("Kind"), { key: "Enter" })).toBe(false);
    expect(fireEvent.keyDown(screen.getByLabelText("Name"), { key: "Enter" })).toBe(true);
    expect(fireEvent.keyDown(screen.getByLabelText("Kind"), { key: "ArrowDown" })).toBe(true);
  });
});
