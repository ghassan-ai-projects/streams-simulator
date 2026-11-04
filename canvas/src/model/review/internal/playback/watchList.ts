export interface Watched {
  readonly entity: string;
  readonly channel: string;
}

export const watchLimit = 8;

export function watchKey(watched: Watched): string {
  return `${watched.entity}/${watched.channel}`;
}

export function isWatched(list: readonly Watched[], watched: Watched): boolean {
  return list.some((entry) => watchKey(entry) === watchKey(watched));
}

export function canWatch(list: readonly Watched[], watched: Watched): boolean {
  return isWatched(list, watched) || list.length < watchLimit;
}

export function toggled(list: readonly Watched[], watched: Watched): readonly Watched[] {
  if (isWatched(list, watched)) return list.filter((entry) => watchKey(entry) !== watchKey(watched));
  return canWatch(list, watched) ? [...list, watched] : list;
}
