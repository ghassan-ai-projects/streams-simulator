import type { Page } from "@playwright/test";

export async function tokenColor(page: Page, token: string, property: "color" | "backgroundColor"): Promise<string> {
  return page.evaluate(
    ([name, prop]) => {
      const probe = document.createElement("span");
      probe.style.setProperty(prop === "color" ? "color" : "background-color", `var(${name})`);
      document.body.append(probe);
      const resolved = getComputedStyle(probe)[prop === "color" ? "color" : "backgroundColor"];
      probe.remove();
      return resolved;
    },
    [token, property] as const,
  );
}
