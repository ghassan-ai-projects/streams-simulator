package domain

type Summary struct {
	Channels  int    `json:"channels"`
	Digest    string `json:"digest"`
	Effectors int    `json:"effectors"`
	Faults    int    `json:"faults"`
	ID        string `json:"id"`
}

func Summarize(c *Compiled) Summary {
	return Summary{Channels: len(c.ChannelNames()), Digest: c.Digest, Effectors: c.EffectorCount(), Faults: c.FaultCount(), ID: c.Spec.ID}
}
