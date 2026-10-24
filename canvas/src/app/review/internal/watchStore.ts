import type { Watched, WindowChoice } from "../../../model/review";
import { toggled } from "../../../model/review";
import { Observable } from "../../shared";
import type { Store } from "../../shared";

export interface WatchState {
  readonly lists: ReadonlyMap<string, readonly Watched[]>;
  readonly open: boolean;
  readonly choice: WindowChoice | null;
}

export interface WatchSession extends Store<WatchState> {
  listOf(plant: string): readonly Watched[];
  toggle(plant: string, watched: Watched): void;
  clear(plant: string): void;
  show(): void;
  hide(): void;
  toggleOpen(): void;
  setChoice(choice: WindowChoice): void;
}

const none: readonly Watched[] = [];

export class WatchStore extends Observable<WatchState> implements WatchSession {
  constructor() {
    super({ lists: new Map(), open: false, choice: null });
  }

  listOf(plant: string): readonly Watched[] {
    return this.snapshot().lists.get(plant) ?? none;
  }

  toggle(plant: string, watched: Watched): void {
    this.replace(plant, toggled(this.listOf(plant), watched));
  }

  clear(plant: string): void {
    this.replace(plant, none);
  }

  show(): void {
    this.set({ ...this.snapshot(), open: true });
  }

  hide(): void {
    this.set({ ...this.snapshot(), open: false });
  }

  toggleOpen(): void {
    this.set({ ...this.snapshot(), open: !this.snapshot().open });
  }

  setChoice(choice: WindowChoice): void {
    this.set({ ...this.snapshot(), choice });
  }

  private replace(plant: string, list: readonly Watched[]): void {
    const state = this.snapshot();
    if (list === this.listOf(plant)) return;
    this.set({ ...state, lists: new Map(state.lists).set(plant, list) });
  }
}
