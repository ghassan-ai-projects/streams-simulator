import { render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { reviewState as state } from "../../../test/reviewFixture";
import { Notices } from "./Notices";

describe("run notices", () => {
  it("shows nothing when the run has nothing to say", () => {
    render(<Notices state={state({})} />);
    expect(screen.queryByRole("alert")).toBeNull();
    expect(screen.queryByRole("status")).toBeNull();
  });

  it("says when the plant file changed after the run, as information", async () => {
    const changed = state({
      verified: { phase: "ready", value: { matches: true, domainChanged: true, detail: "", reason: "", file: "" } },
    });
    const view = render(<Notices state={changed} />);
    expect(screen.getByRole("status")).toHaveTextContent("The plant file changed after this run");
    expect(screen.getByRole("status")).toHaveAttribute("title", expect.stringContaining("plant file changed"));
    expect(screen.queryByRole("alert")).toBeNull();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("warns in plain words when verify could not run, with the binary's message in the title", () => {
    const failed = state({
      verified: {
        phase: "failed",
        failure: { kind: "unreadable", diagnostic: { message: "run.json missing", displayCode: "verify_unreadable" } },
      },
    });
    render(<Notices state={failed} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Not checked: this run could not be replayed for its verdict.");
    expect(screen.getByRole("alert")).toHaveAttribute("title", expect.stringContaining("run.json missing"));
    expect(screen.getByText("run.json missing")).toBeInTheDocument();
  });

  it("warns when playback holds fewer observations than the report delivered", () => {
    render(<Notices state={state({ playback: { phase: "ready", records: 1 } })} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Playback has 1 observations");
  });
});
