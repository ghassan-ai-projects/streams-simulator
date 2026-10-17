import { Observable } from "../../shared";
import type { PreferenceStore } from "../../ports";

export type Side = "left" | "right";

export interface PanelState {
  readonly left: boolean;
  readonly right: boolean;
}

const keys: Readonly<Record<Side, string>> = {
  left: "streamsim.canvas.panel.left",
  right: "streamsim.canvas.panel.right",
};

export class PanelStore extends Observable<PanelState> {
  constructor(private readonly preferences: PreferenceStore) {
    super({ left: preferences.read(keys.left) !== "closed", right: preferences.read(keys.right) !== "closed" });
  }

  show(side: Side): void {
    if (!this.snapshot()[side]) this.toggle(side);
  }

  toggle(side: Side): void {
    const open = !this.snapshot()[side];
    this.preferences.write(keys[side], open ? "open" : "closed");
    this.set({ ...this.snapshot(), [side]: open });
  }
}
