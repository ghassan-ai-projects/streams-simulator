export interface Rect {
  readonly x: number;
  readonly y: number;
  readonly width: number;
  readonly height: number;
}

export interface Viewport {
  readonly x: number;
  readonly y: number;
  readonly zoom: number;
}

export function inView(
  box: Rect,
  viewport: Viewport,
  screen: { readonly width: number; readonly height: number },
): boolean {
  const left = box.x * viewport.zoom + viewport.x;
  const top = box.y * viewport.zoom + viewport.y;
  return (
    left >= 0 &&
    top >= 0 &&
    left + box.width * viewport.zoom <= screen.width &&
    top + box.height * viewport.zoom <= screen.height
  );
}

export function centreOf(box: Rect): { readonly x: number; readonly y: number } {
  return { x: box.x + box.width / 2, y: box.y + box.height / 2 };
}

export type Tier = "full" | "compact" | "overview" | "map";

export function tierOf(zoom: number): Tier {
  if (zoom >= 1) return "full";
  if (zoom >= 0.5) return "compact";
  return zoom >= 0.25 ? "overview" : "map";
}

export function anchorsOnly(tier: Tier): boolean {
  return tier === "overview" || tier === "map";
}

interface Size {
  readonly width: number;
  readonly height: number;
}

function placed(start: number, extent: number, room: number, zoom: number, pad: number): number {
  const used = extent * zoom;
  return used <= room ? pad + (room - used) / 2 - start * zoom : pad - start * zoom;
}

export function openingViewport(
  bounds: Rect,
  screen: Size,
  limits: { readonly pad: number; readonly minZoom: number; readonly maxZoom: number },
): Viewport {
  const room = { width: screen.width - 2 * limits.pad, height: screen.height - 2 * limits.pad };
  const fit = Math.min(room.width / bounds.width, room.height / bounds.height);
  const zoom = Math.min(limits.maxZoom, Math.max(limits.minZoom, fit));
  return {
    x: placed(bounds.x, bounds.width, room.width, zoom, limits.pad),
    y: placed(bounds.y, bounds.height, room.height, zoom, limits.pad),
    zoom,
  };
}
