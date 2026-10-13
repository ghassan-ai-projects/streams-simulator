import { fireEvent, render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { ThemeStore } from "../app/themeStore";
import { App } from "./App";

function store() {
  let value: string | null = null;
  return new ThemeStore(
    { read: () => value, write: (_k, v) => (value = v) },
    { prefersDark: () => false, subscribe: () => () => undefined },
  );
}

describe("App shell", () => {
  it("shows the workspace sections, the empty state and no accessibility violation", async () => {
    const { container } = render(<App themes={store()} />);
    expect(screen.getByRole("heading", { name: "No workspace connected" })).toBeInTheDocument();
    expect(screen.getByRole("navigation", { name: "Workspace contents" })).toBeInTheDocument();
    const results = await axe.run(container);
    expect(results.violations).toEqual([]);
  });

  it("cycles the theme from the top bar and marks the document", () => {
    render(<App themes={store()} />);
    fireEvent.click(screen.getByRole("button", { name: "Theme: follow the system" }));
    expect(screen.getByRole("button", { name: "Theme: light" })).toBeInTheDocument();
    expect(document.documentElement.dataset.theme).toBe("light");
  });
});
