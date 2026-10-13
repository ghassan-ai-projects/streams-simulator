package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const directorOpened = `{"stream":"director","seq":0,"type":"run.opened","at_ns":0,"run_id":"r-1","seed":7}`

func matchingArtifact() *model.RunArtifact {
	return &model.RunArtifact{RunID: "r-1", SimVersion: "9", Seed: 7,
		Domain:  model.ArtifactRef{ID: "d", Version: "1", Digest: "sha256:aa"},
		Adapter: model.ArtifactRef{ID: "n", Version: "1", Digest: "sha256:bb"}}
}

func TestReconcileKeepsARunJsonThatCarriesTheIdentityOfTheJournals(t *testing.T) {
	t.Parallel()
	folder := Reconcile(Folder{Observer: doc(openedEvaluate), Artifact: matchingArtifact()})
	if folder.Artifact == nil || folder.Mismatch != "" {
		t.Fatalf("a matching run.json was dropped: %q", folder.Mismatch)
	}
	alone := Reconcile(Folder{Artifact: matchingArtifact()})
	if alone.Artifact == nil {
		t.Fatal("a run.json with no journal to contradict it must be kept")
	}
}

func TestReconcileDropsARunJsonOfAnotherRunNamingTheFieldsThatDiffer(t *testing.T) {
	t.Parallel()
	cases := map[string]func(*model.RunArtifact){
		"run_id":      func(a *model.RunArtifact) { a.RunID = "r-9" },
		"sim_version": func(a *model.RunArtifact) { a.SimVersion = "8" },
		"domain":      func(a *model.RunArtifact) { a.Domain.Digest = "sha256:zz" },
		"adapter":     func(a *model.RunArtifact) { a.Adapter.Version = "2" },
	}
	for field, edit := range cases {
		t.Run(field, func(t *testing.T) {
			t.Parallel()
			art := matchingArtifact()
			edit(art)
			folder := Reconcile(Folder{Observer: doc(openedEvaluate), Artifact: art})
			if folder.Artifact != nil || !strings.Contains(folder.Mismatch, field+" differ") || strings.Contains(folder.Mismatch, "r-9") {
				t.Fatalf("mismatch %q (artifact kept: %v), want %s named and no value", folder.Mismatch, folder.Artifact != nil, field)
			}
			if TruthAllowed(folder) {
				t.Fatal("truth must not be allowed once run.json is foreign")
			}
		})
	}
}

func TestReconcileCannotCheckARunJsonAgainstAJournalWithoutItsOpening(t *testing.T) {
	t.Parallel()
	folder := Reconcile(Folder{Observer: doc(advanced), Artifact: matchingArtifact()})
	if folder.Artifact != nil || !strings.Contains(folder.Mismatch, "no run.opened record") {
		t.Fatalf("mismatch %q, want the missing run.opened named", folder.Mismatch)
	}
}

func TestReconcileSealedBlanksADirectorSideOfAnotherRun(t *testing.T) {
	t.Parallel()
	other := strings.NewReplacer(`"r-1"`, `"r-9"`).Replace(directorOpened)
	reseeded := strings.NewReplacer(`"seed":7`, `"seed":8`).Replace(directorOpened)
	for name, director := range map[string]string{"run_id": other, "seed": reseeded} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			folder := ReconcileSealed(Folder{Observer: doc(openedEvaluate), Artifact: matchingArtifact(), Director: doc(director), Ledger: doc(`{}`)})
			if folder.Director.Present || folder.Ledger.Present || !strings.Contains(folder.Mismatch, name+" differ") {
				t.Fatalf("director side kept (%v, %v), mismatch %q", folder.Director.Present, folder.Ledger.Present, folder.Mismatch)
			}
		})
	}
}

