package model

func RepeatedStateCode(states []ChannelState) (int, bool) {
	seen := make(map[int]bool, len(states))
	for index, state := range states {
		if seen[state.Value] {
			return index, true
		}
		seen[state.Value] = true
	}
	return 0, false
}
