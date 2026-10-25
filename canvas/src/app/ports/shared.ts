export type Outcome<T, F> = { readonly ok: true; readonly value: T } | { readonly ok: false; readonly failure: F };

export interface Diagnostic {
  readonly message: string;
  readonly displayCode: string;
}

export interface CompositionIssue {
  readonly pointer: string;
  readonly message: string;
}

export interface FileContent {
  readonly path: string;
  readonly content: string;
  readonly etag: string;
}

export interface ComposeSummary {
  readonly channels: number;
  readonly faults: number;
  readonly effectors: number;
  readonly digest: string;
}

declare const domainDocumentBrand: unique symbol;

export type DomainDocument = object & { readonly [domainDocumentBrand]: true };

export type PlantSource =
  { readonly kind: "file"; readonly path: string } | { readonly kind: "example"; readonly id: string };

export function success<T, F = never>(value: T): Outcome<T, F> {
  return { ok: true, value };
}

export function failure<T = never, F = never>(value: F): Outcome<T, F> {
  return { ok: false, failure: value };
}
