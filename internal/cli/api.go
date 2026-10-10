package cli

// Build is the identity of the running binary, injected by the entrypoint
// from the Makefile's linker flags. The release manifest records it.
type Build struct {
	Version string
	Commit  string
}
