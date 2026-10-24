import type { ReactNode } from "react";
import type { InstanceOrigin } from "../../../model/graph";
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

export interface Connected {
  readonly kind: "session" | "watch";
  readonly live: (id: string) => NodeLive;
  readonly observed: (id: string) => ReactNode;
  readonly device: (id: string) => ConnectedDevice | null;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly interrupts: ReactNode;
  readonly run: string;
}
