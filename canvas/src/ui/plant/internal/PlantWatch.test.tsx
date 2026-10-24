import { act, fireEvent, screen, within } from "@testing-library/react";
import { beforeEach, describe, expect, it } from "vitest";
import { headerHeight } from "../../../model/graph";
import { stubFlowDom } from "../../../test/dom";
import { elapse } from "../../../test/launchFixture";
import { attributeValues, canary, namedBy, roles, runningPlant } from "../../../test/runningPlant";
import type { FakeBridge } from "../../../test/fakeBridge";
import { chillerPorts, componentShown } from "../../../test/plantFixture";
import { success } from "../../../test/outcome";
import { observation } from "../../../test/runFixture";

beforeEach(() => {
  stubFlowDom();
});

const startNs = 1_767_225_600e9;
const at = (second: number) => startNs + second * 1e9;

const opened = {
  stream: "observer",
  type: "run.opened",
  seq: 0,
  at_ns: startNs,
  mode: "explore",
  adapter: { id: "a" },
};
const clock = (second: number) => ({
  stream: "observer",
  type: "clock.advanced",
  seq: 99,
  at_ns: at(second),
  to_ns: at(second),
});

function reading(second: number, value: number, extra: object = {}) {
  const wire = observation(second, at(second), "CT-1", "tower.basin_temp", value, "degC");
  return { ...wire, line: JSON.stringify({ ...JSON.parse(wire.line), ...extra }) };
}

function mapBasinTemp(bridge: FakeBridge) {
  const basin = {
    name: "basin_out",
    type: "cooling_water_temp",
    direction: "out",
    channel: "basin_temp",
  } as const;
  bridge.handlers.set("component.show", ({ ref }) =>
    success(componentShown(ref === "lib/tower" ? { ports: [basin] } : chillerPorts)),
  );
}

async function watching(explore: boolean, ...entries: object[]) {
  const app = await runningPlant(explore, mapBasinTemp);
  fireEvent.click(screen.getByRole("button", { name: "Close the run panel" }));
  const outline = screen.getByRole("navigation", { name: "Outline" });
  fireEvent.click(within(outline).getByRole("button", { name: /CT-1/ }));
  await feed(app, ...entries);
  return app;
}

async function feed(app: Awaited<ReturnType<typeof runningPlant>>, ...entries: object[]) {
  await act(async () => {
    app.bridge.stream("w-7", entries, null);
    await elapse(app.timers, 500);
  });
}

const inspector = () => screen.getByRole("complementary", { name: "Inspector" });
const observed = () => within(inspector()).getByRole("region", { name: "Observed readings" });

const steady = [opened, ...[10, 20, 30, 40, 50, 60].map((second) => reading(second, 26 + second / 100)), clock(75)];

