import type { Workspace } from "../model/workspace";

export const sampleWorkspace: Workspace = {
  name: "lab-workspace",
  domains: [{ path: "domains/factory-utilities.domain.json", id: "factory-utilities", title: "Factory utilities" }],
  compositions: [
    { path: "compositions/factory-utilities.composition.json", id: "factory-utilities", title: "Factory utilities" },
  ],
  scenarios: [],
  runs: [
    {
      worldId: "w-1",
      ended: true,
      complete: true,
      incompleteReason: "",
      explore: false,
      seed: 7,
      plant: "factory-utilities",
      durationS: 1800,
      recordedAt: "2026-10-14T10:00:00Z",
    },
    {
      worldId: "w-2",
      ended: false,
      complete: false,
      incompleteReason: "",
      explore: true,
      seed: null,
      plant: "factory-utilities",
      durationS: 0,
      recordedAt: "2026-10-14T11:00:00Z",
    },
  ],
};
