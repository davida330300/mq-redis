package payload

import (
	"encoding/json"
	"strings"
)

type Decision int

const (
	DecisionInline Decision = iota
	DecisionRef
	DecisionMissing
	DecisionConflict
	DecisionInlineTooLarge
	DecisionRefMetaMissing
	DecisionError
)

type Input struct {
	Inline json.RawMessage
	Ref    string
	Size   int64
	Hash   string
}

type RefPayload struct {
	Ref  string `json:"payload_ref"`
	Size int64  `json:"payload_size"`
	Hash string `json:"payload_hash"`
}

func Normalize(input Input, maxInlineBytes int) (Decision, json.RawMessage, error) {
	trimmedRef := strings.TrimSpace(input.Ref)
	trimmedHash := strings.TrimSpace(input.Hash)

	inlineProvided := len(input.Inline) > 0
	refProvided := trimmedRef != ""

	if inlineProvided && refProvided {
		return DecisionConflict, nil, nil
	}
	if !inlineProvided && !refProvided {
		return DecisionMissing, nil, nil
	}
	if inlineProvided {
		if len(input.Inline) > maxInlineBytes {
			return DecisionInlineTooLarge, nil, nil
		}
		return DecisionInline, input.Inline, nil
	}
	if input.Size <= 0 || trimmedHash == "" {
		return DecisionRefMetaMissing, nil, nil
	}
	payload, err := json.Marshal(RefPayload{
		Ref:  trimmedRef,
		Size: input.Size,
		Hash: trimmedHash,
	})
	if err != nil {
		return DecisionError, nil, err
	}
	return DecisionRef, payload, nil
}
