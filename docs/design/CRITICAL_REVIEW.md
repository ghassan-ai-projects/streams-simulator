# Critical Review

Status: adversarial review of the design as committed at `d0fd5b6`
Date: 2026-08-11
Scope: the premises, not the completeness

[GAP_ANALYSIS.md](GAP_ANALYSIS.md) accepts the design's goals and asks whether it is
buildable. This document does the opposite: it grants that the design is buildable and
asks whether building it is the right move, and whether the thing it would produce
answers the question it claims to answer.

Eight findings. Two of them, if accepted, change what gets built.

---

## C-01 · The headline experiment is compromised, and not for the reason the design admits

This is the most important finding in the document.

The product hypothesis is:

> Continuous evidence can be reduced into a stable, versioned Situation such that a model
> receives better context than the raw evidence window and produces a materially more
> useful diagnosis.

The three-arm experiment tests it: deterministic-only, raw-window-dump,
situation-grounded. The simulator supplies the scenarios and the ground truth.

The existing evaluation design discloses one limitation: the physics are synthetic, so a
good score is not a field result. That disclosure is correct and insufficient. **The
binding problem is not that the process is simulated. It is that the feature extractor
and the data generator share an author.**

Concretely: the same person, in the same week, writes `aquaculture-pond.domain.json` —
which decides that `do_probe_fouling` ramps at 0.04/hour and perturbs exactly
`probe_fouling` — and then writes the Situation Model whose bands, windows and hysteresis
are supposed to summarize evidence from that world. The Situation arm is therefore fed a
summary hand-fitted to the exact generative process under test. The raw-window arm is fed
undigested data from the same process.

A win in that comparison establishes that **hand-tuned feature extraction, tuned against
the answer, beats no feature extraction**. That was never in doubt. It does not establish
that Situations-as-an-architecture help on a real, unmodelled process, which is the claim
the product rests on.

The effect is not small and it is not subtle. It is the single largest source of bias in
the experiment, larger than model nondeterminism, larger than scenario triviality, and it
is currently undisclosed.

**Mitigations, in descending order of strength:**

1. **Author the Situation Model blind.** The Situation Model author sees the channel
   declarations, the units, the enums and a nominal trace. They do not see the fault list,
   the observability models, the state variables, or the dynamics. This is a process
   control, it costs a person-week of coordination, and it converts the experiment from
   compromised to defensible.
2. **Hold out fault labels from the model author** even if the same person does both, with
   a time gap and a committed hash of the Situation Model before the fault list is read.
   Weaker, but auditable.
3. **Report a "tuning gap" metric**: score the Situation arm on faults the model author
   *did* see versus faults introduced afterward. If the gap is large, say so in the
   headline.

At minimum, disclose it in the same paragraph as the synthetic-physics caveat, with equal
prominence. As written, [GROUND_TRUTH_AND_SCORING §7](GROUND_TRUTH_AND_SCORING.md) gives
the weaker limitation a rule and the stronger one silence.

## C-02 · The catalog's best argument for itself is also an argument against building most of it

The README's strongest line is that the catalog is "already paying before a line of Go
exists" — writing it surfaced eight design questions about Agentic Stream, four of which
look like real gaps.

That is true, and it is worth noticing what it implies. **The value came from the
modelling exercise, not from the artifact.** Q1 (per-input lateness) did not require
`open-field-irrigation` to be built; it required someone to sit down and notice that a
satellite revisit and a soil probe cannot share a lateness allowance. Q3 (entity
retirement) did not require a running `k8s-cluster`; it required counting pod identities
per hour against a durable-Situation-per-entity design.

So the honest question is: what does *building* the remaining 22 domains buy that
*specifying* them did not?

The answer is real but narrower than the catalog implies:

- **Confirmation** that the predicted gaps are actual — worth something, but the four
  flagged as defects are defects on inspection, not hypotheses.
- **The unpredicted gaps** — the ones nobody thought of. This is the genuine argument,
  and it is unquantifiable in advance.
- **Statistical power** for the graded experiment — which needs *many scenarios in a few
  domains*, not few scenarios in many domains. Twenty-five domains actively hurts here by
  spreading a fixed scenario budget thinner.
- **Domain-generality proof** for the injection invariant — which three unrelated domains
  establish exactly as well as twenty-five. Agentic Stream already made this argument for
  itself and settled on two.

