# Functional Specification

Streams Simulator — Visual Environment for Testing AI Agents in Simulated Physical Worlds

Product concept v1.0

Functional requirements

Product vision

Build a simulated physical world, connect an AI agent, observe its behavior, and discover how it fails before deploying it into the real world.

## 1. Product overview

Streams Simulator is an interactive environment where users can create simulated physical systems, connect AI agents, and evaluate their behavior under normal, abnormal, and dangerous conditions.

Users should be able to build entire environments visually, without programming.

The product brings together three capabilities:

- Streams Simulator: Creates the simulated physical world.
- Agentic Stream: Connects the world and the agent through continuous streams of observations and actions.
- Tamoz or other agents: Observe, decide, and act upon the simulated world.

The product is not simply a simulation tool. Its primary purpose is to help users answer:

Can I trust this AI agent to operate in this environment?

## 2. Core user experience

The user opens Streams Simulator and is presented with a visual canvas.

Factory Temperature Control

Simulation running

00:04:32 elapsed

1× speed

0 failures

The user can watch sensor values change, observe the agent's decisions, inject a fault, and see whether the agent responds correctly.

## 3. Main functional modules

3.1 Visual World Builder

Users create environments by placing objects onto a canvas and connecting them.

Examples: rooms, machines, sensors, pumps, tanks, motors, heaters, valves, robots and controllers.

Objects have configurable properties, behaviors and physical relationships.

3.2 Simulation Control

Users start, pause, resume, stop and reset a simulation. They can change simulation speed, advance one step at a time, and save or restore a scenario.

The environment continues to evolve according to its simulated behavior.

3.3 Agent Connection

Users connect Tamoz or another compatible AI agent, select which information it can observe, and specify which actions it is permitted to perform.

The agent participates in the simulated environment as an actual controller, not simply as a chatbot.

3.4 Failure and Disturbance Studio

Users introduce failures such as sensor inaccuracies, delayed messages, network outages, equipment malfunction, missing information, or unexpected human interference.

Failures may be scheduled, triggered by conditions or introduced manually.

3.5 Observation and Evaluation

Users monitor the physical state, incoming observations, agent decisions and resulting actions.

They can define expected behavior, establish safety boundaries, and determine whether the agent passed or failed each test.

3.6 Replay and Comparison

Users replay earlier runs, inspect the sequence of decisions, compare two agents against the same scenario, and evaluate whether a change improved performance.

Replays must distinguish faithfully recorded execution from fresh agent execution.

## 4. Example use case: Industrial cooling system

