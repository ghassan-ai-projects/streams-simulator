package domain

import (
	"encoding/json"
	"errors"
	"io"
	"strings"
	"testing"
)

func laterBirthLine(t *testing.T, bdSeq uint64, kinds []any) string {
	t.Helper()
	birth := stampedBirth(t, "hardware", kinds)
	birth["bd_seq"] = bdSeq
	raw, err := json.Marshal(birth)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func TestTheFirstOfSeveralBirthsIsTakenAndTheLaterOnesAreReportedWithTheirContent(t *testing.T) {
	t.Parallel()
	first := birthLine(t, nil)
	same := laterBirthLine(t, 4, birthKinds())
	other := laterBirthLine(t, 5, birthKinds()[1:])
	captured, err := parse(first + "\n" + wrapped("observer", same) + "\n" + other + "\n")
	if err != nil {
		t.Fatal(err)
	}
	if captured.BDSeq != 3 || captured.Taken != (Provenance{Line: 1, BDSeq: 3}) || captured.BirthsSeen != 3 || len(captured.Kinds) != 2 {
		t.Fatalf("capture = %+v, want the first birth (bd_seq 3, line 1) of three, whatever follows it", captured)
	}
	want := []LaterBirth{{Line: 2, BDSeq: 4, SameContent: true}, {Line: 3, BDSeq: 5, SameContent: false}}
	if len(captured.LaterBirths) != 2 || captured.LaterBirths[0] != want[0] || captured.LaterBirths[1] != want[1] {
		t.Fatalf("later births = %+v, want %+v", captured.LaterBirths, want)
	}
	if got := keywords(captured.Issues); len(got) != 1 || got[0] != "rebirth_differs" || !strings.Contains(captured.Issues[0].Message, "line 3 (bd_seq 5)") {
		t.Fatalf("issues = %+v, want one rebirth_differs naming the birth at line 3", captured.Issues)
	}
}

func TestLaterBirthsOfTheSameContentRaiseNoIssue(t *testing.T) {
	t.Parallel()
	captured, err := parse(birthLine(t, nil) + "\n" + laterBirthLine(t, 4, birthKinds()) + "\n" + laterBirthLine(t, 5, birthKinds()))
	if err != nil || captured.BirthsSeen != 3 || len(captured.LaterBirths) != 2 || len(captured.Issues) != 0 {
		t.Fatalf("ParseCapture = %+v, %v, want two later births of the same content and no issue", captured, err)
	}
	for _, later := range captured.LaterBirths {
		if !later.SameContent {
			t.Errorf("later birth %+v, want the same content", later)
		}
	}
}

func TestAFirstBirthThatCannotBeReadIsReportedAndTheNextOneIsTaken(t *testing.T) {
	t.Parallel()
	torn := `{"type":"NBIRTH","bd_seq":2,"kinds":[`
	captured, err := parse(torn + "\n" + birthLine(t, nil) + "\n" + torn)
	if err != nil {
		t.Fatal(err)
	}
	if captured.Taken.Line != 2 || captured.BirthsSeen != 1 {
		t.Fatalf("capture taken from line %d with %d births seen, want the first readable birth, line 2", captured.Taken.Line, captured.BirthsSeen)
	}
	if got := keywords(captured.Issues); len(got) != 2 || got[0] != "birth_line_unreadable" || got[1] != "birth_line_unreadable" ||
		!strings.Contains(captured.Issues[0].Message, "line 1") || !strings.Contains(captured.Issues[1].Message, "line 3") {
		t.Fatalf("issues = %+v, want line 1 and line 3 reported as unreadable births", captured.Issues)
	}
}

func TestAFirstBirthThatBreaksTheContractIsRefusedNotSkippedForALaterOne(t *testing.T) {
	t.Parallel()
	broken := birthLine(t, func(b map[string]any) { b["environment"] = "lab" })
	_, err := parse(broken + "\n" + birthLine(t, nil))
	if invalid := birthViolationOf(t, err); invalid.Issues[0].Pointer != "/environment" {
		t.Fatalf("issues = %+v, want the first birth refused at /environment", invalid.Issues)
	}
}

func TestOnlyTheObserverStreamIsEverTaken(t *testing.T) {
	t.Parallel()
	bare := birthLine(t, nil)
	withStream := func(stream string) string {
		return birthLine(t, func(b map[string]any) { b["stream"] = stream })
	}
	cases := []struct {
		name, input string
		cause       error
	}{
		{"a wrapped record of the director", wrapped("director", bare), ErrSealedStream},
		{"a bare record labeled director", withStream("director"), ErrSealedStream},
		{"a bare record labeled with another stream", withStream("replay"), ErrNotObserver},
		{"a wrapped record without a stream", `{"seq":2,"type":"wire","direction":"node","record":` + bare + `}`, ErrNotObserver},
		{"a wrapped record of another stream", wrapped("replay", bare), ErrNotObserver},
		{"a wrapped record whose stream is no text", `{"stream":7,"type":"wire","direction":"node","record":` + bare + `}`, ErrNotObserver},
		{"another case of the key beside the director", `{"stream":"director","Stream":"observer","type":"wire","direction":"node","record":` + bare + `}`, ErrSealedStream},
		{"another case of the key alone", `{"Stream":"observer","type":"wire","direction":"node","record":` + bare + `}`, ErrNotObserver},
		{"a director record after an observer one", wrapped("observer", bare) + "\n" + wrapped("director", bare), ErrSealedStream},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			_, err := parse(tc.input)
			_ = violation(t, err, tc.cause)
		})
	}
}

