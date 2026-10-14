import { render, screen } from "@testing-library/react";
import axe from "axe-core";
import { describe, expect, it } from "vitest";
import { reviewState as state } from "../../test/reviewFixture";
import { Banners, Status } from "./RunStatus";

describe("run status", () => {
  it("badges a divergent replay Not verified with no warning banner", () => {
    const divergent = state({
      verified: {
        phase: "ready",
        value: { matches: false, domainChanged: false, detail: "trace differs", reason: "", file: "" },
      },
    });
    render(
      <>
        <Status state={divergent} />
        <Banners state={divergent} />
      </>,
    );
    expect(screen.getByRole("status")).toHaveTextContent("Not verified");
    expect(screen.queryByRole("alert")).toBeNull();
  });

  it("says when the plant file changed after the run, as information", async () => {
    const changed = state({
      verified: { phase: "ready", value: { matches: true, domainChanged: true, detail: "", reason: "", file: "" } },
    });
    const view = render(<Banners state={changed} />);
    expect(screen.getByRole("status")).toHaveTextContent("The plant file changed after this run");
    expect(screen.queryByRole("alert")).toBeNull();
    expect((await axe.run(view.container)).violations).toEqual([]);
  });

  it("warns in plain words when verify could not run, with the binary's message under Details", () => {
    const failed = state({
      verified: { phase: "failed", failure: { status: 422, code: "x", message: "run.json missing", issues: [] } },
    });
    render(<Banners state={failed} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Not checked: streamsim verify could not replay this run.");
    expect(screen.getByText("run.json missing")).toBeInTheDocument();
  });

  it("warns when playback holds fewer observations than the report delivered", () => {
    render(<Banners state={state({ playback: { phase: "ready", records: 1, adapter: "native-jsonl" } })} />);
    expect(screen.getByRole("alert")).toHaveTextContent("Playback has 1 observations");
  });
});
