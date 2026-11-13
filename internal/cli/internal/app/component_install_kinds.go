package app

import "github.com/ghassan-ai-projects/streams-simulator/internal/component"

const installKindsUsage = "usage: streamsim component install-kinds <example-id> --kinds-dir <dir>"

func cmdInstallKinds(s *session, args []string) (any, error) {
	fs := newFlagSet("component install-kinds", s.stderr)
	var kindsDir string
	fs.StringVar(&kindsDir, kindsDirFlag, "", "the kinds folder the example's kind files and boards are written into")
	if err := parseFlags(fs, args); err != nil {
		return nil, err
	}
	if fs.NArg() != 1 || kindsDir == "" {
		return nil, usagef(installKindsUsage)
	}
	installed, err := component.InstallExampleKinds(fs.Arg(0), kindsDir)
	return installed, wrapCLI(err)
}
