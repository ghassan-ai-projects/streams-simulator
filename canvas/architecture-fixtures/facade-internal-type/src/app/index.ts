import type { Secret } from "./internal/types";
export function make(): Secret {
  return { secret: 1 };
}
