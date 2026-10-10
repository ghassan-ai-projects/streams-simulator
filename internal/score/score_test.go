package score_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
	"github.com/ghassan-ai-projects/streams-simulator/internal/score"
)

func label() *model.GroundTruthRecord {
	return &model.GroundTruthRecord{ScenarioID: "d/0001", Domain: "d", Label: "f", ExpectedEpisode: true}
}

func TestScoreRefusesARunWithoutASubmittedVerdict(t *testing.T) {
	t.Parallel()
	_, err := score.Score(score.Evidence{RunID: "r-1"}, label())
	if err == nil || !strings.Contains(err.Error(), "no verdict submitted for run r-1") {
		t.Fatalf("err = %v", err)
	}
}

func TestScoreCarriesRunIdentityAndEvidenceFlagsOntoTheCard(t *testing.T) {
	t.Parallel()
	gt := label()
	card, err := score.Score(score.Evidence{
		RunID: "r-1", Verdict: &model.Verdict{}, Ledger: []model.LedgerRecord{},
		Reproducible: true, Unblinded: true,
	}, gt)
	if err != nil {
		t.Fatal(err)
	}
	if card.RunID != "r-1" || card.Domain != "d" || card.ScenarioID != "d/0001" || card.GroundTruth != gt {
		t.Fatalf("card identity = %+v", card)
	}
	if !card.Reproducible || !card.Unblinded || card.Bundle == "" {
		t.Fatalf("card flags = %+v", card)
	}
}

func TestScoreRefusesAMissingLabelAndToleratesAMissingDomain(t *testing.T) {
	t.Parallel()
	ev := score.Evidence{RunID: "r-1", Verdict: &model.Verdict{}, Ledger: []model.LedgerRecord{}}
	if _, err := score.Score(ev, nil); !errors.Is(err, score.ErrNoLabel) {
		t.Fatalf("nil label: %v", err)
	}
	gt := label()
	gt.ExpectedEffector = "start_aerator"
	if _, err := score.Score(ev, gt); err != nil {
		t.Fatalf("a run without a domain scores without loop recovery levels: %v", err)
	}
}

func TestOnlineAndOfflineAgreeOnTheMetricsTheyShare(t *testing.T) {
	t.Parallel()
	gt := label()
	verdict := &model.Verdict{}
	ledger := []model.LedgerRecord{}
	online, err := score.Score(score.Evidence{RunID: "r-1", Verdict: verdict, Ledger: ledger}, gt)
	if err != nil {
		t.Fatal(err)
	}
	offline := score.Offline(verdict, gt, ledger, nil, nil)
	if !reflect.DeepEqual(online.Consumer, offline.Consumer) || !reflect.DeepEqual(online.Judgment, offline.Judgment) {
		t.Fatalf("shared metrics differ:\nonline  %+v %+v\noffline %+v %+v", online.Consumer, online.Judgment, offline.Consumer, offline.Judgment)
	}
	if online.Bundle != offline.Bundle {
		t.Fatalf("bundle online %q offline %q", online.Bundle, offline.Bundle)
	}
}
