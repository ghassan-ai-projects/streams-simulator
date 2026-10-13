// Package domain holds the rules of the canvas bridge: pairing and sessions,
// request admission, the workspace paths the canvas may touch, the truth gate
// on run files and the action registry that maps each canvas action to a
// director MCP tool or a streamsim command. It performs no I/O.
package domain

import "errors"

var (
	ErrWrongCode     = errors.New("canvasbridge: the pairing code is wrong")
	ErrPairingLocked = errors.New("canvasbridge: pairing is locked after too many wrong codes; restart the bridge")
	ErrPairingUsed   = errors.New("canvasbridge: the pairing code was already used; restart the bridge to pair again")
	ErrNotAdmitted   = errors.New("canvasbridge: request refused")
	ErrBadPath       = errors.New("canvasbridge: path is not a workspace file the canvas may use")
	ErrReadOnly      = errors.New("canvasbridge: run files are read only")
	ErrSealed        = errors.New("canvasbridge: run file is sealed until the run ends")
	ErrUnknownAction = errors.New("canvasbridge: unknown action")
	ErrBadArguments  = errors.New("canvasbridge: invalid action arguments")
	ErrNonLoopback   = errors.New("canvasbridge: the bridge listens on a loopback address only")
	ErrAbsent        = errors.New("canvasbridge: file not found")
	ErrTooLarge      = errors.New("canvasbridge: file is too large")
	ErrConflict      = errors.New("canvasbridge: the file changed on disk since it was read")
	ErrUnpaired      = errors.New("canvasbridge: this browser is not paired; enter the pairing code the bridge printed")
)