[technotrans: Lebensmittelindustrie](https://images.openai.com/static-rsc-4/64Oeh0dkUY5UU6WhDuqD4jN78voVPO6VHXmsOEkZaaqODgg9qlekLGA6NIaETEssPqKmXot_6lQa3iIHpvs0AtuHMx9Mfnrw1Ymtq12WOOaLlMlGGMoAlegevudx-Q01TfvQhUD7J514MHnKEp-UJlb6FmCnVKkn1MukUVGR_m8?purpose=inline)

[SCADA คืออะไร? ระบบควบคุม + เก็บข้อมูลโรงงาน ฉบับช่างไทย — ENT Group Blog](https://images.openai.com/static-rsc-4/71qWp1I1_m88QzLybhs2cxIoHYuvKunqltObkvx8_9OA1t4tmoQsMFx7t8VFBZGiU4GEEiLzYZeivN82BMSpky0iPvFDcaBrlZNC7uV0owjYBf92u0QISupBEKOh7hHS-pnFs5DhhtM82l3KCfd30ZhGnRoEqaSNWHtJjGHOXP8?purpose=inline)

[Reliance SCADA at TMA Ice Age](https://images.openai.com/static-rsc-4/NxyMH8xQLA0kVfxZTJ2RWDw3sSEm0x58yD1xrk0nnlceX4IDZhh8U0O2KqbKV06cMq8nBP1Y2wloLDVKH1723psIrS_MMCdrsYliBNMeCwhy64Ol1d5YfX64Z-HWFCSH0f9dLZTmSAn_haiC7Za2gKkEM_G2MMkm4u7CVloIlAs?purpose=inline)

10

A factory contains a machine whose temperature must remain between 20°C and 45°C.

A cooling system can reduce the temperature, but continuous operation consumes electricity and increases equipment wear.

An AI agent is responsible for maintaining safe temperatures while minimizing unnecessary cooling.

### User journey

1. Build the environment. Add a factory room, machine, temperature sensor and cooling unit.
2. Define behavior. The machine generates heat, cooling reduces temperature, and external conditions influence the system.
3. Connect an agent. Give it access to temperature observations and permission to control cooling power.
4. Define success. The agent must maintain safe temperatures, respond to emergencies and avoid excessive energy consumption.
5. Run the simulation. Observe the agent while the machine operates normally.
6. Inject failure. Make the temperature sensor report 30°C while the actual temperature is 55°C.
7. Evaluate behavior. Does the agent recognize inconsistent information, request verification, use another sensor, or make an unsafe decision?
8. Replay and compare. Modify the agent's configuration and repeat the scenario under comparable conditions.

The key distinction is that the simulated world maintains its own ground truth. The agent only knows what its permitted observations reveal.

## 5. Scenario Library

The product should provide predefined scenarios so users can start experimenting immediately.

[Ultrasonic Level Sensor](https://images.openai.com/static-rsc-4/otIdjDZ47i_xlo7Fg-QfYmoLdj4gwzRTbQEXebS7_qvYJybL6e7Njptwg7OfZRf_txIVCOLXi9ujyrUUqLL-lBFGbcwyFVStjVLdcGcjLXFuFw4E_ydqjYokDz3zQAOefoPQYskCEVTisDU75CMTgGr5reOWmqDEp3IibPaqyCo?purpose=inline)

Water Tank Management

Control pumps and valves to maintain safe water levels. Test incorrect sensor readings, leaks, pump failures and delayed responses.

[Välkommen till den Nordiska marknaden](https://images.openai.com/static-rsc-4/WcaniSHSVi8KDqFvHdYXottsHCI81jtAgbKYekHfCtPr3JYTGTDVAV0IGOQebx3HjMtI-79FqcVgYDFNJLPYavbi2gzea36U00hYsctyT3yAfVtzjkeEssqUqjBkYqgL2dUwDPHbTusT8niUoefKX7w0_ATMqNYf9iY5M6TRBNw?purpose=inline)

Production Line

Coordinate machines and conveyor belts. Test blockages, overheating, equipment failure and production overload.

[Grand Central Hotel, Northern Ireland | Daikin](https://images.openai.com/static-rsc-4/JeZvINbtngYp5-V2LjmbZnOuMjYqjpIXc9aSt2CWGqKdxmAfvjWaShUneaPI-c9URjZVBUwXltVgPi01a9lhlDmQM4zUocBwqmx84e9G9OcOQ2RkAJqX8LvGeLGe4KN8XPBV2OYfccNONPcWE6CI1pHsSmmKDgKRAd5xGegjgZY?purpose=inline)

Smart Building

Maintain comfortable indoor conditions while reducing energy consumption. Test conflicting room requirements and sensor outages.

[A Perfect Fit](https://images.openai.com/static-rsc-4/9ZEhMD1JUY7ff0rYjK8VHOYf_TCJzAscg8MMr7Lcwh6Z_Uyr9y8ERmn8LbDii92w9_kgo99q5xyRbWpdo-sb5GtxEbiIDWYmHa_cKx9vzDvqLPn6boBlPjosGU1TjdMRZTg95l8K3r48h2bgnPgenCDNwK1qoic7xreEAZjRr28?purpose=inline)

Warehouse Automation

Manage robotic movements and deliveries. Test blocked paths, conflicting instructions and unexpected obstacles.

## 6. Agent testing and evaluation

The product should evaluate an agent across five dimensions.

| Dimension        | What is measured                                                 |
| ---------------- | ---------------------------------------------------------------- |
| Safety           | Did it violate any defined safety boundary?                      |
| Reliability      | Did it complete its objective despite disturbances?              |
| Decision quality | Were its actions appropriate given the information available?    |
| Efficiency       | How much time, energy and other resources did it consume?        |
| Recovery         | How effectively did it return to normal operation after failure? |

Each scenario produces an evaluation report.

Simulation Evaluation

Example report · Cooling System · Run #024

Failed

## 14/16

Tests passed

## 2

Safety violations

## 6.4s

Avg. response

Critical finding

The agent increased cooling power when the reported temperature was already inconsistent with independent sensor measurements. A required verification action was skipped.

Illustrative result, not an actual agent test.

The report should make it clear what happened, why it matters, and how to reproduce it. Where an agent's reasoning cannot be established, the report should distinguish observed behavior from possible explanations.

## 7. Minimum viable product (MVP)

I would avoid building a complete industrial simulation platform initially.

The first release should prove the product's central promise: an AI agent can be tested against realistic changing conditions, and unsafe behavior can be discovered reproducibly.

Feature scope

8 of 12 selected

Visual canvas with objects and connections

MVP

One complete physical scenario: temperature control

MVP

Basic simulation controls and live readings

MVP

Agentic Stream integration

MVP

Tamoz reference agent connection

MVP

Manual fault injection

MVP

Safety assertions and run report

MVP

Record and replay simulation runs

MVP

Multiple physical scenario templates

Later

Side-by-side agent comparison

Later

Collaborative scenario editing

Later

3D environment visualization

Later

Develop selected scope&#x20;

## 8. MVP acceptance criteria

The first version is successful when a user can:

1. Open a ready-made industrial cooling scenario and understand its components.
2. Change environmental conditions through the visual interface.
3. Connect a compatible agent via Agentic Stream.
4. Start a run and see the simulated world react to the agent's actions.
5. Inject a faulty temperature reading during execution.
6. Observe whether the agent takes a safe or unsafe action.
7. Identify the exact moment when a safety boundary was violated.
8. Replay the original run and reproduce the same recorded sequence of events.
9. Export a report containing the scenario, timeline, actions and test outcomes.

## 9. What the product should not become

At least initially, Streams Simulator should not try to become:

- A general-purpose workflow automation platform competing with n8n.
- A complete industrial engineering tool replacing specialist simulation packages.
- A 3D digital-twin rendering platform.
- An agent development framework duplicating Tamoz.
- A replacement for real-world hardware validation and operational safety certification.

Its purpose is to provide a controlled environment for evaluating agent behavior.

## 10. Product success

Beyond completing the MVP, I would track whether the product is actually useful.

| Product metric           | Initial target to validate                                              |
| ------------------------ | ----------------------------------------------------------------------- |
| Time to first agent test | Under 10 minutes with a template                                        |
| Scenario creation        | No coding for a standard scenario                                       |
| Fault reproducibility    | 100% for recorded deterministic simulation behavior                     |
| Test explainability      | Every violation linked to a timestamp and relevant observations/actions |
| Agent compatibility      | Tamoz plus at least one independent agent                               |
| User validation          | At least 3 external teams successfully evaluate their agents            |

These are proposed product goals, not established performance measurements.

## 11. Long-term product vision

The product could eventually support a progression from simulation to operational readiness:

Design a world

Define environment, equipment, behavior and constraints

Connect an agent

Expose observations and permitted actions

Stress-test behavior

Explore disturbances, unexpected conditions and failures

Evaluate readiness

Generate evidence against defined behavioral and safety requirements

Prepare for real-world deployment

Support further hardware testing, human approval and operational validation

The product positioning I would use:

> Streams Simulator is a visual testing environment where developers build simulated physical worlds, connect autonomous AI agents, and discover failures before those agents interact with real equipment.

One product decision matters more than the visual editor itself: the simulated physical world must behave independently of the agent. It needs hidden state, evolving conditions, constraints, disturbances and consequences. Otherwise, you're only testing an agent against scripted examples, not against an environment that can challenge its decisions.

That independence, combined with Agentic Stream and Tamoz, is where I see the real value of your three projects.

------------------

Exactly. And this is one of the most important requirements for Streams Simulator.

It should simulate not only what Tamoz can observe, but also what happens when Tamoz takes an action.

The simulator must understand the consequences of those actions and update the simulated world accordingly.

## The complete interaction loop

1\. Simulated Physical World

Machines, sensors, environment, physical state

Observations / Commands

2\. Agentic Stream

Streaming communication and action delivery

Context / Decisions

3\. Tamoz

Reasoning, tool use, decisions and approvals

The loop repeats as the world changes

## Example: Tamoz controls a cooling system

Suppose Tamoz has access to four actions:

| Action Tamoz can take    | What Streams Simulator does                                    |
| ------------------------ | -------------------------------------------------------------- |
| `start_cooling`          | Activates the cooling system and gradually reduces temperature |
| `set_cooling_power(80%)` | Changes cooling capacity, power consumption and equipment load |
| `stop_cooling`           | Stops cooling, allowing temperature to rise naturally          |
| `emergency_shutdown`     | Shuts down the machine and changes its heat generation         |

The key idea is that the simulator does not simply return "action succeeded."

It models what actually happens afterward.

For example, Tamoz sends `set_cooling_power(80%)`. The simulated cooling unit might be defective and operate at only 20% capacity. The temperature continues to rise despite Tamoz's command.

Now Tamoz must observe that discrepancy and decide what to do next.

## Three important requirements

1\. Actions must have consequences. An action can change temperature, pressure, energy consumption, machine status or other world properties. The effects may be gradual rather than immediate.

2\. Actions can fail or behave unexpectedly. A command may be rejected, delayed, partially executed or successfully delivered while the physical equipment fails to respond as expected.

3\. Tamoz must not automatically know the actual world state. It should only receive observations from available sensors and other permitted sources. The simulator keeps the actual state separately.

## A critical feature: Action Simulation

I would make Action Simulation a first-class functional module of Streams Simulator.

Every simulated device would expose:

- Available actions: What an agent can request.
- Preconditions: When each action is allowed.
- Effects: How the simulated world changes.
- Execution duration: How long the effects take.
- Failure modes: What can go wrong.
- Feedback: What the agent can observe about the action's outcome.

This would allow the user to design an environment visually and automatically expose compatible actions to Tamoz through Agentic Stream.

### One additional feature I strongly recommend

Counterfactual Action Testing.

Imagine Tamoz decides to open a pressure valve. You could pause the simulation, branch the scenario, and compare:

- What happens if Tamoz opens the valve?
- What happens if it keeps the valve closed?
- What happens if the valve is stuck?
- What happens if the command arrives 30 seconds late?

This turns Streams Simulator into more than a testing environment. It becomes a way to explore the consequences of AI decisions under controlled conditions.

My recommendation: Make observations, actions, and physical consequences the three fundamental concepts of Streams Simulator. The visual canvas is simply how users construct and explore them.


------------------

The projects we have as next to this project are :

/Users/ghassan/my-projects/tamoz
/Users/ghassan/my-projects/agentic-stream
