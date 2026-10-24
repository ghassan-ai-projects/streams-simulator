import { screen } from "@testing-library/react";
import { describe, expect, it } from "vitest";
import { mountApp } from "../../../test/appHarness";
import { stubResizeObserver } from "../../../test/dom";
import { twinBridge } from "../../../test/twinFixture";
import { hash } from "../../../test/twinHarness";

describe("a device opened from a link to its twin", () => {
  it("opens on the Behaviour step when the route asks for it", async () => {
    stubResizeObserver();
    mountApp(twinBridge(), `${hash}&step=behaviour`);
    expect(await screen.findByRole("region", { name: "Behaviour of cooler" })).toBeInTheDocument();
  });

  it("opens on the Interface step otherwise", async () => {
    mountApp(twinBridge(), hash);
    expect(await screen.findByRole("region", { name: "Editing cooler" })).toBeInTheDocument();
  });
});
