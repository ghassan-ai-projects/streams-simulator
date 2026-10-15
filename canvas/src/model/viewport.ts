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

export function followZoom(
  box: Rect,
  viewport: Viewport,
  screen: { readonly width: number; readonly height: number },
  focusZoom: number,
): number | null {
  if (inView(box, viewport, screen)) return null;
  return Math.max(viewport.zoom, focusZoom);
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

export const openingFloor = 0.4;

export const textFloorPx = 8;

const pixelSteps = [1, 1.25, 1.5, 2, 2.5, 3, 4, 5, 6, 8, 10, 13, 16, 20] as const;

export function screenPixel(zoom: number): number {
  const wanted = 1 / Math.max(zoom, 0.05);
  return pixelSteps.find((step) => step >= wanted) ?? 20;
}

export type Tier = "full" | "compact" | "overview" | "map";

export function tierOf(zoom: number): Tier {
  if (zoom >= 1) return "full";
  if (zoom >= openingFloor) return "compact";
  return zoom >= 0.25 ? "overview" : "map";
}

interface Size {
  readonly width: number;
  readonly height: number;
}

interface Span {
  readonly start: number;
  readonly extent: number;
}

function centred(content: Span, room: number, zoom: number, pad: number, focus: Span | null): number {
  const used = content.extent * zoom;
  if (used <= room) return pad + (room - used) / 2 - content.start * zoom;
  const lead = pad - content.start * zoom;
  if (focus === null) return lead;
  const trail = pad + room - (content.start + content.extent) * zoom;
  const wanted = pad + room / 2 - (focus.start + focus.extent / 2) * zoom;
  return Math.min(lead, Math.max(trail, wanted));
}

export interface FitLimits {
  readonly pad: number;
  readonly headroom: number;
  readonly minZoom: number;
  readonly maxZoom: number;
}

export function openingViewport(bounds: Rect, screen: Size, limits: FitLimits, anchor: Rect | null = null): Viewport {
  const top = limits.pad + limits.headroom;
  const room = { width: screen.width - 2 * limits.pad, height: screen.height - top - limits.pad };
  const fit = Math.min(room.width / bounds.width, room.height / bounds.height);
  const zoom = Math.min(limits.maxZoom, Math.max(limits.minZoom, fit));
  const across = { start: bounds.x, extent: bounds.width };
  const down = { start: bounds.y, extent: bounds.height };
  return {
    x: centred(across, room.width, zoom, limits.pad, anchor && { start: anchor.x, extent: anchor.width }),
    y: centred(down, room.height, zoom, top, anchor && { start: anchor.y, extent: anchor.height }),
    zoom,
  };
}

interface Axis {
  readonly start: number;
  readonly extent: number;
  readonly room: number;
  readonly margin: number;
}

function keptInside(offset: number, wasCut: boolean, axis: Axis): number {
  const { start, extent, room, margin } = axis;
  const first = margin - start;
  if (offset < first) return wasCut ? offset : first;
  const last = room - margin - (start + extent);
  return extent <= room - 2 * margin && offset > last ? last : offset;
}

export function steppedViewport(current: Viewport, screen: Size, zoom: number, bounds: Rect, margin: number): Viewport {
  const ratio = zoom / current.zoom;
  const x = screen.width / 2 - (screen.width / 2 - current.x) * ratio;
  const y = screen.height / 2 - (screen.height / 2 - current.y) * ratio;
  const cutX = current.x + bounds.x * current.zoom < margin - 0.5;
  const cutY = current.y + bounds.y * current.zoom < margin - 0.5;
  const across = { start: bounds.x * zoom, extent: bounds.width * zoom, room: screen.width, margin };
  const down = { start: bounds.y * zoom, extent: bounds.height * zoom, room: screen.height, margin };
  return { x: keptInside(x, cutX, across), y: keptInside(y, cutY, down), zoom };
}

export function sameViewport(a: Viewport, b: Viewport): boolean {
  return Math.abs(a.x - b.x) < 0.5 && Math.abs(a.y - b.y) < 0.5 && Math.abs(a.zoom - b.zoom) < 1e-4;
}
