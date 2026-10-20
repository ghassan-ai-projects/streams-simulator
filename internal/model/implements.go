package model

type ImplementsChecked struct {
	Kind        string `json:"kind"`
	KindDigest  string `json:"kind_digest"`
	BoardType   string `json:"board_type,omitempty"`
	BoardDigest string `json:"board_digest,omitempty"`
}

type DomainCheck func(*DomainSpec) ([]ImplementsChecked, error)

func (s *DomainSpec) HasImplements() bool {
	for i := range s.Kinds {
		if s.Kinds[i].Implements != "" {
			return true
		}
	}
	return false
}
