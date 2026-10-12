package adapter

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Load reads and validates an adapter file. The validation covers the
// output-adapter schema plus adapter-specific cross-checks (transform
// arity, source names, identity preservation).
func Load(path string) (*model.Adapter, error) {
	return files.Load(path)
}

// LoadDir loads every adapter file in dir (an empty dir selects "adapters"),
// keyed by adapter id. A directory that holds no adapter is refused.
func LoadDir(dir string) (map[string]*model.Adapter, error) {
	return files.LoadDir(dir)
}

// LoadBytes validates an adapter from an in-memory source document. It is
// used by artifact replay so a run can carry its own adapter definition.
func LoadBytes(raw []byte, src string) (*model.Adapter, error) {
	return layer.LoadBytes(raw, src)
}

// Verify proves the adapter file at adapterPath correct with the consumer
// absent: it renders the conformance fixture, validates every record against
// the adapter's declared output schema and byte-compares the output to the
// committed golden. fixturePath names a native-event JSONL fixture ("" selects
// the embedded one); base is the directory the adapter's declared paths
// resolve against.
func Verify(adapterPath, fixturePath, base string) (*VerifyResult, error) {
	return files.Verify(adapterPath, fixturePath, base)
}

// VerifyFailure decides whether a verification failed: the message is empty
// for one that passed, and diverged is true when a declared check ran and
// failed. An adapter that declares no schema and no golden has nothing to
// compare, so it never counts as verified.
func VerifyFailure(res *VerifyResult) (message string, diverged bool) {
	return layer.VerifyFailure(res)
}

// Begin starts a streaming session and returns its framing and preamble
// records. A session must not be mixed with RenderRun on one Engine.
func (e *Engine) Begin() ([]string, error) {
	return e.engine.Begin()
}

// RenderRecord renders a single event through the record template, returning
// "" when the when-guard excludes it.
func (e *Engine) RenderRecord(ev *model.SimEvent) (string, error) {
	return e.engine.RenderRecord(ev)
}

// RenderStreamRecord renders one event in an active streaming session,
// applying array separators when the adapter declares json-array encoding.
func (e *Engine) RenderStreamRecord(ev *model.SimEvent) (string, error) {
	return e.engine.RenderStreamRecord(ev)
}

// End closes the streaming session and returns its postamble and framing
// records; worldEndNS is bound before postamble evaluation.
func (e *Engine) End(worldEndNS int64) ([]string, error) {
	return e.engine.End(worldEndNS)
}

// RenderRun renders preamble, every event and postamble into the adapter's
// encoding; worldEndNS is the run's final clock.
func (e *Engine) RenderRun(events []model.SimEvent, worldEndNS int64) ([]byte, error) {
	return e.engine.RenderRun(events, worldEndNS)
}

// Lookup returns the adapter with the given id from a loaded set, or
// ErrUnknownAdapter.
func Lookup(adapters map[string]*model.Adapter, id string) (*model.Adapter, error) {
	return layer.Lookup(adapters, id)
}