Combined with G-01 (each domain carries an unscheduled hand-authored Situation Model) and
G-08 (no cost model), the case for 25 built domains is weak. The case for 25 *specified*
domains, of which 4–6 are built, is strong.

**Recommendation**: keep the catalog at 25 as a design instrument and a roadmap. Build
Tier 1 only — `rotating-machinery` (byte-equivalence proof), `host-system-health`
(cheap, counters), `open-field-irrigation` (Q1), `k8s-cluster` (Q3), `aquaculture-pond`
(closed loop), `discrete-line-oee` (Q7, injection probe). Six. Everything else becomes a
domain that gets built when a specific question needs it, which is also how it should have
been framed to begin with.

## C-03 · The instrument now costs more than the system it measures, and nobody decided that

Agentic Stream V0.1 is 25 build days plus a 5-day evaluation phase, one engineer. The
simulator is 30 build days for M0–M5 — and G-01 argues that is understated by 50–130%
once Situation Models are counted, putting the honest figure at 45–70 days.

So the test instrument is **1.8–2.8× the system under test**, for a product whose core
hypothesis is not yet demonstrated.

That is not automatically wrong; instruments legitimately cost more than prototypes, and
some of this cost is already committed via V0.1's existing fault-model generator. But it
is a decision, and no document in this repository records anyone making it.

There is a second reading worth stating plainly. A deterministic, seeded, closed-loop,
25-domain world simulator with generative ground truth, sealed oracles and an
anti-benchmark-flaw methodology is **a product**. Arguably a more differentiated one than
a single-node Situation Runtime — the DST-for-agentic-systems space has no obvious
occupant, and "we can prove your agent doesn't hallucinate success" is a sharper pitch
than "we compute Situations."

The design does not consider this and should not decide it. But it should say out loud
that a 45–70 day investment in an instrument, before the thing being measured has
validated its hypothesis, is either a testing decision or a pivot, and the two want
different scopes.

## C-04 · "MCP-first" is, after the transport ruling, mostly convenience — and the design still presents it as load-bearing

The ruling was right. But it is worth stating what MCP is *left* doing.

Post-ruling, MCP provides: a control API for a harness, and an effector RPC. Both are
things a plain HTTP+JSON API or a Go interface would do at least as well — with less
protocol surface, no version negotiation, no capability-token plumbing, and, critically,
**no two-server oracle-leak problem**, because a harness-only HTTP API on a loopback port
is not something the system under test is configured with by accident.

The two genuine benefits that remain:

1. **Conversational driving.** A human in Claude Code, or Tamoz, can operate the simulator
   without a CLI. This is real and it is nice.
2. **Exploration collapses into a command log.** This is the design's most-praised
   property — and it is a property of the **command log**, not of MCP. A CLI or HTTP API
   that appends to the same log gets the identical benefit.

So the honest trade is: *one ergonomic affordance, in exchange for a protocol dependency,
two servers, capability tokens, prefix-indistinguishability testing, a configuration lint
in Tamoz, and an entire class of leak that would not otherwise exist.*

That may still be the right trade — the affordance is genuinely valuable for a research
lab, and the effector-over-MCP path is needed regardless because it is how the stream's
adapter will talk to a simulated plant. But the design currently asserts the value rather
than pricing it, and a reviewer should force the price onto the page.

**Minimum action**: state the trade explicitly in the README, and note that the
oracle-leak control surface (two servers, tokens, differential testing, the Tamoz lint) is
a cost *caused by* the MCP choice, not an inherent property of simulating a plant.

## C-05 · `succeeds_but_no_effect` — the design's best test — was made easy by its own domain spec

I called this the highest-value test in the design, and it is. Then
`aquaculture-pond.domain.json` gave the pond an independent `pond.aerator_current`
channel, described in the spec as "the channel that distinguishes
`succeeds_but_no_effect` from a genuine recovery."

If the Situation Model reads aerator current, then catching a no-effect acknowledgement is
a **deterministic two-signal comparison**: the command said start, the current says off.
No reasoning, no outcome reconciliation, no learning loop. The test passes and proves
nothing about the property it was built to prove.

The valuable version of this test is the one where the *only* evidence of no-effect is
that **the outcome did not occur** — oxygen did not recover over the expected time
constant — which forces the reconciliation window, the outcome definition, and the
counterfactual expectation to all be right.

