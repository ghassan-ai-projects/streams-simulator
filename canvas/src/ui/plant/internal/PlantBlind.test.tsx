import { screen } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { stubFlowDom } from "../../../test/dom";
import { attributeValues, canary, namedBy, roles, runningPlant } from "../../../test/runningPlant";

beforeEach(() => {
  stubFlowDom();
});

describe("the plant while a run is going", () => {
  it("shows no planned fault anywhere on the page in Evaluate", async () => {
    await runningPlant(false);
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(roles.map(namedBy)).toEqual(roles.map(() => 0));
    expect(screen.getByRole("group", { name: /^CH-1, chiller/ })).toBeInTheDocument();
    expect(document.body.textContent).not.toContain("Fault ·");
  });

  it("keeps the operator's own effector call visible in Evaluate", async () => {
    await runningPlant(false);
    expect(await screen.findByRole("group", { name: /^CT-1, .*effector call: start · accepted/ })).toBeInTheDocument();
  });

  it("shows the planned fault in Explore", async () => {
    await runningPlant(true);
    expect(document.body.textContent).toMatch(canary);
    expect(attributeValues().some((value) => canary.test(value))).toBe(true);
    expect(namedBy("group")).toBe(1);
  });
});
