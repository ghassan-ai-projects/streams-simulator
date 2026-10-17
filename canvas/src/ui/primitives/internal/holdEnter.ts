import type { KeyboardEvent } from "react";

export function holdEnter(event: KeyboardEvent<HTMLFormElement>) {
  if (event.key === "Enter" && event.target instanceof HTMLSelectElement) event.preventDefault();
}
