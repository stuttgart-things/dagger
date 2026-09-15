package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
)

// envelope is one line of `cosign verify-attestation` output: a DSSE envelope
// whose payload is an in-toto statement.
type envelope struct {
	PayloadType string `json:"payloadType"`
	Payload     string `json:"payload"`
}

// statement is the part of an in-toto statement VerifyAttestation hands back.
type statement struct {
	PredicateType string          `json:"predicateType"`
	Predicate     json.RawMessage `json:"predicate"`
}

// singlePredicate returns the predicate carried by the attestations cosign
// printed as verified, one envelope per line.
//
// Several attestations carrying the same predicate are one answer. Several
// carrying different predicates are an error: returning any one of them would
// be a guess at which the caller meant, and the others verify just as well.
func singlePredicate(output string) (json.RawMessage, error) {
	var predicates []json.RawMessage
	count := 0

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		var env envelope
		if err := json.Unmarshal([]byte(line), &env); err != nil {
			return nil, fmt.Errorf("cosign printed a line that is not a DSSE envelope: %w", err)
		}
		payload, err := base64.StdEncoding.DecodeString(env.Payload)
		if err != nil {
			return nil, fmt.Errorf("attestation payload is not base64: %w", err)
		}
		var st statement
		if err := json.Unmarshal(payload, &st); err != nil {
			return nil, fmt.Errorf("attestation payload is not an in-toto statement: %w", err)
		}
		if len(st.Predicate) == 0 || string(st.Predicate) == "null" {
			return nil, fmt.Errorf("attestation of type %s carries no predicate", st.PredicateType)
		}

		count++
		if !slices.ContainsFunc(predicates, func(p json.RawMessage) bool {
			return bytes.Equal(p, st.Predicate)
		}) {
			predicates = append(predicates, st.Predicate)
		}
	}

	switch len(predicates) {
	case 0:
		return nil, errors.New("cosign verified no attestation")
	case 1:
		return predicates[0], nil
	}

	return nil, fmt.Errorf(
		"%d verified attestations of this type carry %d different predicates, "+
			"and none of them is more the answer than the others; "+
			"attest with replace so that only one remains",
		count, len(predicates))
}
