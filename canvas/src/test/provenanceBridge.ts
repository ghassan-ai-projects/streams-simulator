import { success } from "../app/ports";
import type { FakeBridge } from "./fakeBridge";
import { miniBridge, miniWorkspace } from "./screenFixture";
import { assumptionsWire, diffWire, lineageOf, miniLineagePlant } from "./provenanceFixture";

export const miniDomainPath = "domains/mini.domain.json";

function withoutComposition(bridge: FakeBridge): FakeBridge {
  bridge.workspaceOutcome = success({
    ...miniWorkspace,
    compositions: [],
    kinds: [{ path: "kinds/fan.kind.json", kind: "fan", title: "" }],
  });
  return bridge;
}

export function provenanceBridge(
  extra: Record<string, unknown>,
  read: "composition" | "none" = "composition",
): FakeBridge {
  const bridge = miniBridge();
  if (read === "none") withoutComposition(bridge);
  bridge.actions.set("domain.lineage", success(lineageOf(miniLineagePlant(extra))));
  bridge.actions.set("domain.assumptions", success(assumptionsWire));
  bridge.actions.set("domain.diff", success(diffWire));
  return bridge;
}
