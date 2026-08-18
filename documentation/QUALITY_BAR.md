# Documentation Quality Bar

This is the acceptance standard for the Streams Simulator documentation. The documentation is complete only when a new contributor, an evaluator, and an operator can use it without relying on private context or tribal knowledge.

## Definition of world-class

The documentation must be:

- **Trustworthy** — every behavior claim, command, example, contract, and architectural statement is consistent with the repository and its tests.
- **Task-oriented** — readers can quickly find the path for evaluating the project, trying it, extending it, and contributing to it.
- **Layered** — the landing page is concise; deeper design, contracts, research, and operational details are available without making the first read overwhelming.
- **Reproducible** — examples use real commands and repository paths, state their prerequisites, and explain expected results.
- **Open-source ready** — licensing, contribution, conduct, security reporting, support expectations, compatibility, and release practices are discoverable and internally consistent.
- **Maintainable** — each fact has an obvious home, links are intentional, terminology is stable, and documentation can be checked as the code evolves.
- **Accessible** — prose is plain and precise, headings are meaningful, tables are used where they improve scanning, diagrams have textual explanations, and no essential information exists only in a diagram.

## Hard gates

The final documentation must pass every gate below. A single failed hard gate means the work is not done.

### 1. Entry and orientation

- The root README explains what the project is, why it exists, its trust model, and who should use it.
- The README provides a short, working quickstart and links to the documentation hub.
- The documentation hub explains the reading paths for users, contributors, maintainers, and researchers.
- The repository has no competing or ambiguous documentation home.

### 2. Repository truth

- Public commands, package paths, configuration names, file names, APIs, protocol versions, and examples are verified against the current repository.
- Documentation does not describe planned behavior as implemented behavior.
- Design documents clearly label status: implemented, partially implemented, proposed, or historical.
- Every changed behavior has either a test, a documented limitation, or an explicit reason why verification is not applicable.

### 3. Core model and guarantees

- The simulator’s purpose, boundaries, deterministic inputs, outputs, and replay model are explained.
- The world, event lifecycle, perturbation/delivery model, adapters, sinks, MCP roles, ledger, truth, scoring, and effectors are covered at the right level.
- The nine non-negotiable design guarantees are visible and traceable to their documentation and verification evidence.
- Failure modes and the distinction between transport misses and reasoning misses are explained.

### 4. Usable workflows

- A reader can run the smallest useful example from a clean checkout.
- A reader can configure a domain, adapter, sink, seed, and command flow using documented inputs.
- A reader can inspect results, ground truth, delivery evidence, and verdicts.
- A reader can reproduce a failure from its artifact or command log.
- A contributor can build, test, lint, and validate changes using documented commands.

### 5. Reference completeness

- Every user-facing command, flag, configuration field, JSON contract, MCP tool/role, adapter/sink mode, and error category has a discoverable reference.
- Normative requirements are distinguishable from explanatory guidance and examples.
- Versioning, compatibility, canonicalization, determinism, and security boundaries are documented where they affect interoperability or trust.
- Examples are minimal, valid, and copyable; large examples link to maintained fixtures where appropriate.

### 6. Open-source readiness

- The repository’s license and third-party dependency posture are discoverable.
- Contribution, code review, testing expectations, issue/bug reporting, security reporting, and code of conduct guidance are present or explicitly linked to authoritative files.
- The support boundary and compatibility expectations are clear.
- Release, change-log, and deprecation expectations are documented, even if the project is pre-release.
- No private paths, credentials, internal-only references, or machine-specific assumptions appear in published guidance.

### 7. Architecture and diagrams

- The dependency direction and major data flows are explained in prose and represented with small, focused diagrams where a visual materially improves comprehension.
- Each diagram has a title, a stated scope, readable labels, and a nearby textual explanation.
- Diagrams do not introduce concepts, names, or flows that the prose and code cannot support.
- Mermaid or another repository-friendly format is used so diagrams remain reviewable and editable.

### 8. Editorial quality

- Headings describe reader questions or stable concepts.
- Prose uses consistent terms, active voice, concrete nouns, and short paragraphs.
- Acronyms are expanded at first use; terms with precise meanings are defined once and reused consistently.
- Links are relative where possible, descriptive, and free of known dead targets.
- Markdown renders cleanly; code fences declare languages; lists and tables are formatted consistently.

### 9. Maintenance and verification

- The documentation hub identifies authoritative sources and avoids duplicated facts.
- Each public fact has exactly one authority; duplicate schemas, fixtures, examples, and generated copies are explicitly identified and verified.
- Public pages do not link to stale or superseded design claims without a status label and an authoritative replacement.
- Release claims distinguish working-tree state from committed release-manifest state.
- The archive README clearly identifies internal-only material and historical evidence.
- A lightweight documentation review/check process is documented and runnable.
- The final review checks links, Markdown structure, examples, commands, diagrams, and code alignment.
- Review findings are recorded and resolved; remaining gaps are explicit, prioritized, and not disguised as completeness.

## Scored dimensions

After all hard gates pass, score each dimension from 0 to 3:

| Score | Meaning |
| --- | --- |
| 0 | Missing or materially misleading |
| 1 | Present but shallow, fragmented, or difficult to use |
| 2 | Useful and mostly complete, with minor gaps |
| 3 | Clear, complete, verified, and easy to maintain |

The target is **3** in every dimension:

1. Orientation and information architecture
2. Quickstart and first successful run
3. Conceptual model
4. Architecture and data flow
5. Configuration and protocol reference
6. Determinism, truth, scoring, and failure analysis
7. Extension and contribution workflows
8. Operations, security, and troubleshooting
9. Open-source project governance
10. Writing, accessibility, links, and diagrams
11. Code/test alignment
12. Maintenance and change detection

Every substantive page ends with a short `Next reads` section. Security reporting, support, contribution, conduct, license, changelog, and release paths are either actionable and linked or explicitly marked as missing.

## Evidence required at handoff

The final review must include:

- a documentation inventory and disposition of the previous documentation files;
- a map from major documented claims to code, tests, or explicit design status;
- a claim-to-source matrix covering commands, versions, domain and adapter counts, MCP tools, artifact files, contracts, and release status;
- the commands used to validate Markdown, links, examples, and repository checks;
- a short list of deliberate limitations and future documentation work;
- review results from an independent completeness/gap pass and a code-alignment pass.

## Stop condition

The work may be called complete only when all hard gates pass, every scored dimension is at least 2, no unresolved correctness or navigation issue remains, and the remaining limitations are named in the documentation hub or a maintained backlog.

## Next reads

- [Documentation hub](README.md)
- [Claim-to-source matrix](governance/claims-matrix.md)
- [Current limitations](limitations.md)
