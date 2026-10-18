package domain

type Admission struct {
	Environment string
	ProtocolMax int
}

func (a Admission) Refusal(m Message, uncommissioned bool) string {
	switch {
	case ProtocolInvalid(m):
		return "protocol_invalid"
	case EnvironmentFaultFor(a.Environment, m) != nil:
		return "environment_mismatch"
	case ContractUnsupported(m):
		return "contract_unsupported"
	case m.ProtocolMin > int64(a.ProtocolMax) || m.ProtocolMax < Version:
		return "version_unsupported"
	case uncommissioned || KindsMisdigested(m.Kinds) || KindsDuplicated(m.Kinds):
		return "inventory_mismatch"
	}
	return ""
}

func (a Admission) Reason(refusal string, m Message) string {
	if fault := EnvironmentFaultFor(a.Environment, m); refusal == "environment_mismatch" && fault != nil {
		return fault.Reason
	}
	return ""
}