func TestReconcileSealedKeepsAMatchingDirectorSideAndChecksAnExploreRunByTheObserver(t *testing.T) {
	t.Parallel()
	same := ReconcileSealed(Folder{Observer: doc(openedEvaluate), Artifact: matchingArtifact(), Director: doc(directorOpened)})
	if !same.Director.Present || same.Mismatch != "" {
		t.Fatalf("a matching director stream was dropped: %q", same.Mismatch)
	}
	explore := ReconcileSealed(Folder{Observer: doc(openedExplore), Director: doc(directorOpened)})
	if explore.Director.Present || !strings.Contains(explore.Mismatch, "run_id differ") {
		t.Fatalf("explore run r-2 with a director stream of r-1: %q", explore.Mismatch)
	}
	for _, folder := range []Folder{{Mismatch: "set"}, {Observer: doc(openedEvaluate)}, {Observer: doc(openedEvaluate), Director: doc(advanced)}} {
		if got := ReconcileSealed(folder); got.Mismatch != folder.Mismatch {
			t.Fatalf("nothing to check, yet mismatch became %q", got.Mismatch)
		}
	}
}

func TestAssembleOfAForeignFolderNamesWhyItHasNoTruth(t *testing.T) {
	t.Parallel()
	folder := Reconcile(Folder{Observer: doc(openedEvaluate), Artifact: &model.RunArtifact{RunID: "r-9"}, Ledger: doc(`{}`)})
	report, err := Assemble(folder, nil)
	if err != nil {
		t.Fatal(err)
	}
	if report.Truth != nil || report.Run.Ended || report.Run.RunID != "r-1" || report.Run.IncompleteReason != report.TruthUnavailable ||
		!strings.Contains(report.TruthUnavailable, "run_id") {
		t.Fatalf("report %+v, unavailable %q", report.Run, report.TruthUnavailable)
	}
	if text := string(Markdown(report)); !strings.Contains(text, "Truth unavailable: run.json belongs to another run: ") || !strings.Contains(text, "run\\_id") {
		t.Fatalf("markdown lacks the reason:\n%s", text)
	}
}

func TestAssembleDropsATornLastLineOfEachStreamAndNamesIt(t *testing.T) {
	t.Parallel()
	cut := func(d Document) Document { return Document{Raw: d.Raw[:len(d.Raw)-3], Present: true} }
	ledger := doc(`{"delivery_id":1,"delivered":true,"delivery_reason":"delivered"}`, `{"delivery_id":2,"deliv`)
	report, err := Assemble(Folder{Observer: cut(doc(openedExplore, advanced)), Director: cut(doc(directorOpened, directorOpened)), Ledger: Document{Raw: ledger.Raw[:len(ledger.Raw)-1], Present: true}}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if got := strings.Join(report.TornTail, ","); got != "journal.jsonl,sealed/journal.jsonl,ledger.jsonl" || *report.Counts.LedgerRows != 1 {
		t.Fatalf("torn %q, ledger rows %v", got, *report.Counts.LedgerRows)
	}
	if text := string(Markdown(report)); !strings.Contains(text, "## Torn tail\n\n- journal.jsonl: the last line was cut short and dropped.") {
		t.Fatalf("markdown lacks the torn tail:\n%s", text)
	}
}

func TestObservedCountsLeaveOutAdapterFraming(t *testing.T) {
	t.Parallel()
	line := `{"stream":"observer","seq":%d,"type":"observation","at_ns":1,"line":"x"%s}`
	records := []string{openedExplore, strings.Replace(strings.Replace(line, "%d", "1", 1), "%s", `,"framing":"preamble"`, 1),
		strings.Replace(strings.Replace(line, "%d", "2", 1), "%s", "", 1)}
	report, err := Assemble(Folder{Observer: doc(records...)}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if *report.Counts.Observed != 1 {
		t.Fatalf("observed = %d, want only the delivered line", *report.Counts.Observed)
	}
}

func TestACodeSpanHoldingBackticksAtItsEdgesIsPaddedAndFenced(t *testing.T) {
	t.Parallel()
	if got := escapeCell(code("`a``")); got != "``` `a`` ```" {
		t.Fatalf("code span = %q", got)
	}
}
