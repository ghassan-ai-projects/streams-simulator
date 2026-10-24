import { render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import type { ObservedRow } from "../../../model/review";
import { ObservedTable } from "./ObservedTable";

const row = (overrides: Partial<ObservedRow> = {}): ObservedRow => ({
  key: "units_made",
  label: "units made",
  present: true,
  value: "180 013",
  unit: "count",
  unitTitle: "",
  word: false,
  attention: null,
  outside: "",
  observedAt: "00:10:00",
  measuredAt: "",
  age: "now",
  quality: "",
  absent: "",
  trend: null,
  trendNote: "",
  ...overrides,
});

const legend = "A quality word appears only when a reading is suspect, out of range, stale or disconnected.";

describe("the observed readings table", () => {
  it("says on its value header and in a legend line when a quality word appears, and shows none for a good reading", () => {
    const { container } = render(<ObservedTable rows={[row()]} clock="at 00:10:00" />);
    expect(screen.getByRole("columnheader", { name: "Value" })).toHaveAttribute("title", legend);
    expect(container.querySelector("[data-part='quality-legend']")).toHaveTextContent(legend);
    expect(container.querySelector("[data-part='quality']")).toBeNull();
    expect(screen.getByText("180 013")).toBeInTheDocument();
  });

  it("shows the quality word of a suspect reading beside its value, with the same legend", async () => {
    const { container } = render(<ObservedTable rows={[row({ quality: "stale" })]} clock="at 00:10:00" />);
    expect(screen.getByText("stale")).toHaveAttribute("data-part", "quality");
    expect(container.querySelector("[data-part='quality-legend']")).toBeInTheDocument();
    expect((await axe.run(container)).violations).toEqual([]);
  });
});
