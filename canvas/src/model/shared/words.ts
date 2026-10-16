export function humanise(id: string): string {
  const parts = id.split("__");
  const meaningful = parts.length > 1 ? parts.slice(1) : parts;
  return meaningful.map((part) => part.replaceAll("_", " ")).join(" · ");
}

export function kindLabel(kind: string): string {
  return humanise(kind).replace(/ \d+$/, "");
}

export function shortChannel(channel: string): string {
  const dot = channel.indexOf(".");
  return dot < 0 ? channel : channel.slice(dot + 1);
}
