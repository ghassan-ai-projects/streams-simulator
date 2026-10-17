import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { useScreen } from "./useScreen";

function Screen({ title }: { readonly title: string }) {
  const heading = useScreen(title);
  return (
    <>
      <dialog open aria-label="Help">
        <button type="button">Close</button>
      </dialog>
      <h1 ref={heading} tabIndex={-1}>
        {title}
      </h1>
    </>
  );
}

describe("a screen heading", () => {
  it("takes focus when a screen arrives", () => {
    render(<Screen title="Plant" />);
    expect(screen.getByRole("heading", { name: "Plant" })).toHaveFocus();
    expect(document.title).toBe("Plant · Streams Simulator");
  });

  it("leaves focus in an open dialog when the screen arrives late", () => {
    const { rerender } = render(<Screen title="Plant" />);
    screen.getByRole("button", { name: "Close" }).focus();
    rerender(<Screen title="Factory utilities" />);
    expect(screen.getByRole("button", { name: "Close" })).toHaveFocus();
    expect(document.title).toBe("Factory utilities · Streams Simulator");
  });
});
