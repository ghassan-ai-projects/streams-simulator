package architecture

import (
	"go/ast"
	"slices"
	"strings"
	"testing"
)

// ioImports are the imports that give a file access to the outside world.
var ioImports = []string{
	"os", "os/exec", "os/signal", "net", "net/http",
	"crypto/rand", "math/rand", "math/rand/v2", "syscall", "io/ioutil",
}

// clockCalls are the time functions that read or wait on the wall clock.
var clockCalls = []string{
	"Now", "Since", "Until", "Sleep", "After", "AfterFunc", "NewTimer", "NewTicker", "Tick",
}

// ioEdge declares why a production file may touch files, sockets, processes,
// entropy or the wall clock. debt names the round that moves the use to an
// edge package; foundation and core packages may only hold debt.
type ioEdge struct {
	uses []string
	why  string
	debt string
}

// ioEdges is the complete inventory of I/O and clock sites (STANDARD M3, M4).
// A file that uses a capability not listed here fails the gate, and a listed
// capability a file no longer uses fails it too, so the table only shrinks.
var ioEdges = map[string]ioEdge{
	"cmd/streamsim/main.go":                                 {uses: []string{"import:os"}, why: "process exit"},
	"cmd/streamsim/function_length.go":                      {uses: []string{"call:filepath.WalkDir", "import:os"}, why: "development tool, build-ignored"},
	"internal/kernel/internal/files/files.go":               {uses: []string{"import:os"}, why: "the one file-system edge: reads, owner-only directories, atomic owner-only writes"},
	"internal/kernel/internal/files/log.go":                 {uses: []string{"import:os"}, why: "owner-only append logs: buffered, flushed at boundaries, synced on close"},
	"internal/kernel/internal/sockets/sockets.go":           {uses: []string{"import:net", "import:os", "import:syscall"}, why: "unix sockets published owner-only by rename; dialed only when owner-only"},
	"internal/kernel/internal/process/process.go":           {uses: []string{"import:os", "import:os/signal", "import:syscall"}, why: "a context that ends on interrupt or SIGTERM"},
	"internal/kernel/operations.go":                         {uses: []string{"import:net"}, why: "the facade names the listener and connection types it returns"},
	"internal/wall/wall_clock.go":                           {uses: []string{"call:time.Now"}, why: "the declared wall-clock seam"},
	"internal/sink/internal/httppush/httppush.go":           {uses: []string{"import:net/http"}, why: "http-push sink: POST each line"},
	"internal/director/internal/capability/capability.go":   {uses: []string{"import:crypto/rand"}, why: "operator capability token entropy"},
	"internal/director/internal/runner/runner.go":           {uses: []string{"call:time.NewTimer", "call:time.Now", "call:time.Until", "go-statement"}, why: "the wall pacer that makes a job's simulated time follow wall time at its rate, and the goroutine a job runs on"},
	"internal/mcp/internal/transport/transport.go":          {uses: []string{"go-statement", "import:net", "import:net/http"}, why: "operator role bound over streamable HTTP and the director session served until its transport closes"},
	"internal/canvasbridge/internal/entropy/entropy.go":     {uses: []string{"import:crypto/rand"}, why: "canvas pairing codes and session ids"},
	"internal/canvasbridge/internal/files/files.go":         {uses: []string{"call:filepath.Abs", "import:os"}, why: "the canvas workspace opened as one os.Root; reads and listings inside it"},
	"internal/canvasbridge/internal/files/write.go":         {uses: []string{"import:os"}, why: "atomic conditional writes inside the workspace root"},
	"internal/canvasbridge/internal/process/clock.go":       {uses: []string{"call:time.NewTicker"}, why: "the poll interval of journal following and workspace watching"},
	"internal/canvasbridge/internal/process/command.go":     {uses: []string{"import:os/exec"}, why: "streamsim commands the canvas actions run, the bridge's own executable"},
	"internal/canvasbridge/internal/process/director.go":    {uses: []string{"import:os/exec"}, why: "the director child the canvas drives over MCP on stdio"},
	"internal/canvasbridge/internal/process/lastwords.go":   {uses: []string{"call:time.NewTimer", "go-statement", "import:os", "import:os/exec"}, why: "the director child's stderr drained through a pipe so a crash is reported with its last words, bounded in wall time"},
	"internal/canvasbridge/internal/process/self.go":        {uses: []string{"import:os"}, why: "the bridge's own executable path"},
	"internal/canvasbridge/internal/process/control.go":     {uses: []string{"call:context.WithTimeout", "import:os"}, why: "a connected session's control socket: a private directory the bridge owns for it, and one request line sent to the node's owner-only socket with a bounded wait"},
	"internal/canvasbridge/internal/process/lasting.go":     {uses: []string{"call:time.NewTimer", "go-statement", "import:os", "import:os/exec", "import:syscall"}, why: "a connected session's streamsim child the bridge owns until it is stopped: started in its own process group, waited on, interrupted, and its group killed after a bounded wait"},
	"internal/canvasbridge/internal/transport/admission.go": {uses: []string{"import:net/http"}, why: "request admission and pairing over HTTP"},
	"internal/canvasbridge/internal/transport/events.go":    {uses: []string{"import:net/http"}, why: "Server-Sent Events of run journals and workspace changes"},
	"internal/canvasbridge/internal/transport/respond.go":   {uses: []string{"import:net/http"}, why: "JSON responses and request bodies"},
	"internal/canvasbridge/internal/transport/routes.go":    {uses: []string{"import:net/http"}, why: "the canvas API routes and the built UI"},
	"internal/canvasbridge/internal/transport/transport.go": {uses: []string{"go-statement", "import:net", "import:net/http", "import:os"}, why: "loopback listener served until the context ends; the UI folder opened as an os.Root"},

	"internal/run/internal/quiesce/quiesce.go":    {uses: []string{"call:time.NewTimer"}, why: "wall-clock deadline for await_consumer waits"},
	"internal/cli/internal/process/process.go":    {uses: []string{"call:time.Now", "import:os"}, why: "standard streams and wall clock of the running process"},
	"internal/testsupport/testsupport.go":         {uses: []string{"import:os"}, why: "repository root discovery and short socket directories for tests"},
	"internal/device/internal/uds/uds.go":         {uses: []string{"go-statement", "import:net"}, why: "unix socket listener, accept loop and session"},
	"internal/device/internal/uds/serve.go":       {uses: []string{"import:net", "import:os"}, why: "device socket served until the process is interrupted"},
	"internal/plantwire/internal/uds/link.go":     {uses: []string{"call:time.Now", "go-statement", "import:net"}, why: "plant wire node: dial the host's owner-only socket (owner checked), read its records, write with a deadline"},
	"internal/plantwire/internal/uds/await.go":    {uses: []string{"call:time.NewTimer", "call:time.Now", "call:time.Since"}, why: "wall-time cap of the plant wire barrier and the wait it measured; the wall pacer of a paced session, on the monotonic clock"},
	"internal/plantwire/internal/uds/listen.go":   {uses: []string{"call:time.After", "import:net", "import:os"}, why: "reference host: owner-only socket published by rename, one node session, a handshake deadline, delayed barrier answers"},
	"internal/plantwire/internal/uds/gateways.go": {uses: []string{"go-statement", "import:net", "import:os"}, why: "relay: the gateway's owner-only socket published by the relay and accepted until closed, one link per connection"},
	"internal/plantwire/internal/uds/control.go":  {uses: []string{"call:context.WithTimeout", "go-statement", "import:net", "import:os"}, why: "running node: the owner-only control socket an operator's controls arrive on, one goroutine per connection, each request bounded in wall time"},
	"internal/plantwire/internal/uds/ticks.go":    {uses: []string{"call:time.NewTicker"}, why: "relay and reference host: wall-clock ticks that count missed heartbeats, pace the redial of a lost host and pace the host's own heartbeats"},
	"internal/plantwire/internal/uds/sim.go":      {uses: []string{"call:time.NewTimer", "call:time.Now", "call:time.Since", "go-statement"}, why: "mixed relay: the in-process link to the simulated node, its wall-time barrier cap and the goroutine that runs the node's session"},
}

