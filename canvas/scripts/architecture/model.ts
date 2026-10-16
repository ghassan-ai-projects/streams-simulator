export interface Violation {
  readonly rule: string;
  readonly file: string;
  readonly target?: string;
  readonly line: number;
  readonly message: string;
}

export interface DebtEntry {
  readonly rule: string;
  readonly file: string;
  readonly target?: string;
  readonly count: number;
  readonly phase: string;
  readonly reason?: string;
}

export interface ModuleSpec {
  readonly kind: string;
  readonly layer: string;
  readonly lifetime: string;
  readonly responsibility: string;
  readonly folders?: readonly string[];
  readonly files?: readonly string[];
  readonly mayImport?: readonly string[];
  readonly mayImportAny?: boolean;
  readonly callers?: readonly string[];
  readonly entrypoint: string | null;
}

export interface CssCap {
  readonly global: number;
  readonly important: number;
  readonly phase: string;
}

export interface CssSpec {
  readonly globalEntrypoints: readonly string[];
  readonly globalOwners: readonly string[];
  readonly vendorFiles: readonly string[];
  readonly defaultCap: number;
  readonly caps: Readonly<Record<string, { readonly cap: number; readonly phase: string }>>;
  readonly globalCaps: Readonly<Record<string, CssCap>>;
  readonly colorExceptions: Readonly<
    Record<string, { readonly literals: number; readonly reason: string; readonly phase: string }>
  >;
}

export interface RuleSpec {
  readonly actionNames: readonly string[];
  readonly protocolOwner: string;
  readonly protocolProductionOnly: boolean;
  readonly restrictedAppImports: readonly string[];
  readonly restrictedAppGlobals: readonly string[];
  readonly restrictedUiImports: readonly string[];
  readonly clockGlobals: readonly string[];
  readonly booleanFactNames: readonly string[];
  readonly fakeTimerTests: readonly string[];
  readonly depcruiseRules: readonly { readonly name: string; readonly severity: string }[];
}

export interface EmittedSpec {
  readonly order: readonly { readonly name: string; readonly selector: string }[];
  readonly important: number;
  readonly zIndex: number;
  readonly positioned: number;
  readonly note: string;
}

export interface Register {
  readonly version: number;
  readonly source: string;
  readonly modules: Readonly<Record<string, ModuleSpec>>;
  readonly targetModules: Readonly<Record<string, ModuleSpec>>;
  readonly targets: Readonly<Record<string, string>>;
  readonly features: Readonly<Record<string, string>>;
  readonly testSupport: readonly string[];
  readonly css: CssSpec;
  readonly floors: Readonly<Record<string, Readonly<Record<string, number>>>>;
  readonly rules: RuleSpec;
  readonly emittedCss: EmittedSpec;
  readonly debt: readonly DebtEntry[];
}

export function violation(rule: string, file: string, line: number, message: string, target?: string): Violation {
  return target === undefined ? { rule, file, line, message } : { rule, file, target, line, message };
}

export function debtKey(entry: { readonly rule: string; readonly file: string; readonly target?: string }): string {
  return `${entry.rule}\u0000${entry.file}\u0000${entry.target ?? ""}`;
}

export function violationKey(entry: Violation): string {
  return debtKey(entry);
}
