export function humanise(id: string): string {
  const parts = id.split("__");
  const meaningful = parts.length > 1 ? parts.slice(1) : parts;
  return meaningful.map(spaced).join(" · ");
}

export function kindLabel(kind: string): string {
  return humanise(kind).replace(/ \d+$/, "");
}

export function shortChannel(channel: string): string {
  const dot = channel.indexOf(".");
  return dot < 0 ? channel : channel.slice(dot + 1);
}

export function seedWords(seed: number | null): string {
  return seed === null ? "" : `seed ${String(seed)}`;
}

export function counted(n: number, noun: string, many = `${noun}s`): string {
  return `${String(n)} ${n === 1 ? noun : many}`;
}

export function spaced(name: string): string {
  return name.replaceAll("_", " ");
}
