package tezos

import (
	"encoding/json"
	"strconv"
)

// OperationElem must be implemented by all operation elements
type OperationElem interface {
	OperationElemKind() string
}

// GenericOperationElem is a most generic element type
type GenericOperationElem struct {
	Kind string `json:"kind"`
}

// OperationElemKind implements OperationElem
func (e *GenericOperationElem) OperationElemKind() string {
	return e.Kind
}

// OperationElements is a slice of OperationElem with custom JSON unmarshaller
type OperationElements []OperationElem

// UnmarshalJSON implements json.Unmarshaler
func (e *OperationElements) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*e = make(OperationElements, len(raw))

opLoop:
	for i, r := range raw {
		var tmp GenericOperationElem
		if err := json.Unmarshal(r, &tmp); err != nil {
			return err
		}

		switch tmp.Kind {
		case "endorsement":
			(*e)[i] = &EndorsementOperationElem{}

		default:
			(*e)[i] = &tmp
			continue opLoop

			// TODO: add other item kinds
		}

		if err := json.Unmarshal(r, (*e)[i]); err != nil {
			return err
		}
	}

	return nil
}

// EndorsementOperationElem represents an endorsement operation
type EndorsementOperationElem struct {
	GenericOperationElem
	Level    int                           `json:"level"`
	Metadata *EndorsementOperationMetadata `json:"metadata"`
}

// EndorsementOperationMetadata represents an endorsement operation metadata
type EndorsementOperationMetadata struct {
	BalanceUpdates []*BalanceUpdate `json:"balance_updates"`
	Delegate       string           `json:"delegate"`
	Slots          []uint           `json:"slots"`
}

// BalanceUpdate represents a balance update operation
type BalanceUpdate struct {
	Kind     string        `json:"kind"`
	Category string        `json:"category"`
	Delegate string        `json:"delegate"`
	Level    int           `json:"level"`
	Change   BalanceChange `json:"change"`
	Contract string        `json:"contract"`
}

// BalanceChange is a string encoded int64
type BalanceChange int64

// UnmarshalJSON implements json.Unmarshaler
func (b *BalanceChange) UnmarshalJSON(data []byte) (err error) {
	var s string
	if err = json.Unmarshal(data, &s); err != nil {
		return err
	}

	*(*int64)(b), err = strconv.ParseInt(s, 0, 64)

	return err
}

// Operation represents an operation included into block
type Operation struct {
	Protocol  string            `json:"protocol"`
	ChainID   string            `json:"chain_id"`
	Hash      string            `json:"hash"`
	Branch    string            `json:"branch"`
	Contents  OperationElements `json:"contents"`
	Signature string            `json:"signature"`
}

/*
OperationAlt is a heterogeneously encoded Operation with hash as a first array member, i.e.
	[
		"...", // hash
		{
			"protocol": "...",
			...
		}
	]
instead of
	{
		"protocol": "...",
		"hash": "...",
		...
	}
*/
type OperationAlt Operation

// UnmarshalJSON implements json.Unmarshaler
func (o *OperationAlt) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &o.Hash, (*Operation)(o))
}

// OperationWithError represents unsuccessful operation
type OperationWithError struct {
	Operation
	Error Errors `json:"error"`
}

// OperationWithErrorAlt is a heterogeneously encoded OperationWithError with hash as a first array member.
// See OperationAlt for details
type OperationWithErrorAlt OperationWithError

// UnmarshalJSON implements json.Unmarshaler
func (o *OperationWithErrorAlt) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &o.Hash, (*OperationWithError)(o))
}
