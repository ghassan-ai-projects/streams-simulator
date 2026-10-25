import type { Workspace } from "../model/workspace";

export const sampleWorkspace: Workspace = {
  name: "lab-workspace",
  domains: [
    {
      path: "domains/factory-utilities.domain.json",
      id: "factory-utilities",
      title: "Factory utilities",
      instances: 0,
    },
  ],
  compositions: [
    {
      path: "compositions/factory-utilities.composition.json",
      id: "factory-utilities",
      title: "Factory utilities",
      instances: 0,
    },
  ],
  scenarios: [],
  kinds: [],
  components: [],
  runs: [
    {
      worldId: "w-1",
      ended: true,
      complete: true,
      running: false,
      incompleteReason: "",
      explore: false,
      seed: 7,
      plant: "factory-utilities",
      durationS: 1800,
      recordedAt: "2026-10-14T10:00:00Z",
      faults: 0,
    },
    {
      worldId: "w-2",
      ended: false,
      complete: false,
      running: false,
      incompleteReason: "",
      explore: true,
      seed: null,
      plant: "factory-utilities",
      durationS: 0,
      recordedAt: "2026-10-14T11:00:00Z",
      faults: 0,
    },
  ],
};
