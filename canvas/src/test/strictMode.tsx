import { StrictMode } from "react";
import type { ReactElement } from "react";
import { render } from "@testing-library/react";

export function renderStrict(ui: ReactElement) {
  return render(<StrictMode>{ui}</StrictMode>);
}
