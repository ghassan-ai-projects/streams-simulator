import type { InstanceOrigin } from "../../../model/graph";
import type { NodeLive } from "../../graph";

export interface Connected {
  readonly live: (id: string) => NodeLive;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly run: string;
}
