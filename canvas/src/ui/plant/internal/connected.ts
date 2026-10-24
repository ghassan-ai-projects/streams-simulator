import type { ReactNode } from "react";
import type { InstanceOrigin, ViewPresentation } from "../../../model/graph";
import type { NodeLive } from "../../graph";

export interface ConnectedDevice {
  readonly line: string;
  readonly readings: readonly {
    readonly name: string;
    readonly valueText: string;
    readonly stateWords: string;
    readonly unitTitle: string;
  }[];
}

export interface PlannedFault {
  readonly entity: string;
  readonly fault: string;
  readonly atS: number;
}

export interface Connected {
  readonly kind: "session" | "watch";
  readonly live: (id: string) => NodeLive;
  readonly observed: (id: string) => ReactNode;
  readonly device: (id: string) => ConnectedDevice | null;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly interrupts: ReactNode;
  readonly run: string;
  readonly view?: ViewPresentation;
  readonly notes?: ReadonlyMap<string, number>;
  readonly trends?: (select: (entity: string) => void, planned: readonly PlannedFault[]) => ReactNode;
  readonly trendsToggle?: ReactNode;
}