describe("the inspector while a run is going", () => {
  it("shows the value with its unit symbol, when it was observed and how old it is", async () => {
    await watching(true, ...steady);
    const row = within(observed()).getByRole("row", { name: /basin temp/ });
    expect(row).toHaveTextContent("26.6");
    expect(row).toHaveTextContent("°C");
    expect(within(row).getByText("00:01:00")).toHaveAttribute("data-part", "observed-at");
    expect(within(row).getByText("15 s old")).toBeInTheDocument();
    expect(within(row).getByText("°C")).toHaveAttribute("title", "unit code: degC");
  });

  it("draws a small trend for the channel, described in words with how its gaps are inferred", async () => {
    await watching(true, ...steady);
    const trend = within(observed()).getByRole("img", { name: /basin temp trend/ });
    expect(trend).toHaveAttribute("data-part", "trend");
    expect(trend).toHaveAccessibleName(/Last 1 min of simulated time, from 26.1 to 26.6 °C\. No gaps\./);
    expect(trend.querySelectorAll("polyline")).toHaveLength(1);
  });

  it("breaks the trend where the channel went quiet and says how it knows", async () => {
    const gap = [10, 20, 30, 40, 50, 300, 310, 320].map((second) => reading(second, 20 + second / 100));
    await watching(true, opened, ...gap, clock(330));
    const trend = within(observed()).getByRole("img", { name: /basin temp trend/ });
    expect(trend.querySelectorAll("polyline")).toHaveLength(2);
    expect(trend).toHaveAccessibleName(/1 gap left blank.*inferred from the observed intervals/);
  });

  it("says a channel that has not reported yet is awaiting its first reading, and claims no cause", async () => {
    await watching(true, opened, clock(5));
    const row = within(observed()).getByRole("row", { name: /basin temp/ });
    expect(row).toHaveTextContent("Awaiting first reading");
    expect(observed().textContent).not.toMatch(/broken|failed|offline|sensor|fault/i);
  });

  it("names the quality the instrument gave a reading in plain words", async () => {
    await watching(true, opened, reading(10, 26, { quality: "out_of_range" }), clock(12));
    const row = within(observed()).getByRole("row", { name: /basin temp/ });
    expect(within(row).getByText("out of range")).toHaveAttribute("data-part", "quality");
  });

  it("measures the age from the run clock, so a run that paused does not look stale", async () => {
    const app = await watching(true, opened, reading(10, 26), clock(10));
    await feed(app);
    expect(within(observed()).getByText("just now")).toBeInTheDocument();
    await feed(app, clock(190));
    expect(within(observed()).getByText("3 min old")).toBeInTheDocument();
  });

  it("puts the same reading on the graph node", async () => {
    await watching(true, ...steady);
    const node = screen.getByRole("group", { name: /^CT-1, tower/ });
    expect(node).toHaveTextContent("26.6");
    expect(node).toHaveTextContent("°C");
  });

  it("keeps the model's channel list below the readings, folded", async () => {
    await watching(true, ...steady);
    const channels = within(inspector()).getByRole("region", { name: "Channels" });
    expect(channels.querySelector("details")).not.toHaveAttribute("open");
    const html = inspector().innerHTML;
    expect(html.indexOf("Observed now")).toBeLessThan(html.indexOf("Links in"));
  });

  it("says nothing was observed once the journal has closed on a channel that never reported", async () => {
    const app = await watching(true, opened, clock(5));
    await act(async () => {
      app.bridge.stream("w-7", [{ stream: "observer", type: "run.closed", seq: 100, at_ns: at(6), incomplete: false }]);
      await elapse(app.timers, 500);
    });
    expect(within(observed()).getByText("Never observed in this run")).toBeInTheDocument();
  });

  it("stops following the run when it is over", async () => {
    const app = await watching(true, ...steady);
    expect(app.bridge.openFollows("w-7")).toBe(1);
    await act(async () => {
      await elapse(app.timers, 40_000);
    });
    expect(app.bridge.openFollows("w-7")).toBe(0);
  });
});

describe("the same inspector in Evaluate", () => {
  it("reads only the observer journal: no report, verification or file from the run", async () => {
    const app = await watching(false, ...steady);
    expect(app.bridge.calls.map((call) => call.name).filter((name) => name.startsWith("run."))).toEqual([]);
    expect(app.bridge.reads.filter((path) => path.startsWith("runs/"))).toEqual([]);
  });

  it("shows the readings and still no planned fault anywhere, in text, attributes, names or trend titles", async () => {
    await watching(false, ...steady);
    expect(observed()).toHaveTextContent("26.6");
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(roles.map(namedBy)).toEqual(roles.map(() => 0));
    expect(document.body.innerHTML).not.toMatch(canary);
    expect(document.body.textContent).not.toContain("Fault ·");
  });
});

const acknowledged = (second: number, accepted: boolean) => ({
  stream: "observer",
  type: "effector.acknowledged",
  seq: 50 + second,
  at_ns: at(second),
  command_id: `c-${String(second)}`,
  effector: "set_fan_speed",
  entity_id: "CT-1",
  response: { accepted, simulated: true, world_id: "w-7", command_id: `c-${String(second)}` },
});

