// Package files is the file-system edge of the release module: the directories
// whose files a release digests, the signing key file and the manifest file.
package files

import "github.com/ghassan-ai-projects/streams-simulator/internal/kernel"

func Names(dir string) ([]string, error) {
	return kernel.Names(dir)
}

func Read(path string) ([]byte, error) {
	return kernel.Read(path)
}

func Write(path string, data []byte) error {
	return kernel.WriteAtomic(path, data)
}
