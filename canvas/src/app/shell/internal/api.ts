import type { Route, Theme, ThemePreference } from "../../../model/shell";
import type { Store } from "../../shared";

export type Side = "left" | "right";

export interface PanelState {
  readonly left: boolean;
  readonly right: boolean;
}

export interface ThemeState {
  readonly preference: ThemePreference;
  readonly theme: Theme;
}

export interface LeaveAttempt {
  readonly route: Route;
  readonly replace: boolean;
}

export interface PanelStore extends Store<PanelState> {
  show(side: Side): void;
  toggle(side: Side): void;
}

export interface ThemeStore extends Store<ThemeState> {
  cycle(): void;
}

export interface NavigationStore extends Store<Route> {
  readonly leaving: Store<LeaveAttempt | null>;
  guardUnsaved(check: () => boolean): () => void;
  go(route: Route): void;
  replace(route: Route): void;
  leave(): void;
  stay(): void;
}
