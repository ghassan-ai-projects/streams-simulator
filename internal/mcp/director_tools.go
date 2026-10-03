package mcp

import (
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

// NewDirectorServer builds the director-role server.
func NewDirectorServer(d *Director) *mcp.Server {
	s := mcp.NewServer(implementation, nil)
	registerCatalogTools(s, d)
	registerWorldTools(s, d)
	registerInjectionTools(s, d)
	registerRunTools(s, d)
	registerTruthTools(s, d)
	registerDirectorResources(s, d)
	return s
}

func registerCatalogTools(s *mcp.Server, d *Director) {
	addTool(s, toolDef{name: "sim.catalog.list", description: "List the installed domains with their property vectors and stresses.", schema: toolSchema("sim.catalog.list"), handler: d.handleCatalogList})
	addTool(s, toolDef{name: "sim.catalog.describe", description: "Describe a domain: channels, faults, effectors, profiles, fidelity tiers.", schema: toolSchema("sim.catalog.describe"), handler: d.handleCatalogDescribe})
	addTool(s, toolDef{name: "sim.catalog.coverage", description: "The axis-coverage matrix; which axes are thin.", schema: toolSchema("sim.catalog.coverage"), handler: d.handleCatalogCoverage})
	addTool(s, toolDef{name: "sim.adapter.list", description: "Installed output adapters.", schema: toolSchema("sim.adapter.list"), handler: d.handleAdapterList})
}

func registerWorldTools(s *mcp.Server, d *Director) {
	addTool(s, toolDef{name: "sim.world.create", description: "Create a world: domain, seed, sink, adapter, time mode.", schema: toolSchema("sim.world.create"), handler: d.handleWorldCreate})
	addTool(s, toolDef{name: "sim.world.describe", description: "Describe a world: config, digest, clock, emitted count.", schema: toolSchema("sim.world.describe"), handler: d.handleWorldDescribe})
	addTool(s, toolDef{name: "sim.world.destroy", description: "Destroy a world; final counts and run artifact.", schema: toolSchema("sim.world.destroy"), handler: d.handleWorldDestroy})
	addTool(s, toolDef{name: "sim.clock.advance", description: "Advance by a relative nanosecond delta or to an absolute epoch nanosecond; await_consumer blocks on quiescence.", schema: toolSchema("sim.clock.advance"), handler: d.handleClockAdvance})
	addTool(s, toolDef{name: "sim.clock.state", description: "The clock, next scheduled event, pending effects.", schema: toolSchema("sim.clock.state"), handler: d.handleClockState})
}

func registerInjectionTools(s *mcp.Server, d *Director) {
	addTool(s, toolDef{name: "sim.fault.inject", description: "Inject a world fault into an entity.", schema: toolSchema("sim.fault.inject"), handler: d.handleFaultInject})
	addTool(s, toolDef{name: "sim.fault.clear", description: "Clear a fault by id.", schema: toolSchema("sim.fault.clear"), handler: d.handleFaultClear})
	addTool(s, toolDef{name: "sim.fault.list", description: "Active faults. Director only.", schema: toolSchema("sim.fault.list"), handler: d.handleFaultList})
	addTool(s, toolDef{name: "sim.perturb.apply", description: "Apply a delivery perturbation.", schema: toolSchema("sim.perturb.apply"), handler: d.handlePerturbApply})
	addTool(s, toolDef{name: "sim.perturb.clear", description: "Clear a perturbation by id.", schema: toolSchema("sim.perturb.clear"), handler: d.handlePerturbClear})
	addTool(s, toolDef{name: "sim.env.inject", description: "Inject an environment fault against a configured target.", schema: toolSchema("sim.env.inject"), handler: d.handleEnvInject})
}

func registerRunTools(s *mcp.Server, d *Director) {
	addTool(s, toolDef{name: "sim.run.begin", description: "Open a run; truth is sealed.", schema: toolSchema("sim.run.begin"), handler: d.handleRunBegin})
	addTool(s, toolDef{name: "sim.run.end", description: "Close the run; writes the run artifact.", schema: toolSchema("sim.run.end"), handler: d.handleRunEnd})
}

func registerTruthTools(s *mcp.Server, d *Director) {
	addTool(s, toolDef{name: "sim.truth.seal", description: "Install and seal a director-only ground-truth record before run.begin.", schema: toolSchema("sim.truth.seal"), handler: d.handleTruthSeal})
	addTool(s, toolDef{name: "sim.run.verify", description: "Verify a run artifact reproduces.", schema: toolSchema("sim.run.verify"), handler: d.handleRunVerify})
	addTool(s, toolDef{name: "sim.truth.reveal", description: "Reveal sealed truth; unblind stamps the run permanently.", schema: toolSchema("sim.truth.reveal"), handler: d.handleTruthReveal})
	addTool(s, toolDef{name: "sim.truth.seal_status", description: "Sealing state of a run.", schema: toolSchema("sim.truth.seal_status"), handler: d.handleTruthSealStatus})
	addTool(s, toolDef{name: "sim.score", description: "Score a run from its submitted verdict, sealed truth, ledger and effector log.", schema: toolSchema("sim.score"), handler: d.handleScore})
	addTool(s, toolDef{name: "sim.scenario.audit", description: "Trivial-baseline audit of one injection.", schema: toolSchema("sim.scenario.audit"), handler: d.handleScenarioAudit})
	addTool(s, toolDef{name: "sim.entity.retire", description: "Retire an entity.", schema: toolSchema("sim.entity.retire"), handler: d.handleEntityRetire})

}

func registerDirectorResources(s *mcp.Server, d *Director) {
	s.AddResource(&mcp.Resource{URI: "sim://catalog", Name: "Domain catalog", MIMEType: "application/json"}, d.readCatalogResource)
	s.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "sim://domains/{id}/spec", Name: "Domain spec"}, d.readDomainResource)
}
