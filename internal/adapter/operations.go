package adapter

import (
	layer "github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/files"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

func Load(path string) (*model.Adapter, error) {
	return files.Load(path)
}

func LoadDir(dir string) (map[string]*model.Adapter, error) {
	return files.LoadDir(dir)
}

func LoadBytes(raw []byte, src string) (*model.Adapter, error) {
	return layer.LoadBytes(raw, src)
}

func Verify(adapterPath, fixturePath, base string) (*VerifyResult, error) {
	return files.Verify(adapterPath, fixturePath, base)
}

func VerifyFailure(res *VerifyResult) (message string, diverged bool) {
	return layer.VerifyFailure(res)
}

func (e *Engine) Begin() ([]string, error) {
	return e.engine.Begin()
}

func (e *Engine) RenderRecord(ev *model.SimEvent) (string, error) {
	return e.engine.RenderRecord(ev)
}

func (e *Engine) RenderStreamRecord(ev *model.SimEvent) (string, error) {
	return e.engine.RenderStreamRecord(ev)
}

func (e *Engine) End(worldEndNS int64) ([]string, error) {
	return e.engine.End(worldEndNS)
}

func (e *Engine) RenderRun(events []model.SimEvent, worldEndNS int64) ([]byte, error) {
	return e.engine.RenderRun(events, worldEndNS)
}

func Lookup(adapters map[string]*model.Adapter, id string) (*model.Adapter, error) {
	return layer.Lookup(adapters, id)
}