**Action**: `succeeds_but_no_effect` needs two variants, scored separately.
`confirmed_no_effect` (a confirming channel exists) is a mechanism test and belongs in
CI. `silent_no_effect` (no confirming channel; only the absent outcome betrays it) is the
loop test, and it is the one whose false-success rate must be zero.

## C-06 · The determinism engineering is TigerBeetle-grade for a workload that is not

Cross-architecture byte-identical floats, a `simdet` build tag, resolution quantization, a
named PRNG substream tree, a map-iteration lint. This is serious deterministic-simulation
rigor, and it was correctly sourced.

But look at what it is for. TigerBeetle needs cross-architecture determinism because it
runs millennia of simulated time across thousands of cores hunting rare interleavings; the
seed must reproduce on any machine in the fleet. This simulator's stated workload is ~100
scenarios per domain, evaluated largely against **model outputs, which are nondeterministic
by the design's own admission**.

The determinism that carries the value is: *a specific failing scenario reproduces from
its artifact*. That needs the seed, the command log, and a pinned platform. It does not
need `linux/amd64` and `darwin/arm64` to agree bit-for-bit on a float.

Keep: the run artifact, the PRNG tree (cheap, and it prevents a genuinely nasty bug class),
the single-goroutine world core. Downgrade: cross-architecture identity from an M1
release-blocking gate to a nice-to-have, with CI pinned to one architecture and a periodic
cross-check. That is real time back in M1, which G-01 says is where the schedule is
already wrong.

The counter-argument is that resolution quantization — the mechanism proposed to fix
cross-arch divergence — is independently justified because real instruments have a
resolution. Agreed; keep the quantization, drop the gate.

## C-07 · The blinding controls protect against the model cheating, not against us

The oracle-leak work is good: two servers, compile-time separation, prefix
indistinguishability, suspicious-detection monitoring, a scrambled-label control. All of
it defends the boundary between the **system under test** and the truth.

None of it defends the boundary between the **experimenters** and the result. And the
experimenters are the ones with a stake in the outcome, the ability to regenerate suites
with a different seed, the ability to adjust `first_observable_snr` from 3.0 to 2.5, and
the ability to decide that a domain "wasn't ready" after seeing its numbers.

The prior art the design leans on knows this. Pre-registration exists in the V0.1
evaluation design precisely for this reason, and this document reintroduces a dozen new
tunable knobs — SNR thresholds, negative-class fraction, prevalence weights, trivial-audit
accuracy cutoff, suite composition — with no statement of when they are frozen.

**Action**: the suite composition, the SNR constants, the audit cutoff, and the domain set
must be committed and hashed **before the first graded run**, exactly as the decision rule
already is. Changing any of them afterward invalidates the run and requires re-running, not
re-tuning. This costs nothing to adopt now and is impossible to adopt credibly later.

## C-08 · The simulator has no named customer, and is therefore sized for none

Three plausible customers, three incompatible optimal designs:

| If the customer is… | Then what matters | And what is waste |
|---|---|---|
| **Us, improving Agentic Stream** | 4–6 domains, mechanism metrics, the 8 design questions, the closed loop | the graded suite, blinding, 19 domains, the anti-flaw methodology |
| **Prospects, as a demo** | `http-live + scaled`, a UI, 2 photogenic domains, the closed loop | determinism, ground truth, scoring, blinding, 23 domains |
| **A public benchmark** | the anti-flaw methodology, blinding, 25 domains, statistical power | the closed loop, effectors, infrastructure faults |

The design serves all three, which is why it is 45–70 days. Notice that **the closed loop
is valuable to two of the three and the graded suite to one**, and that the intersection of
all three is roughly: determinism, 4–6 domains, the closed loop, and mechanism metrics —
about M0 through M4, which is 24 of the 30 planned days and contains every one of the
design's stated non-negotiables except the trivial-baseline audit.

That is not a coincidence, and it is the recommendation: **build M0–M4 with six domains,
then decide who the customer is with a working instrument in hand.** M5 is where the
customer question becomes unavoidable, and it is also where six of the twelve gaps first
bite. Deciding before M5 rather than during it is nearly free.

---

## What I would change before writing code

Ranked by (value ÷ cost), highest first:

1. **Freeze the tunable constants and commit them** (C-07). One hour. Impossible to do
   credibly later.
