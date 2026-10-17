import { check } from "./architecture/check.ts";
import { report } from "./architecture/report.ts";
import { runFixtures } from "./architecture/fixtures.ts";

const registerPath = "architecture.json";
const fixturesRoot = "architecture-fixtures";

const fixtureFailures = runFixtures(fixturesRoot);
for (const failure of fixtureFailures) process.stdout.write(`  fixture ${failure}\n`);
const code = report("canvas architecture", check(registerPath, "architecture"), "architecture");
process.exitCode = fixtureFailures.length === 0 ? code : 1;
