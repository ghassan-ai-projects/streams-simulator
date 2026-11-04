import { render, screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { StatusLine } from "./StatusLine";

describe("a status line", () => {
  it("shows the text and uses the detail as its title", () => {
    render(<StatusLine status={{ text: "Not saved", detail: "disk full", tone: "warn" }} />);
    const line = screen.getByRole("status");
    expect(line).toHaveTextContent("Not saved");
    expect(line).toHaveAttribute("title", "disk full");
    expect(line).not.toHaveAttribute("id");
  });

  it("falls back to the text as its title and carries an id when given", () => {
    render(<StatusLine id="twin-status" status={{ text: "Ready", detail: null, tone: "ok" }} />);
    const line = screen.getByRole("status");
    expect(line).toHaveAttribute("title", "Ready");
    expect(line).toHaveAttribute("id", "twin-status");
  });
});
