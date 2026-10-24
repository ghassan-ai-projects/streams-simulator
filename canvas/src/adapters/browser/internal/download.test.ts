import { describe, expect, it, vi } from "vitest";
import { browserFileSaver, documentDownloads } from "./download";

function environment(failing = false) {
  const anchor = { href: "", download: "", click: vi.fn(), remove: vi.fn() };
  const append = vi.fn();
  const later: (() => void)[] = [];
  const created: Blob[] = [];
  const env = {
    createObjectURL: (blob: Blob) => {
      if (failing) throw new Error("no URL");
      created.push(blob);
      return "blob:handover";
    },
    revokeObjectURL: vi.fn(),
    createAnchor: () => anchor,
    append,
    later: (run: () => void) => {
      later.push(run);
    },
  };
  return { env, anchor, append, later, created };
}

describe("saving a text file through the browser", () => {
  it("downloads the text under the name, as markdown, and releases the object URL afterwards", async () => {
    const { env, anchor, append, later, created } = environment();
    expect(browserFileSaver(env).save("w-1-handover.md", "# Run w-1\n")).toBe(true);
    expect(anchor).toMatchObject({ href: "blob:handover", download: "w-1-handover.md" });
    expect(append).toHaveBeenCalledWith(anchor);
    expect(anchor.click).toHaveBeenCalledOnce();
    expect(anchor.remove).toHaveBeenCalledOnce();
    expect(created[0]?.type).toBe("text/markdown;charset=utf-8");
    expect(await created[0]?.text()).toBe("# Run w-1\n");
    expect(env.revokeObjectURL).not.toHaveBeenCalled();
    later.forEach((run) => {
      run();
    });
    expect(env.revokeObjectURL).toHaveBeenCalledWith("blob:handover");
  });

  it("says the file was not saved when the browser cannot make an object URL", () => {
    const { env, anchor } = environment(true);
    expect(browserFileSaver(env).save("x.md", "x")).toBe(false);
    expect(anchor.click).not.toHaveBeenCalled();
  });
});

describe("the downloads of a page", () => {
  it("makes an object URL, appends a real anchor and releases the URL on the next turn", () => {
    vi.useFakeTimers();
    const create = vi.fn(() => "blob:page");
    const revoke = vi.fn();
    Object.assign(window.URL, { createObjectURL: create, revokeObjectURL: revoke });
    const clicked = vi.spyOn(HTMLAnchorElement.prototype, "click").mockImplementation(() => undefined);
    expect(browserFileSaver(documentDownloads(window)).save("w-1-handover.md", "text")).toBe(true);
    expect(create).toHaveBeenCalledOnce();
    expect(clicked).toHaveBeenCalledOnce();
    expect(document.querySelector("a[download]")).toBeNull();
    expect(revoke).not.toHaveBeenCalled();
    vi.runAllTimers();
    expect(revoke).toHaveBeenCalledWith("blob:page");
    vi.useRealTimers();
    clicked.mockRestore();
  });
});
