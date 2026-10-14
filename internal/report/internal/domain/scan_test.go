package domain

import (
	"strings"
	"testing"

	"github.com/ghassan-ai-projects/streams-simulator/internal/journal"
)

const (
	observationLine = `{"stream":"observer","seq":2,"type":"observation","at_ns":10,"line":"x"}`
	framingLine     = `{"stream":"observer","seq":3,"type":"observation","at_ns":11,"framing":"preamble","line":"y"}`
	sessionLine     = `{"stream":"observer","seq":4,"type":"consumer.reported","at_ns":12,"quiesced_through_ns":1,"verdict_submitted":false}`
	stateLine       = `{"stream":"director","seq":2,"type":"state","at_ns":10,"entity_id":"a"}`
	commandLine     = `{"stream":"director","seq":3,"type":"command","at_ns":10,"command":{"seq":0,"at_ns":10,"op":"clock.advance","args":{"to_ns":10}}}`
)

func TestAnObserverScanKeepsEveryRecordButObservationsAndCountsTheDelivered(t *testing.T) {
	t.Parallel()
	scan := doc(openedEvaluate, observationLine, framingLine, advanced, observationLine).scan(ObserverJournal)
	if scan.Err != nil {
		t.Fatal(scan.Err)
	}
	if scan.Delivered != 2 {
		t.Errorf("delivered = %d, want the two observations that are not framing", scan.Delivered)
	}
	types := []string{}
	for _, record := range scan.Records {
		types = append(types, record.Type)
	}
	if got := strings.Join(types, " "); got != "run.opened clock.advanced" {
		t.Errorf("kept %q, want only the opening and the clock advance", got)
	}
}

func TestAScanKeepsTheFirstRecordWhateverItsType(t *testing.T) {
	t.Parallel()
	scan := doc(observationLine, openedEvaluate).scan(ObserverJournal)
	if len(scan.Records) != 2 || scan.Records[0].Type != journal.TypeObservation {
		t.Fatalf("records = %+v, want the leading observation kept so records[0] is the file's first", scan.Records)
	}
}

func TestADirectorScanKeepsOnlyWhatTheReportReads(t *testing.T) {
	t.Parallel()
	scan := doc(`{"stream":"director","seq":0,"type":"run.opened","at_ns":0,"run_id":"r","seed":1}`, stateLine, commandLine, stateLine).scan(DirectorJournal)
	if scan.Err != nil || len(scan.Records) != 2 || scan.Records[1].Type != journal.TypeCommand {
		t.Fatalf("records = %+v (%v), want the opening and the command", scan.Records, scan.Err)
	}
}

func TestAScanDigestsMatchTheWholeBufferFunctions(t *testing.T) {
	t.Parallel()
	d := doc(openedEvaluate, observationLine, sessionLine, advanced)
	scan := d.scan(ObserverJournal)
	if scan.ReplayDigest != journal.Digest(d.Raw) {
		t.Errorf("replay digest = %s, want the journal's own %s", scan.ReplayDigest, journal.Digest(d.Raw))
	}
	if again := d.scan(DigestOnly); again.ReplayDigest != "" || again.RawDigest != scan.RawDigest {
		t.Errorf("digest-only scan = %+v, want a raw digest and no replay digest", again)
	}
}

func TestAScanStopsDecodingAtTheFirstBadRecordAndNamesItsLine(t *testing.T) {
	t.Parallel()
	scan := doc(openedEvaluate, `not json`, advanced).scan(ObserverJournal)
	if scan.Err == nil || !strings.Contains(scan.Err.Error(), "journal: line 2") {
		t.Fatalf("err = %v, want the second line named", scan.Err)
	}
	if len(scan.Records) != 1 {
		t.Errorf("records = %d, want those before the bad line", len(scan.Records))
	}
}

func TestALedgerScanTalliesRowsAndNamesABadLine(t *testing.T) {
	t.Parallel()
	good := doc(`{"delivered":true,"delivery_reason":"ok"}`, ``, `{"delivered":false,"delivery_reason":"late"}`).scan(LedgerRows)
	if good.Err != nil || good.Ledger.Rows != 2 || good.Ledger.Delivered != 1 || good.Ledger.DroppedByReason["late"] != 1 {
		t.Fatalf("scan = %+v (%v), want two rows, one delivered, one dropped late", good.Ledger, good.Err)
	}
	bad := doc(`{"delivered":true}`, `{`).scan(LedgerRows)
	if bad.Err == nil || !strings.Contains(bad.Err.Error(), "ledger.jsonl line 2") {
		t.Fatalf("err = %v, want the second line named", bad.Err)
	}
}

func TestATornTailIsReportedAndStillDigested(t *testing.T) {
	t.Parallel()
	torn := Document{Raw: []byte(openedEvaluate + "\n" + `{"stream"`), Present: true}.scan(ObserverJournal)
	if !torn.Torn || len(torn.Records) != 1 {
		t.Fatalf("scan = %+v, want a torn tail and the one committed record", torn)
	}
}
