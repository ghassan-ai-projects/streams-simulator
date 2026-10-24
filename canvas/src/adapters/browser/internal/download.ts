import type { FileSaver } from "../../../app/ports";

interface Anchor {
  href: string;
  download: string;
  click(): void;
  remove(): void;
}

interface DownloadEnvironment {
  createObjectURL(blob: Blob): string;
  revokeObjectURL(url: string): void;
  createAnchor(): Anchor;
  append(anchor: Anchor): void;
  later(run: () => void): void;
}

const markdown = "text/markdown;charset=utf-8";

export function browserFileSaver(env: DownloadEnvironment): FileSaver {
  return {
    save(name, text) {
      try {
        const url = env.createObjectURL(new Blob([text], { type: markdown }));
        const anchor = env.createAnchor();
        anchor.href = url;
        anchor.download = name;
        env.append(anchor);
        anchor.click();
        anchor.remove();
        env.later(() => {
          env.revokeObjectURL(url);
        });
        return true;
      } catch {
        return false;
      }
    },
  };
}

export function documentDownloads(page: Window & typeof globalThis): DownloadEnvironment {
  return {
    createObjectURL: (blob) => page.URL.createObjectURL(blob),
    revokeObjectURL: (url) => {
      page.URL.revokeObjectURL(url);
    },
    createAnchor: () => page.document.createElement("a"),
    append: (anchor) => {
      if (anchor instanceof page.HTMLAnchorElement) page.document.body.append(anchor);
    },
    later: (run) => {
      page.setTimeout(run, 0);
    },
  };
}
