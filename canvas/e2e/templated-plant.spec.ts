import { copyFileSync } from "node:fs";
import { join } from "node:path";
import { expect, test } from "./support/test";
import { repository } from "./support/workspace";

const plants = [
  { id: "aquaculture-pond", count: "8 instances", instance: "site-a/pond-3", kind: "pond" },
  { id: "host-system-health", count: "12 instances", instance: "rack-a/server-12", kind: "server" },
];

for (const { id, count, instance, kind } of plants) {
  test(`${id}, declared by an id template, shows the instances the binary names`, async ({ page, lab }) => {
    const file = `${id}.domain.json`;
    copyFileSync(join(repository, "domains", file), join(lab.workspace, "domains", file));
    await page.goto(`/#/plant?domain=domains%2F${id}.domain.json`);
    await expect(page.getByRole("region", { name: `Plant ${id}` }).getByText(new RegExp(`^${count}`))).toBeVisible();
    await page
      .getByRole("navigation", { name: "Outline" })
      .getByRole("button", { name: `${instance} ${kind}`, exact: true })
      .click();
    const inspector = page.getByRole("complementary", { name: "Inspector" });
    await expect(inspector.getByRole("heading", { name: instance })).toBeVisible();
    await expect(inspector.getByRole("region", { name: "Channels" })).not.toBeEmpty();
  });
}