func TestIOStaysInDeclaredEdges(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	for _, file := range productionFiles(t) {
		used := ioUses(file)
		seen[file.path] = true
		declared := ioEdges[file.path].uses
		if !slices.Equal(used, declared) {
			t.Errorf("%s: I/O uses %v, declared %v: move the code to an edge or update ioEdges with a reason",
				file.path, used, declared)
		}
	}
	for path := range ioEdges {
		if !seen[path] {
			t.Errorf("stale ioEdges entry: %s is not a production file", path)
		}
	}
}

func TestPureKindsHoldOnlyScheduledIODebt(t *testing.T) {
	t.Parallel()
	for path, edge := range ioEdges {
		kind := infoOf(packageDirectory(path)).kind
		pure := kind == kindFoundation || kind == kindCore
		if pure && edge.debt == "" {
			t.Errorf("%s: %s package declares I/O without scheduled debt", path, kind)
		}
		if edge.debt != "" && !roundIsPlanned(t, edge.debt) {
			t.Errorf("%s: debt %q must name a round row in PLAN.md", path, edge.debt)
		}
		if edge.why == "" {
			t.Errorf("%s: ioEdges entry needs a reason", path)
		}
	}
}

// fsCalls are path/filepath functions that touch the file system or process
// working directory; contextClockCalls are context constructors that read the
// wall clock.
var (
	fsCalls           = []string{"Abs", "EvalSymlinks", "Glob", "Walk", "WalkDir"}
	contextClockCalls = []string{"WithTimeout", "WithDeadline"}
)

