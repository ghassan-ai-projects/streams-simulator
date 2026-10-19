import type { Clipboard } from "../../../app/ports";

interface ClipboardApi {
  writeText(text: string): Promise<void>;
}

export function browserClipboard(api: ClipboardApi | undefined): Clipboard {
  return {
    async copy(text) {
      if (api === undefined) return false;
      try {
        await api.writeText(text);
        return true;
      } catch {
        return false;
      }
    },
  };
}
