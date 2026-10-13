package domain

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/ghassan-ai-projects/streams-simulator/internal/model"
)

const builtInNamespace = "streamsim/"

type File struct {
	Path string
	Raw  []byte
}

type Directory struct {
	Path        string
	Equipment   []File
	Instruments []File
	Vocabulary  *File
}

type Library struct {
	equipment   map[string]*Equipment
	instruments map[string]*Instrument
	vocabulary  *Vocabulary
	contracts   *contracts
	owners      map[string]string
}

func LoadLibrary(directories []Directory) (*Library, error) {
	set, err := compileContracts()
	if err != nil {
		return nil, err
	}
	lib := newLibrary(set)
	if err := lib.admitAll(directories); err != nil {
		return nil, err
	}
	if err := lib.check(); err != nil {
		return nil, err
	}
	return lib, nil
}

func newLibrary(set *contracts) *Library {
	return &Library{equipment: map[string]*Equipment{}, instruments: map[string]*Instrument{}, vocabulary: newVocabulary(),
		contracts: set, owners: map[string]string{}}
}

func (l *Library) admitAll(directories []Directory) error {
	for position, directory := range directories {
		if err := l.admit(position, directory); err != nil {
			return err
		}
	}
	return nil
}

func (l *Library) admit(position int, directory Directory) error {
	if err := l.admitVocabulary(directory.Vocabulary); err != nil {
		return err
	}
	for _, file := range sortedFiles(directory.Equipment) {
		if err := l.admitEquipment(position, file); err != nil {
			return err
		}
	}
	return l.admitInstruments(position, directory.Instruments)
}

func (l *Library) admitInstruments(position int, files []File) error {
	for _, file := range sortedFiles(files) {
		if err := l.admitInstrument(position, file); err != nil {
			return err
		}
	}
	return nil
}

func sortedFiles(files []File) []File {
	sorted := slices.Clone(files)
	slices.SortFunc(sorted, func(a, b File) int { return strings.Compare(a.Path, b.Path) })
	return sorted
}

func (l *Library) admitVocabulary(file *File) error {
	if file == nil {
		return nil
	}
	doc, err := parseFile[Vocabulary](l.contracts.vocabulary, file.Raw, file.Path)
	if err != nil {
		return err
	}
	return l.mergeVocabulary(doc.value, file.Path)
}

func (l *Library) mergeVocabulary(add *Vocabulary, path string) error {
	if err := l.vocabulary.merge(add); err != nil {
		return refuse(model.ContractVocabulary, path, "", "%v", err)
	}
	return nil
}

type claim struct {
	contract model.Contract
	position int
	path, id string
	version  int
}

func (l *Library) admitEquipment(position int, file File) error {
	equipment, err := parseEquipment(l.contracts.component, file.Raw, file.Path)
	if err != nil {
		return err
	}
	return keep(l, l.equipment, claim{model.ContractComponent, position, file.Path, equipment.ID, equipment.Version}, equipment)
}

func (l *Library) admitInstrument(position int, file File) error {
	instrument, err := parseInstrument(l.contracts.instrument, file.Raw, file.Path)
	if err != nil {
		return err
	}
	return keep(l, l.instruments, claim{model.ContractInstrument, position, file.Path, instrument.ID, instrument.Version}, instrument)
}

func keep[T any](l *Library, destination map[string]*T, who claim, value *T) error {
	if err := l.claim(who); err != nil {
		return err
	}
	destination[reference(who.id, who.version)] = value
	return nil
}

func (l *Library) claim(who claim) error {
	if who.position > 0 && strings.HasPrefix(who.id, builtInNamespace) {
		return refuse(who.contract, who.path, "/id", "%s is in the reserved %q namespace, which only the built-in library (the first directory) may use", who.id, builtInNamespace)
	}
	key := reference(who.id, who.version)
	if first, taken := l.owners[key]; taken {
		return refuse(who.contract, who.path, "/id", "%s is already defined by %s", key, first)
	}
	l.owners[key] = who.path
	return nil
}

func reference(id string, version int) string { return fmt.Sprintf("%s@%d", id, version) }

func (l *Library) check() error {
	for _, key := range slices.Sorted(maps.Keys(l.equipment)) {
		if err := l.checkEquipment(l.equipment[key]); err != nil {
			return err
		}
	}
	for _, key := range slices.Sorted(maps.Keys(l.instruments)) {
		if err := l.checkInstrument(l.instruments[key]); err != nil {
			return err
		}
	}
	return nil
}
