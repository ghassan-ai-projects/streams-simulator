import type { SavedTrends, TrendSize, Watched, WindowChoice } from "../../../model/review";
import { savedTrendsOf, savedTrendsText, toggled } from "../../../model/review";
import type { PreferenceStore } from "../../ports";
import { Observable } from "../../shared";
import type { Store } from "../../shared";

export interface WatchState {
  readonly lists: ReadonlyMap<string, readonly Watched[]>;
  readonly open: boolean;
  readonly choice: WindowChoice | null;
  readonly size: TrendSize | null;
}

export interface WatchSession extends Store<WatchState> {
  listOf(plant: string): readonly Watched[];
  toggle(plant: string, watched: Watched): void;
  clear(plant: string): void;
  show(): void;
  hide(): void;
  toggleOpen(): void;
  setChoice(choice: WindowChoice): void;
  setSize(size: TrendSize): void;
  restore(plant: string, exists: (watched: Watched) => boolean): void;
}

const none: readonly Watched[] = [];
const nowhere: PreferenceStore = { read: () => null, write: () => undefined };

const keyOf = (plant: string) => `streamsim.trends.${plant}`;

export class WatchStore extends Observable<WatchState> implements WatchSession {
  private active: string | null = null;
  private readonly restored = new Set<string>();

  constructor(private readonly kept: PreferenceStore = nowhere) {
    super({ lists: new Map(), open: false, choice: null, size: null });
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
    this.change({ open: true });
  }

  hide(): void {
    this.change({ open: false });
  }

  toggleOpen(): void {
    this.change({ open: !this.snapshot().open });
  }

  setChoice(choice: WindowChoice): void {
    this.set({ ...this.snapshot(), choice });
  }

  setSize(size: TrendSize): void {
    this.change({ size });
  }

  restore(plant: string, exists: (watched: Watched) => boolean): void {
    this.active = plant;
    if (this.restored.has(plant)) return;
    this.restored.add(plant);
    const saved = savedTrendsOf(this.kept.read(keyOf(plant)));
    if (saved === null) return;
    const lists = new Map(this.snapshot().lists);
    if (this.listOf(plant).length === 0) lists.set(plant, saved.watched.filter(exists));
    this.set({ ...this.snapshot(), lists, open: saved.open, size: saved.size });
  }

  private change(next: Partial<WatchState>): void {
    this.set({ ...this.snapshot(), ...next });
    this.save();
  }

  private replace(plant: string, list: readonly Watched[]): void {
    const state = this.snapshot();
    if (list === this.listOf(plant)) return;
    this.active = plant;
    this.set({ ...state, lists: new Map(state.lists).set(plant, list) });
    this.save();
  }

  private save(): void {
    if (this.active === null) return;
    const { open, size } = this.snapshot();
    const saved: SavedTrends = { open, size, watched: this.listOf(this.active) };
    this.kept.write(keyOf(this.active), savedTrendsText(saved));
  }
}
