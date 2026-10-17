import { check } from "./architecture/check.ts";
import { report } from "./architecture/report.ts";

process.exitCode = report("canvas css", check("architecture.json", "css"), "css");
