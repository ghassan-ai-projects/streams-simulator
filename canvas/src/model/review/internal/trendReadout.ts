import { showReading } from "../../graph";
import type { Plant } from "../../plant";
import { channelOf } from "../../plant";
import { qualityWords } from "../../shared/quality";
import { codeTitle } from "../../shared/units";
import { ageWords } from "./observedRows";
import type { PlaybackIndex } from "./playback";
import { runClock } from "./timeline";
import type { Watched } from "./watchList";

export interface Readout {
  readonly value: string;
  readonly unit: string;
  readonly unitTitle: string;
  readonly observedAt: string;
  readonly age: string;
  readonly quality: string;
  readonly words: string;
}

export interface ReadoutFacts {
  readonly index: PlaybackIndex;
  readonly plant: Plant | null;
  readonly atNs: number;
  readonly referenceNs: number;
  readonly startNs: number;
}

export function readoutOf(facts: ReadoutFacts, watched: Watched): Readout | null {
  const reading = facts.index.reading(watched.entity, watched.channel, facts.atNs);
  if (reading === null) return null;
  const spec = facts.plant === null ? null : channelOf(facts.plant, watched.entity, watched.channel);
  const shown = showReading(reading, spec);
  const observedAt = runClock(reading.atNs, facts.startNs);
  const quality = qualityWords(reading.quality);
  const unit = shown.unit === "" ? "" : ` ${shown.unit}`;
  const age = ageWords(facts.referenceNs - reading.atNs);
  const marked = quality === "" ? "" : `, ${quality}`;
  return {
    value: shown.value,
    unit: shown.unit,
    unitTitle: codeTitle(reading.unit),
    observedAt,
    age,
    quality,
    words: `${shown.value}${unit}${marked}, observed at ${observedAt}, ${age}`,
  };
}
