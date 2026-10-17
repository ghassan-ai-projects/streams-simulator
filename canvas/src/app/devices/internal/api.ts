import type { KindFile, Selection } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { DeviceFailure, Outcome } from "../../ports";
import type { Store } from "../../shared";
import type { DeviceState } from "./deviceDraft";

export interface DeviceRow {
  readonly path: string;
  readonly name: string;
  readonly metrics: number | null;
  readonly commands: number | null;
}

export interface DeviceDesk {
  rows(entries: readonly KindEntry[]): Promise<readonly DeviceRow[]>;
  create(name: string): Promise<Outcome<string, DeviceFailure>>;
}

export interface DeviceSession extends Store<DeviceState> {
  open(path: string): Promise<void>;
  close(): void;
  reload(): Promise<void>;
  select(selection: Selection): void;
  edit(change: (kind: KindFile) => KindFile): void;
  addMetric(): void;
  addCommand(): void;
  remove(selection: Selection): void;
  revert(): void;
  save(): Promise<void>;
  check(): Promise<void>;
}