describe("the trends panel in Evaluate", () => {
  it("shows the operator's own effector calls on the shared axis and nothing of the planned fault anywhere", async () => {
    await watching(false, ...steady, acknowledged(30, true), acknowledged(45, false));
    const inspectorNow = within(inspector());
    fireEvent.click(inspectorNow.getByRole("button", { name: "Watch basin temp" }));
    fireEvent.click(screen.getByRole("button", { name: /^Trends/ }));
    const panel = screen.getByRole("region", { name: "Trends" });
    fireEvent.click(within(panel).getByText(/^Events in this window/));
    expect(panel).toHaveTextContent("Effector call · CT-1 set fan speed · accepted");
    expect(panel).toHaveTextContent("Effector call · CT-1 set fan speed · refused");
    expect(panel.querySelectorAll("[data-kind='fault']")).toHaveLength(0);
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(roles.map(namedBy)).toEqual(roles.map(() => 0));
    expect(document.body.innerHTML).not.toMatch(canary);
    expect(document.body.textContent).not.toMatch(/Fault starts|Fault ·/);
  });
});

describe("the trends panel in Explore", () => {
  it("marks the planned faults the run clock has reached, and the operator's calls, on the shared axis", async () => {
    await watching(true, ...steady, acknowledged(30, true));
    fireEvent.click(within(inspector()).getByRole("button", { name: "Watch basin temp" }));
    fireEvent.click(screen.getByRole("button", { name: /^Trends/ }));
    const panel = screen.getByRole("region", { name: "Trends" });
    fireEvent.click(within(panel).getByText(/^Events in this window/));
    expect(panel).toHaveTextContent(/Fault starts · CH-1 /);
    expect(panel).toHaveTextContent("Effector call · CT-1 set fan speed · accepted");
    expect(panel.querySelectorAll("[data-kind='fault']")).toHaveLength(1);
  });
});

describe("a fault injected during the run", () => {
  it("is marked on its node and on the trends overlay from the instant it was injected in Explore", async () => {
    const app = await watching(true, ...steady, clock(130));
    await act(async () => {
      await app.stores.launch.injectNow("CT-1", "tower__scaling");
    });
    expect(screen.getByRole("group", { name: /^CT-1, Fault: scaling/ })).toBeInTheDocument();
    fireEvent.click(within(inspector()).getByRole("button", { name: "Watch basin temp" }));
    fireEvent.click(screen.getByRole("button", { name: /^Trends/ }));
    const panel = screen.getByRole("region", { name: "Trends" });
    fireEvent.click(within(panel).getByText(/^Events in this window/));
    expect(panel).toHaveTextContent(/Fault starts · CT-1 /);
    expect(panel.querySelectorAll("[data-kind='fault']")).toHaveLength(2);
  });

  it("is refused in Evaluate: nothing is sent, no fault appears, and a call given now stays visible", async () => {
    const app = await watching(false, ...steady);
    await act(async () => {
      await app.stores.launch.injectNow("CT-1", "chiller__CANARY_TRUTH_9f3a");
      await app.stores.launch.invokeNow("CT-1", "start", {});
    });
    expect(app.bridge.calls.map((call) => call.name)).not.toContain("job.inject_fault");
    expect(await screen.findByRole("group", { name: /^CT-1, .*effector call: start · accepted/ })).toBeInTheDocument();
    expect(document.body.textContent).not.toMatch(canary);
    expect(attributeValues().filter((value) => canary.test(value))).toEqual([]);
    expect(roles.map(namedBy)).toEqual(roles.map(() => 0));
    expect(document.body.innerHTML).not.toMatch(canary);
  });
});

describe("the room a live node has for its readout", () => {
  it("holds the header, the note line and every readout row it draws", async () => {
    await watching(true, ...steady);
    const node = document.querySelector<HTMLElement>('.react-flow__node[data-id="CT-1"]');
    const rows = node?.querySelectorAll("dl > div").length ?? 0;
    const height = Number.parseFloat(node?.style.height ?? "0");
    expect(rows).toBeGreaterThan(0);
    expect(height).toBeGreaterThanOrEqual(headerHeight("CT-1") + 16 + 24 + rows * 24);
  });
});
