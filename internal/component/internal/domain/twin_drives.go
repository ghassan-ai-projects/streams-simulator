package domain

import (
	"maps"
	"slices"
)

func (t *twinBuild) planDrives() error {
	for _, k := range t.kinds {
		for _, c := range k.commands {
			for j := range c.choice.Drives {
				if err := t.planDrive(k, c, j); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (t *twinBuild) planDrive(k *twinKind, c *commandPlan, j int) error {
	drive, at := c.choice.Drives[j], c.at+pointer("drives", indexPart(j))
	target, err := t.drivenMetric(drive, at)
	if err != nil {
		return err
	}
	if err := t.checkDrive(c, drive, target, at); err != nil {
		return err
	}
	target.driven = append(target.driven, coupling{driver: k, state: valueName(drive.From), gain: drive.Gain})
	target.tauS = drive.TimeConstantS
	return nil
}

func (t *twinBuild) drivenMetric(drive driveChoice, at string) (*metricPlan, error) {
	kind := t.byName[drive.Kind]
	if kind == nil {
		return nil, t.refuse(ErrUnknownChoice, at+"/kind", "no kind %q is chosen", drive.Kind)
	}
	plan := kind.plan(drive.Metric)
	if plan == nil {
		return nil, t.refuse(ErrUnknownChoice, at+"/metric", "kind %q has no metric %q", drive.Kind, drive.Metric)
	}
	return plan, nil
}

func (t *twinBuild) checkDrive(c *commandPlan, drive driveChoice, target *metricPlan, at string) error {
	switch {
	case !slices.Contains(slices.Collect(maps.Values(c.choice.Sets)), drive.From):
		return t.refuse(ErrResponseMismatch, at+"/from", "command %q does not set metric %q", c.file.Name, drive.From)
	case drive.Gain == 0:
		return t.refuse(ErrResponseMismatch, at+"/gain", "a drive with gain 0 moves nothing")
	case target.file.Writable || target.file.DataType == "boolean":
		return t.refuse(ErrResponseMismatch, at+"/metric", "metric %q is set by a command or is boolean: only a read-only number or integer follows a drive", target.file.Name)
	case target.tauS != 0 && target.tauS != drive.TimeConstantS:
		return t.refuse(ErrResponseMismatch, at+"/time_constant_s", "metric %q follows its drives with time constant %v s, not %v s", target.file.Name, target.tauS, drive.TimeConstantS)
	}
	return nil
}
