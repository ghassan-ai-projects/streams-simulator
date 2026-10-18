import type { KindFile, Selection, TwinChoices } from "../../../model/devices";
import type { KindEntry } from "../../../model/workspace";
import type { CaptureSource, DeviceFailure, Outcome } from "../../ports";
import type { Store } from "../../shared";
import type { CaptureState } from "./captureState";
import type { DeviceState } from "./deviceDraft";
import type { TwinState } from "./twinDraft";

export interface DeviceRow {
  readonly path: string;
  readonly name: string;
  readonly metrics: number | null;
  readonly commands: number | null;
  readonly status: "invalid" | "needs-twin" | "twin-ready";
}

export interface CaptureSession extends Store<CaptureState> {
  read(source: CaptureSource, kinds: readonly KindEntry[]): Promise<void>;
  retry(): Promise<void>;
  save(name: string): Promise<void>;
  replace(name: string): Promise<void>;
  skip(name: string): void;
  close(): void;
}

export interface DeviceDesk {
  rows(entries: readonly KindEntry[], domains: readonly string[]): Promise<readonly DeviceRow[]>;
  create(name: string): Promise<Outcome<string, DeviceFailure>>;
  capture(): CaptureSession;
}

export interface TwinSession extends Store<TwinState> {
  open(path: string): Promise<void>;
  reload(): Promise<void>;
  retry(): Promise<void>;
  close(): void;
  edit(change: (choices: TwinChoices) => TwinChoices): void;
  include(path: string): Promise<void>;
  leave(path: string): void;
  persist(): Promise<void>;
  discard(): void;
  create(): Promise<void>;
  replace(): Promise<void>;
}

export interface DeviceSession extends Store<DeviceState> {
  readonly twin: TwinSession;
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
