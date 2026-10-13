package domain

import (
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/testsupport"
)

func TestSummarizeCountsWhatTheDomainDeclares(t *testing.T) {
	t.Parallel()
	compiled, err := Load(testsupport.Domain("aquaculture-pond"))
	if err != nil {
		t.Fatal(err)
	}
	got := Summarize(compiled)
	want := Summary{Channels: len(compiled.ChannelNames()), Digest: compiled.Digest, Effectors: compiled.EffectorCount(), Faults: compiled.FaultCount(), ID: "aquaculture-pond"}
	if got != want || got.Channels == 0 || got.Faults == 0 || got.Effectors == 0 {
		t.Fatalf("Summarize = %+v, want %+v with every count positive", got, want)
	}
}
