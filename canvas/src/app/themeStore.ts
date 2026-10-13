import { nextPreference, parsePreference, resolveTheme } from "../model/theme";
import type { Theme, ThemePreference } from "../model/theme";
import type { PreferenceStore, SystemTheme } from "./ports";

const key = "streamsim.canvas.theme";

export interface ThemeState {
  readonly preference: ThemePreference;
  readonly theme: Theme;
}

export class ThemeStore {
  private state: ThemeState;
  private readonly listeners = new Set<() => void>();

  constructor(
    private readonly preferences: PreferenceStore,
    private readonly system: SystemTheme,
  ) {
    this.state = this.stateFor(parsePreference(preferences.read(key)));
    system.subscribe(() => {
      this.update(this.state.preference);
    });
  }

  readonly snapshot = (): ThemeState => this.state;

  readonly subscribe = (listener: () => void): (() => void) => {
    this.listeners.add(listener);
    return () => this.listeners.delete(listener);
  };

  cycle(): void {
    const preference = nextPreference(this.state.preference);
    this.preferences.write(key, preference);
    this.update(preference);
  }

  private update(preference: ThemePreference): void {
    this.state = this.stateFor(preference);
    this.listeners.forEach((listener) => {
      listener();
    });
  }

  private stateFor(preference: ThemePreference): ThemeState {
    return { preference, theme: resolveTheme(preference, this.system.prefersDark()) };
  }
}
