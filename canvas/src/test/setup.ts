import "@testing-library/jest-dom/vitest";
import { cleanup } from "@testing-library/react";
import { afterEach, vi } from "vitest";
import { installQuickComputedStyle } from "./computedStyle";

installQuickComputedStyle();

afterEach(() => {
  cleanup();
  vi.unstubAllGlobals();
});