2. **Split `succeeds_but_no_effect` into confirmed and silent variants** (C-05). One hour.
   Rescues the design's best test.
3. **Cut to six built domains; keep 25 specified** (C-02). A paragraph. Removes 20–40 days
   and improves the experiment's statistical power rather than harming it.
4. **Fix the observability model to accept detector expressions** (G-03). A day of schema
   and solver work now; a rewrite of 13 domains later.
5. **Design the quiescence barrier** (G-02). A day. Without it the closed loop is not
   reproducible, which negates the entire determinism architecture in the one mode that
   matters most.
6. **Blind the Situation Model authoring** (C-01). A process change, roughly a person-week
   of coordination. Converts the headline experiment from compromised to defensible.
7. **Downgrade cross-architecture determinism from a gate** (C-06). Frees several days in
   the milestone that is already understated.
8. **Write the scorecard contract** (G-04). Two days. It is the only artifact anyone
   outside the project will read.

Items 1–3 cost about half a day combined and are the difference between an instrument that
can be trusted and one that cannot. Item 6 is the one that decides whether the eventual
report means anything.

## What holds up

Stated plainly, because a review that only attacks is not a review:

- **The transport analysis is correct and the ruling improved it.** The reasoning from
  MCP's delivery semantics to "not the evidence plane" is sound, and the consequences
  (zero blocking stream changes, a smaller plant surface) are real, not rationalized.
- **The three-injection-surface separation** — world faults, transport perturbations,
  infrastructure faults — is the right decomposition and is not obvious. Most simulators
  conflate at least two.
- **The prefix-indistinguishability test** is a genuinely good idea. Differential testing
  against a checklist-based control is the correct instinct for side channels.
- **Importing the anomaly-benchmark critique** was the highest-value research decision in
  the document, and the three-onset-timestamp mechanism is the right response even though
  G-03 shows the current implementation of it is wrong.
- **The domain catalog as a design-question generator** worked. Eight questions, four
  probable defects, at the cost of a document. That is an excellent return, and C-02 does
  not dispute it — it observes that the return was already collected.

---

## Disposition after the decoupling revision

Written against `d0fd5b6`. The owner's decoupling ruling — the simulator is an independent
product, consumers know nothing of it and it knows nothing of them — resolved two findings
and materially changed two more.

| # | Status |
|---|---|
| **C-01** Compromised headline experiment | **Improved, not resolved.** The separation is now structural on the simulator's side: it has no idea what a consumer's configuration looks like, and cannot be tuned toward one. The remaining risk is a single author writing both a domain's fault model and the consumer configuration meant to summarize it. Carried as decision D-3. |
| **C-02** 25 domains is 19 too many | **Accepted.** Six built, 25 specified. |
| **C-03** Instrument costs more than the system it measures | **RESOLVED as stated, and reframed.** The comparison no longer applies: this is not an instrument for one system, it is a product whose first customer happens to be nearby. The plan dropped from ~74 to ~58 days by removing the ~23 spent building another product. The question C-03 actually raised — is this a testing decision or a pivot? — is now decision D-2, where it belongs. |
| **C-04** MCP is mostly convenience, and the oracle-leak surface is a cost it causes | **Partially resolved.** One server with two roles instead of two servers removes half the surface; four generic operator tools remove the rest of the domain knowledge. The trade is still real and still worth stating: conversational driving in exchange for a protocol dependency and a role-separation control. But the cost is now much smaller than when it was two servers and per-effector tools. |
| **C-05** `succeeds_but_no_effect` made easy by its own domain spec | **Accepted and fixed in the contract.** Two modes, scored separately; `silent_no_effect` is the one whose rate must be zero. |
| **C-06** Determinism is TigerBeetle-grade for a workload that is not | **Partially accepted.** Resolution quantization stays — it is physically honest and independently justified. Cross-architecture byte identity is no longer a release-blocking gate. |
| **C-07** Blinding protects against the model cheating, not against us | **Open, scheduled in S0.** Freeze and hash the tunable constants before the first graded run. |
| **C-08** No named customer | **Open**, as decision D-2 — and the decoupling makes "a conformance benchmark for stream processors generally" a materially more plausible answer than it was, which is worth weighing rather than assuming away. |

The observation in C-02 that has aged best, and which the decoupling sharpens: the
catalog's value was collected by writing it. That remains the strongest argument for
building six domains and the strongest argument against building nineteen more.
