import type { CssCap, ModuleSpec, Register } from "./model.ts";

export function movedAllowance(previous: Register, current: Register, path: string, entry: CssCap): boolean {
  const moved = Object.entries(previous.css.globalCaps).some(
    ([old, cap]) =>
      !(old in current.css.globalCaps) &&
      basename(old) === basename(path) &&
      movedBetweenModules(previous, current, old, path) &&
      cap.global >= entry.global &&
      cap.important >= entry.important,
  );
  return moved || approvedVendorAllowanceMove(previous, current, path, entry);
}

export function approvedVendorAllowanceMove(
  previous: Register,
  current: Register,
  path: string,
  entry: CssCap,
): boolean {
  if (path !== "src/ui/graph/internal/vendor.module.css") return false;
  if (entry.global !== 6 || entry.important !== 2 || !newVendorRegistration(previous, current)) return false;
  const old = previous.css.globalCaps["src/ui/plant/GraphControls.module.css"];
  return old?.global === 0 && old.important === 2;
}

function newVendorRegistration(previous: Register, current: Register): boolean {
  const path = "src/ui/graph/internal/vendor.module.css";
  return (
    previous.css.globalCaps["src/ui/plant/GraphControls.module.css"] !== undefined &&
    current.css.globalCaps["src/ui/plant/GraphControls.module.css"] === undefined &&
    current.css.vendorFiles.includes(path) &&
    !previous.css.vendorFiles.includes(path) &&
    previous.modules.ui?.folders?.includes("src/ui") === true &&
    current.modules["ui/graph"]?.folders?.includes("src/ui/graph") === true
  );
}

export function moved<T extends { readonly phase: string }>(
  before: Readonly<Record<string, T>>,
  after: Readonly<Record<string, T>>,
  previous: Register,
  current: Register,
  path: string,
  count: number,
): boolean {
  return Object.keys(before).some((old) => {
    if (old in after || old === path) return false;
    const entry = before[old];
    return (
      entry !== undefined &&
      basename(old) === basename(path) &&
      movedBetweenModules(previous, current, old, path) &&
      size(entry) >= count
    );
  });
}

export function movedBetweenModules(previous: Register, current: Register, before: string, after: string): boolean {
  const source = ownerAt(previous.modules, before);
  const destination = ownerAt(current.modules, after);
  if (source === undefined || destination === undefined || source.name === destination.name) return false;
  if (source.module.layer !== destination.module.layer) return false;
  return !(previous.modules[destination.name]?.folders ?? []).includes(destination.folder);
}

export function basename(path: string): string {
  return path.slice(path.lastIndexOf("/") + 1);
}

export function movedDebt(previous: Register, current: Register, entry: Register["debt"][number]): boolean {
  const moved = previous.debt.filter(
    (old) =>
      old.rule === entry.rule &&
      !current.debt.some((now) => now.rule === old.rule && now.file === old.file && now.target === old.target) &&
      old.count >= entry.count &&
      movedDebtIdentity(previous, current, old, entry),
  );
  return moved.length > 0;
}

function movedDebtIdentity(
  previous: Register,
  current: Register,
  old: Register["debt"][number],
  entry: Register["debt"][number],
): boolean {
  if (old.file === entry.file && old.target !== undefined && entry.target !== undefined) {
    return (
      basename(old.target) === basename(entry.target) &&
      movedBetweenModules(previous, current, old.target, entry.target)
    );
  }
  return (
    old.target === entry.target &&
    basename(old.file) === basename(entry.file) &&
    movedBetweenModules(previous, current, old.file, entry.file)
  );
}

function ownerAt(
  modules: Register["modules"],
  path: string,
): { readonly name: string; readonly module: ModuleSpec; readonly folder: string } | undefined {
  return Object.entries(modules)
    .flatMap(([name, module]) =>
      (module.folders ?? [])
        .filter((folder) => path.startsWith(`${folder}/`))
        .map((folder) => ({ name, module, folder })),
    )
    .sort((left, right) => right.folder.length - left.folder.length)[0];
}

function size(entry: { readonly phase: string }): number {
  const value = entry as { readonly cap?: number; readonly literals?: number };
  return value.cap ?? value.literals ?? 0;
}
