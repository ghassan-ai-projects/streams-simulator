import AxeBuilder from "@axe-core/playwright";
import { expect, test } from "./support/test";

test.use({ target: "empty" });
test.setTimeout(120_000);

const title = "Belt conveyor, three drives with margin";

test("the belt conveyor opens with one counted wire per pair of instances and channels on its nodes", async ({
  page,
}) => {
  await page.goto("/");
  await page.getByRole("button", { name: new RegExp(`^Try it\\s*: ${title}$`) }).click();
  const panel = page.getByRole("complementary", { name: `Run ${title}` });
  await panel.getByRole("button", { name: "Close the run panel" }).click();
  const graph = page.getByRole("application", { name: "Plant graph" });
  const toggle = page.getByRole("button", { name: "Group links" });
  const wires = page.locator(".react-flow__edge");
  await expect(toggle).toHaveAttribute("aria-pressed", "true");
  await expect(wires).toHaveCount(4);
  await expect(graph.getByRole("img", { name: "D-1 and B-1, 11 links: 7 to B-1, 4 to D-1" })).toBeAttached();
  const drive = graph.getByRole("group", { name: /^D-1, conveyor drive/ });
  await expect(drive.getByText("winding")).toBeVisible();
  await expect(drive.getByText("°C")).toBeVisible();
  await expect(graph.getByText("drag out")).toHaveCount(0);
  const audit = await new AxeBuilder({ page }).analyze();
  expect(audit.violations).toEqual([]);

  await toggle.click();
  await expect(toggle).toHaveAttribute("aria-pressed", "false");
  await expect(wires).toHaveCount(35);
  await expect(drive.getByText("drag out")).toBeVisible();
  await expect(drive.getByText("winding")).toHaveCount(0);
});
