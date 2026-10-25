import type { InstanceOrigin, NodeInterrupt, NodeStatus, PlantView, ReadoutBody, Shown } from "../../graph";
import { attentionOf, hallOf, showReading, worstOf } from "../../graph";
import type { Plant } from "../../plant";
import { channelOf } from "../../plant";
import { interruptOf } from "./interruptSelect";
import { modeLine } from "./interruptWords";
import type { PlaybackIndex } from "./playback";
import { deviceReadoutBody, readoutBody } from "./reviewRows";
import type { ReadoutPhase } from "./reviewRows";
import type { SessionLive } from "./sessionLive";
import { deviceOf, deviceReadings } from "./sessionLive";

export interface LiveNode {
  readonly shown: readonly Shown[];
  readonly observed: ReadonlySet<string>;
  readonly readout: ReadoutBody;
  readonly status: NodeStatus;
  readonly count: number;
  readonly fault: string;
  readonly effector: string;
  readonly origin: InstanceOrigin | null;
  readonly interrupt: NodeInterrupt | null;
}

export interface LiveFacts {
  readonly sessionLive?: SessionLive | null;
  readonly index: PlaybackIndex;
  readonly plant: Plant | null;
  readonly view: PlantView | null;
  readonly faulted: ReadonlyMap<string, readonly string[]>;
  readonly effectors: ReadonlyMap<string, string>;
  readonly origins: ReadonlyMap<string, InstanceOrigin>;
  readonly positionNs: number;
  readonly phase?: ReadoutPhase;
}

const noReadout: ReadoutBody = { rows: [], hidden: 0, hiddenAttention: null };

function statusOf(attention: NodeStatus, faulted: boolean): NodeStatus {
  if (attention === "alarm") return "alarm";
  return faulted ? "fault" : attention;
}

export function isRealBoard(origins: ReadonlyMap<string, InstanceOrigin>, id: string): boolean {
  return origins.get(id)?.kind === "real";
}

function shownAt(facts: LiveFacts, id: string): readonly Shown[] {
  const { index, plant, positionNs } = facts;
  if (isRealBoard(facts.origins, id)) {
    const device = deviceOf(facts.sessionLive ?? null, id);
    return device === null ? [] : deviceReadings(device).map((reading) => showReading(reading, null));
  }
  return index
    .readings(id, positionNs)
    .map((reading) => showReading(reading, plant === null ? null : channelOf(plant, id, reading.channel)));
}

function readoutOf(facts: LiveFacts, id: string, shown: readonly Shown[], observed: ReadonlySet<string>): ReadoutBody {
  const node = facts.view?.nodes.find((entry) => entry.id === id);
  if (node === undefined) return noReadout;
  if (isRealBoard(facts.origins, id) && shown.length > 0) return deviceReadoutBody(node.pins, node.channels, shown);
  return readoutBody(node.pins, node.channels, shown, observed, facts.phase);
}

function interruptShown(facts: LiveFacts, id: string): NodeInterrupt | null {
  const covered = interruptOf(facts.sessionLive ?? null, id);
  return covered === null ? null : { mode: covered.mode, latched: covered.latched, words: modeLine(covered) };
}

export function liveOf(facts: LiveFacts, id: string): LiveNode {
  const shown = shownAt(facts, id);
  const observed = new Set([...facts.index.observed(id), ...shown.map((reading) => reading.channel)]);
  return {
    shown,
    observed,
    readout: readoutOf(facts, id, shown, observed),
    status: statusOf(attentionOf(shown), facts.faulted.has(id)),
    count: 0,
    fault: (facts.faulted.get(id) ?? []).join(", "),
    effector: facts.effectors.get(id) ?? "",
    origin: facts.origins.get(id) ?? null,
    interrupt: interruptShown(facts, id),
  };
}

function hallNode(worst: { readonly status: NodeStatus; readonly count: number }): LiveNode {
  return {
    shown: [],
    observed: new Set<string>(),
    readout: noReadout,
    status: worst.status,
    count: worst.count,
    fault: "",
    effector: "",
    origin: null,
    interrupt: null,
  };
}

export function hallLiveOf(facts: LiveFacts, id: string): LiveNode {
  const members = facts.plant?.instances.filter((instance) => hallOf(instance.id) === id) ?? [];
  if (members.length === 0) return liveOf(facts, id);
  return hallNode(worstOf(members.map((member) => liveOf(facts, member.id).status)));
}
