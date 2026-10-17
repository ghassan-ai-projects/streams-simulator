import { nextPreference, parsePreference, resolveTheme } from "../../../model/shell";
import type { Theme, ThemePreference } from "../../../model/shell";
import type { PreferenceStore, SystemTheme } from "../../ports";
import { Observable } from "../../shared";

const key = "streamsim.canvas.theme";

export interface ThemeState {
  readonly preference: ThemePreference;
  readonly theme: Theme;
}

function stateFor(preference: ThemePreference, system: SystemTheme): ThemeState {
  return { preference, theme: resolveTheme(preference, system.prefersDark()) };
}

export class ThemeStore extends Observable<ThemeState> {
  constructor(
    private readonly preferences: PreferenceStore,
    private readonly system: SystemTheme,
  ) {
    super(stateFor(parsePreference(preferences.read(key)), system));
    system.subscribe(() => {
      this.apply(this.snapshot().preference);
    });
  }

  cycle(): void {
    const preference = nextPreference(this.snapshot().preference);
    this.preferences.write(key, preference);
    this.apply(preference);
  }

  private apply(preference: ThemePreference): void {
    this.set(stateFor(preference, this.system));
  }
}
