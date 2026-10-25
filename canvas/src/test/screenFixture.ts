import { success } from "../app/ports";
import { FakeBridge } from "./fakeBridge";
import { campus, chillerPorts, componentShown, composition, domain, towerPorts } from "./plantFixture";
import { journalWire, reportWire } from "./runFixture";
import { sampleWorkspace } from "./workspaceFixture";

const miniDomain = "domains/mini.domain.json";
const miniComposition = "compositions/mini.composition.json";

export const miniWorkspace = {
  ...sampleWorkspace,
  domains: [{ path: miniDomain, id: "mini", title: "Mini plant", instances: 3 }],
  compositions: [{ path: miniComposition, id: "mini", title: "Mini plant", instances: 3 }],
  runs: sampleWorkspace.runs.map((run) => ({ ...run, plant: "mini" })),
};

export function miniBridge() {
  const fake = new FakeBridge();
  fake.sessionOutcome = success(true);
  fake.workspaceOutcome = success(miniWorkspace);
  fake.addFile(miniDomain, domain);
  fake.addFile(miniComposition, composition);
  fake.handlers.set("component.show", ({ ref }) =>
    success(componentShown(ref === "lib/tower" ? towerPorts : chillerPorts)),
  );
  return fake;
}

export function reviewBridge() {
  const fake = miniBridge();
  fake.actions.set("run.report", success(reportWire));
  fake.actions.set("run.verify", success({ matches: true }));
  fake.addFile("runs/w-1/run.json", { domain_spec: domain });
  return fake;
}

export function streamJournal(fake: FakeBridge, runId = "w-1") {
  fake.stream(runId, journalWire);
}

export const planHash = (hall: string | null = null) =>
  `#/plant?domain=${encodeURIComponent(miniDomain)}${hall === null ? "" : `&hall=${hall}`}`;

const campusDomain = "domains/campus.domain.json";

export function campusBridge(halls = 3) {
  const fake = miniBridge();
  fake.workspaceOutcome = success({
    ...miniWorkspace,
    domains: [
      ...miniWorkspace.domains,
      { path: campusDomain, id: "campus", title: "Campus", instances: halls * 2 + 1 },
    ],
  });
  fake.addFile(campusDomain, campus(halls));
  return fake;
}

export const campusHash = (hall: string | null = null) =>
  `#/plant?domain=${encodeURIComponent(campusDomain)}${hall === null ? "" : `&hall=${hall}`}`;

export const reviewHash = (worldId = "w-1") => `#/review?world=${worldId}`;

export function campusReviewBridge(halls = 3) {
  const fake = reviewBridge();
  fake.addFile("runs/w-1/run.json", { domain_spec: campus(halls) });
  return fake;
}