// ioUses lists the sorted I/O, clock and concurrency capabilities a file uses.
// Function values such as `Clock: time.Now` count as uses.
func ioUses(file productionFile) []string {
	uses := map[string]bool{}
	for _, path := range file.paths {
		if slices.Contains(ioImports, path) {
			uses["import:"+path] = true
		}
	}
	if file.imports["."] == "time" {
		uses["import:time(dot)"] = true
	}
	ast.Inspect(file.source, func(node ast.Node) bool {
		if use, ok := capabilityOf(file, node); ok {
			uses[use] = true
		}
		return true
	})
	list := make([]string, 0, len(uses))
	for use := range uses {
		list = append(list, use)
	}
	slices.Sort(list)
	if len(list) == 0 {
		return nil
	}
	return list
}

func capabilityOf(file productionFile, node ast.Node) (string, bool) {
	switch n := node.(type) {
	case *ast.GoStmt:
		return "go-statement", true
	case *ast.SelectorExpr:
		return selectorCapability(file, n)
	}
	return "", false
}

func selectorCapability(file productionFile, selector *ast.SelectorExpr) (string, bool) {
	ident, ok := selector.X.(*ast.Ident)
	if !ok {
		return "", false
	}
	name := selector.Sel.Name
	switch file.imports[ident.Name] {
	case "time":
		return "call:time." + name, slices.Contains(clockCalls, name)
	case "path/filepath":
		return "call:filepath." + name, slices.Contains(fsCalls, name)
	case "context":
		return "call:context." + name, slices.Contains(contextClockCalls, name)
	}
	return "", false
}

// roundIsPlanned reports whether PLAN.md has a status row for the round.
func roundIsPlanned(t *testing.T, round string) bool {
	t.Helper()
	plan, err := repositoryRoot(t).ReadFile("governance/MODULARITY_PLAN.md")
	if err != nil {
		t.Fatal(err)
	}
	return strings.Contains(string(plan), "| "+round+" |")
}
