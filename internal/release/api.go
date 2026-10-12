package release

import "github.com/ghassan-ai-projects/streams-simulator/internal/release/internal/app"

// Options say what to publish: the output path, the directories whose files
// are digested, the author and independent reviewer identities, the optional
// signing key file, the build identity of the simulator and the clock that
// stamps the manifest.
type Options = app.Options

// Result is what a publication left behind: the manifest path, whether it was
// signed, and the simulator identity it records.
type Result = app.Result
