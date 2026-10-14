package transport

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/ghassan-ai-projects/streams-simulator/internal/canvasbridge/internal/domain"
)

var statuses = []struct {
	err    error
	status int
	code   string
}{
	{domain.ErrNotAdmitted, http.StatusForbidden, "not_admitted"},
	{domain.ErrUnpaired, http.StatusUnauthorized, "unpaired"},
	{domain.ErrWrongCode, http.StatusUnauthorized, "wrong_code"},
	{domain.ErrPairingLocked, http.StatusUnauthorized, "pairing_locked"},
	{domain.ErrPairingUsed, http.StatusUnauthorized, "pairing_used"},
	{domain.ErrBadPath, http.StatusBadRequest, "bad_path"},
	{domain.ErrBadArguments, http.StatusBadRequest, "bad_arguments"},
	{domain.ErrUnknownAction, http.StatusNotFound, "unknown_action"},
	{domain.ErrAbsent, http.StatusNotFound, "not_found"},
	{domain.ErrReadOnly, http.StatusForbidden, "read_only"},
	{domain.ErrSealed, http.StatusForbidden, "sealed"},
	{domain.ErrConflict, http.StatusConflict, "conflict"},
	{domain.ErrTooLarge, http.StatusRequestEntityTooLarge, "too_large"},
}

func writeError(w http.ResponseWriter, err error) {
	status, failure := classify(err)
	writeJSON(w, status, map[string]any{"error": failure})
}

func classify(err error) (int, *domain.Failure) {
	var failure *domain.Failure
	if errors.As(err, &failure) {
		return http.StatusUnprocessableEntity, failure
	}
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return http.StatusRequestEntityTooLarge, &domain.Failure{Code: "too_large", Message: "the request body is too large"}
	}
	return classifySentinel(err)
}

func classifySentinel(err error) (int, *domain.Failure) {
	for _, entry := range statuses {
		if errors.Is(err, entry.err) {
			return entry.status, &domain.Failure{Code: entry.code, Message: err.Error()}
		}
	}
	return http.StatusInternalServerError, &domain.Failure{Code: "bridge_failed", Message: err.Error()}
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func readBody(w http.ResponseWriter, r *http.Request, limit int64, into any) bool {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, limit))
	decoder.UseNumber()
	if err := decoder.Decode(into); err != nil {
		writeError(w, decodeFailure(err))
		return false
	}
	return true
}

func decodeFailure(err error) error {
	var tooLarge *http.MaxBytesError
	if errors.As(err, &tooLarge) {
		return err
	}
	return errors.Join(domain.ErrBadArguments, err)
}
