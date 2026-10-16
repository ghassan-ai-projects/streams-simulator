import type { Bridge } from "./ports";
import { EditorStore } from "./editorStore";
import type { EditorState } from "./editorStore";
import { PlantStore } from "./plantStore";
import type { PlantState } from "./plantStore";
import { ReviewStore } from "./reviewStore";
import type { ReviewState } from "./reviewStore";
import { createOperationIdentity } from "./shared/operationIdentity";
import type { Clock, Scheduler } from "./clock";
import type { RunEntry, Workspace } from "../model/bridge";
import type { Composition, NewEquipment } from "../model/composition";
import type { PlaybackIndex } from "../model/playback";
import type { Badge } from "../model/report";
import type { Presets } from "../model/library";

export interface PlantSession {
  readonly snapshot: () => PlantState;
  readonly subscribe: (listener: () => void) => () => void;
  load(path: string, workspace: Workspace): Promise<void>;
  show(key: string, domain: unknown): Promise<void>;
  select(id: string | null): void;
  close(): void;
}

export interface EditorSession {
  readonly snapshot: () => EditorState;
  readonly subscribe: (listener: () => void) => () => void;
  open(path: string): Promise<void>;
  close(): void;
  reload(): Promise<void>;
  savable(): boolean;
  unsavedChanges(): boolean;
  select(id: string | null): void;
  edit(change: (doc: Composition) => Composition, selected?: string | null): void;
  revert(): void;
  undo(): void;
  redo(): void;
  add(entry: NewEquipment): Promise<void>;
  presets(use: string): Promise<Presets>;
  save(): Promise<void>;
}

export interface ReviewSession {
  readonly snapshot: () => ReviewState;
  readonly subscribe: (listener: () => void) => () => void;
  readonly index: PlaybackIndex;
  open(worldId: string, known?: RunEntry): void;
  refreshCompleted(worldId: string, known: RunEntry | undefined): void;
  close(): void;
  badgeFor(worldId: string): Badge | null;
  seek(positionNs: number): void;
  play(): void;
  pause(): void;
  setSpeed(speed: number): void;
  tick(elapsedMs: number): void;
  select(entity: string | null): void;
  reveal(positionNs: number, entity: string): void;
  showTab(tab: ReviewState["tab"]): void;
}

export type { Draft, EditorState } from "./editorStore";
export type { PlantState } from "./plantStore";
export type { ReviewState } from "./reviewStore";

export interface PlantModule {
  createSession(): PlantSession;
}

export interface EditorModule {
  createSession(): EditorSession;
}

export interface ReviewModule {
  createSession(): ReviewSession;
}

export function createPlantModule(bridge: Bridge): PlantModule {
  return { createSession: () => new PlantStore(bridge, createOperationIdentity()) };
}

export function createEditorModule(bridge: Bridge): EditorModule {
  return { createSession: () => new EditorStore(bridge, createOperationIdentity()) };
}

export function createReviewModule(bridge: Bridge, clock: Clock, scheduler: Scheduler): ReviewModule {
  return {
    createSession: () => new ReviewStore(bridge, clock, scheduler, createOperationIdentity()),
  };
}
