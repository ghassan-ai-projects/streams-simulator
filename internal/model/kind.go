package model

func (k *Kind) StateDecl(name string) *State {
	for i := range k.State {
		if k.State[i].Name == name {
			return &k.State[i]
		}
	}
	return nil
}

func (k *Kind) DynamicsOf(state string) *Dynamics {
	for i := range k.Dynamics {
		if k.Dynamics[i].Target == state {
			return &k.Dynamics[i]
		}
	}
	return nil
}

func (k *Kind) ChannelDecl(name string) *Channel {
	for i := range k.Channels {
		if k.Channels[i].Name == name {
			return &k.Channels[i]
		}
	}
	return nil
}

func (k *Kind) FaultDecl(id string) *Fault {
	for i := range k.Faults {
		if k.Faults[i].ID == id {
			return &k.Faults[i]
		}
	}
	return nil
}

func (k *Kind) EffectorDecl(name string) *Effector {
	for i := range k.Effectors {
		if k.Effectors[i].Name == name {
			return &k.Effectors[i]
		}
	}
	return nil
}