func TestAMemberNamedTwiceInARecordIsNeverResolvedByTheLastOne(t *testing.T) {
	t.Parallel()
	bare := birthLine(t, nil)
	cases := map[string]string{
		"the stream named twice":                     `{"stream":"director","stream":"observer","type":"wire","direction":"node","record":` + bare + `}`,
		"the stream named twice, in the other order": `{"stream":"observer","stream":"director","type":"wire","direction":"node","record":` + bare + `}`,
		"the type named twice":                       `{"type":"NDEATH","type":"wire","direction":"node","record":` + bare + `}`,
		"a member of the record twice":               strings.Replace(bare, `"environment":"hardware"`, `"environment":"hardware","environment":"simulated"`, 1),
	}
	for name, input := range cases {
		_, err := parse(input)
		_ = violation(t, err, ErrNoBirth)
		if !errors.Is(err, ErrDuplicateMember) {
			t.Errorf("%s: error = %v, want the duplicate member named as the cause", name, err)
		}
	}
}

func TestARecordOverTheLineLimitIsSkippedWithoutHoldingItAndTheNextBirthIsTaken(t *testing.T) {
	t.Parallel()
	huge := `{"type":"NBIRTH","pad":"` + strings.Repeat(" ", MaxLineBytes) + `"}`
	captured, err := parse(huge + "\n" + birthLine(t, nil))
	if err != nil || captured.Taken.Line != 2 || len(captured.Issues) != 1 || captured.Issues[0].Keyword != "birth_line_unreadable" ||
		!strings.Contains(captured.Issues[0].Message, "line 1") {
		t.Fatalf("ParseCapture = %+v, %v, want line 1 reported as unreadable and the birth of line 2 taken", captured, err)
	}
	_, err = parse(huge)
	_ = violation(t, err, ErrNoBirth)
	if !errors.Is(err, ErrTooLarge) {
		t.Fatalf("error = %v, want the line limit named as the cause", err)
	}
}

func TestANBIRTHBetweenTheLineLimitAndTheContractLimitIsRefusedByTheContract(t *testing.T) {
	t.Parallel()
	raw := `{"type":"NBIRTH","pad":"` + strings.Repeat(" ", MaxBirthBytes) + `"}`
	_, err := parse(raw)
	invalid := violation(t, err, ErrTooLarge)
	if invalid.Issues[0].Keyword != "birth_too_large" || invalid.Issues[0].Located {
		t.Fatalf("issues = %+v, want one unlocated birth_too_large issue", invalid.Issues)
	}
}

type repeating struct {
	line string
	left int
	at   int
}

func (r *repeating) Read(p []byte) (int, error) {
	if r.left <= 0 {
		return 0, io.EOF
	}
	p = p[:min(len(p), r.left)]
	for i := range p {
		p[i] = r.line[r.at%len(r.line)]
		r.at++
	}
	r.left -= len(p)
	return len(p), nil
}

func filler(lines int) io.Reader {
	return &repeating{line: strings.Repeat("x", 99) + "\n", left: lines * 100}
}

const linesBeyondTheScanLimit = MaxScanBytes/100 + 10000

func TestAJournalBeyondTheScanLimitStillYieldsAnEarlyBirthAndSaysTheScanWasCut(t *testing.T) {
	t.Parallel()
	journal := io.MultiReader(strings.NewReader(birthLine(t, nil)+"\n"), filler(linesBeyondTheScanLimit))
	captured, err := ParseCapture(journal)
	if err != nil || len(captured.Kinds) != 2 || captured.Taken.Line != 1 {
		t.Fatalf("ParseCapture = %+v, %v, want the early birth taken from a journal larger than the scan limit", captured, err)
	}
	if got := keywords(captured.Issues); len(got) != 1 || got[0] != "scan_truncated" {
		t.Fatalf("issues = %+v, want scan_truncated: births_seen is a lower bound", captured.Issues)
	}
}

func TestAJournalWithinTheScanLimitIsReadToTheEndWithoutAnIssue(t *testing.T) {
	t.Parallel()
	journal := io.MultiReader(strings.NewReader(birthLine(t, nil)+"\n"), filler(150000), strings.NewReader(laterBirthLine(t, 4, birthKinds())+"\n"))
	captured, err := ParseCapture(journal)
	if err != nil || captured.BirthsSeen != 2 || len(captured.LaterBirths) != 1 || captured.LaterBirths[0].Line < 1000 || len(captured.Issues) != 0 {
		t.Fatalf("ParseCapture = %+v, %v, want the later birth found far down a journal read to its end", captured, err)
	}
}

func TestABirthBeyondTheScanLimitIsNotFoundAndTheLimitIsNamed(t *testing.T) {
	t.Parallel()
	journal := io.MultiReader(filler(linesBeyondTheScanLimit), strings.NewReader(birthLine(t, nil)+"\n"))
	_, err := ParseCapture(journal)
	invalid := violation(t, err, ErrNoBirth)
	if !strings.Contains(invalid.Issues[0].Message, "in the first 33554432 bytes") {
		t.Fatalf("issues = %+v, want the scan limit named", invalid.Issues)
	}
}

type failing struct{ err error }

func (f failing) Read([]byte) (int, error) { return 0, f.err }

func TestAReadThatFailsIsReportedAndIsNotTakenForTheEndOfTheInput(t *testing.T) {
	t.Parallel()
	broken := errors.New("disk gone")
	_, err := ParseCapture(io.MultiReader(strings.NewReader(birthLine(t, nil)+"\n"), failing{broken}))
	if !errors.Is(err, broken) {
		t.Fatalf("ParseCapture = %v, want the read failure", err)
	}
}
