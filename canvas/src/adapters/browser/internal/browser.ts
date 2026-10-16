import type { Location, PreferenceStore, SystemTheme } from "../../../app/ports";

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

export function hashLocation(window: Window): Location {
  return {
    read: () => window.location.hash,
    write(hash, replace = false) {
      if (window.location.hash === hash) return;
      if (replace) window.history.replaceState(null, "", hash);
      else window.location.hash = hash;
    },
    subscribe(listener) {
      window.addEventListener("hashchange", listener);
      return () => {
        window.removeEventListener("hashchange", listener);
      };
    },
  };
}
