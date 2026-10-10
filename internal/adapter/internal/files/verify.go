package files

import (
	"fmt"
	"os"

	"github.com/ghassan-ai-projects/streams-simulator/internal/adapter/internal/domain"
	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

// Verify proves the adapter file at adapterPath correct with the consumer
// absent. fixturePath is a native-event JSONL fixture ("" selects the
// embedded one); base is the directory the adapter's declared conformance
// paths resolve against.
func Verify(adapterPath, fixturePath, base string) (*domain.VerifyResult, error) {
	a, err := Load(adapterPath)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	fixture, err := loadFixture(fixturePath)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return domain.Verify(a, fixture, base, verificationFiles{})
}

func loadFixture(path string) ([]model.SimEvent, error) {
	if path == "" {
		return embeddedFixture()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("adapter: adapter: %w", err)
	}
	events, err := domain.DecodeFixtureRecords(raw)
	if err != nil {
		return nil, fmt.Errorf("adapter: %w", err)
	}
	return events, nil
}

func embeddedFixture() ([]model.SimEvent, error) {
	events, err := domain.FixtureEvents()
	if err != nil {
		return nil, fmt.Errorf("adapter: embedded fixture: %w", err)
	}
	return events, nil
}

// verificationFiles reads the schema and golden files an adapter declares.
type verificationFiles struct{}

func (verificationFiles) OutputSchema(path string) ([]byte, error) {
	return readFile(path)
}

func (verificationFiles) Golden(path string) ([]byte, error) {
	return readFile(path)
}

func readFile(path string) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("%w", err)
	}
	return raw, nil
}
