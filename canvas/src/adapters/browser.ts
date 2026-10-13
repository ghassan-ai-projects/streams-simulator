import type { PreferenceStore, SystemTheme } from "../app/ports";

export function localPreferences(storage: Storage): PreferenceStore {
  return {
    read(key) {
      try {
        return storage.getItem(key);
      } catch {
        return null;
      }
    },
    write(key, value) {
      try {
        storage.setItem(key, value);
      } catch {
        return;
      }
    },
  };
}

export function mediaSystemTheme(media: MediaQueryList): SystemTheme {
  return {
    prefersDark: () => media.matches,
    subscribe(listener) {
      media.addEventListener("change", listener);
      return () => {
        media.removeEventListener("change", listener);
      };
    },
  };
}
