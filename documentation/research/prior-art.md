# Prior-art summary

The project draws from deterministic simulation and industrial stream concepts without pretending to implement their full standards.

## Adopted ideas

- explicit event time and observed time;
- data-defined channels and entities;
- deterministic replay from versioned inputs;
- declared delivery faults and terminal delivery evidence;
- oracle separation from the system under test;
- scenario validity checks against trivial baselines.

## Deliberately not adopted

The simulator is not an implementation of FMI, Sparkplug B, ISO 13374, ISA-18.2, or EEMUA 191. Those sources inform vocabulary and design questions; they do not define the `streamsim` contracts.

## Source

The source comparison and citations are maintained in [`docs/research/PRIOR_ART.md`](../../docs/research/PRIOR_ART.md). Update that record when adopting a new external standard or protocol.

## Next reads

- [Benchmark validity](benchmark-validity.md)
- [Contracts](../reference/contracts.md)
