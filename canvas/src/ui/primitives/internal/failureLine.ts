export function failureLine(diagnostic: { readonly message: string; readonly displayCode: string }): string {
  return `${diagnostic.message} (${diagnostic.displayCode})`;
}
