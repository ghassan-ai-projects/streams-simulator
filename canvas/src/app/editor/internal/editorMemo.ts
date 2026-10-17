export interface StoreMemo {
  readonly etag: string;
  readonly domainPath: string;
  readonly domainEtag: string;
}

export const emptyMemo: StoreMemo = { etag: "", domainPath: "", domainEtag: "" };

export function outputEtagOf(memo: StoreMemo, output: string): string {
  return output === memo.domainPath ? memo.domainEtag : "";
}
