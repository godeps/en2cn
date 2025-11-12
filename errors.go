package en2cn

import "errors"

var (
	// ErrIPAUnavailable indicates that the requested English word was not found in the IPA database.
	ErrIPAUnavailable = errors.New("ipa not found for word")
	// ErrNoCandidate indicates that no viable Chinese candidate existed in the candidate DB.
	ErrNoCandidate = errors.New("no candidate words available")
)
