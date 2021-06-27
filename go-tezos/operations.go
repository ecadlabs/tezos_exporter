package tezos

import (
	"encoding/json"
	"math/big"
)

// OperationElem must be implemented by all operation elements
type OperationElem interface {
	OperationElemKind() string
}

// BalanceUpdatesOperation is implemented by operations providing balance updates
type BalanceUpdatesOperation interface {
	BalanceUpdates() BalanceUpdates
}

// OperationWithFee is implemented by operations with fees
type OperationWithFee interface {
	OperationFee() *big.Int
}

// GenericOperationElem is a most generic element type
type GenericOperationElem struct {
	Kind string `json:"kind" yaml:"kind"`
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
		case "transaction":
			(*e)[i] = &TransactionOperationElem{}
		case "ballot":
			(*e)[i] = &BallotOperationElem{}
		case "proposals":
			(*e)[i] = &ProposalOperationElem{}
		case "seed_nonce_revelation":
			(*e)[i] = &SeedNonceRevelationOperationElem{}
		case "double_endorsement_evidence":
			(*e)[i] = &DoubleEndorsementEvidenceOperationElem{}
		case "double_baking_evidence":
			(*e)[i] = &DoubleBakingEvidenceOperationElem{}
		case "activate_account":
			(*e)[i] = &ActivateAccountOperationElem{}
		case "reveal":
			(*e)[i] = &RevealOperationElem{}
		case "origination":
			(*e)[i] = &OriginationOperationElem{}
		case "delegation":
			(*e)[i] = &DelegationOperationElem{}
		default:
			(*e)[i] = &tmp
			continue opLoop
		}

		if err := json.Unmarshal(r, (*e)[i]); err != nil {
			return err
		}
	}

	return nil
}

// EndorsementOperationElem represents an endorsement operation
type EndorsementOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Level                int                          `json:"level" yaml:"level"`
	Metadata             EndorsementOperationMetadata `json:"metadata" yaml:"metadata"`
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *EndorsementOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// EndorsementOperationMetadata represents an endorsement operation metadata
type EndorsementOperationMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates" yaml:"balance_updates"`
	Delegate       string         `json:"delegate" yaml:"delegate"`
	Slots          []int          `json:"slots" yaml:"slots,flow"`
}

// TransactionOperationElem represents a transaction operation
type TransactionOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Source               string                       `json:"source" yaml:"source"`
	Fee                  *BigInt                      `json:"fee" yaml:"fee"`
	Counter              *BigInt                      `json:"counter" yaml:"counter"`
	GasLimit             *BigInt                      `json:"gas_limit" yaml:"gas_limit"`
	StorageLimit         *BigInt                      `json:"storage_limit" yaml:"storage_limit"`
	Amount               *BigInt                      `json:"amount" yaml:"amount"`
	Destination          string                       `json:"destination" yaml:"destination"`
	Parameters           map[string]interface{}       `json:"parameters,omitempty" yaml:"parameters,omitempty"`
	Metadata             TransactionOperationMetadata `json:"metadata" yaml:"metadata"`
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *TransactionOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// OperationFee implements OperationWithFee
func (el *TransactionOperationElem) OperationFee() *big.Int {
	if el.Fee != nil {
		return &el.Fee.Int
	}
	return big.NewInt(0)
}

// TransactionOperationMetadata represents a transaction operation metadata
type TransactionOperationMetadata struct {
	BalanceUpdates  BalanceUpdates             `json:"balance_updates" yaml:"balance_updates"`
	OperationResult TransactionOperationResult `json:"operation_result" yaml:"operation_result"`
}

// TransactionOperationResult represents a transaction operation result
type TransactionOperationResult struct {
	Status              string                 `json:"status" yaml:"status"`
	Storage             map[string]interface{} `json:"storage,omitempty" yaml:"storage,omitempty"`
	BalanceUpdates      BalanceUpdates         `json:"balance_updates,omitempty" yaml:"balance_updates,omitempty"`
	OriginatedContracts []string               `json:"originated_contracts,omitempty" yaml:"originated_contracts,omitempty"`
	ConsumedGas         *BigInt                `json:"consumed_gas,omitempty" yaml:"consumed_gas,omitempty"`
	StorageSize         *BigInt                `json:"storage_size,omitempty" yaml:"storage_size,omitempty"`
	PaidStorageSizeDiff *BigInt                `json:"paid_storage_size_diff,omitempty" yaml:"paid_storage_size_diff,omitempty"`
	Errors              Errors                 `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// BallotOperationElem represents a ballot operation
type BallotOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Source               string                 `json:"source" yaml:"source"`
	Period               int                    `json:"period" yaml:"period"`
	Proposal             string                 `json:"proposal" yaml:"proposal"`
	Ballot               string                 `json:"ballot" yaml:"ballot"`
	Metadata             map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// ProposalOperationElem represents a proposal operation
type ProposalOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Source               string                 `json:"source" yaml:"source"`
	Period               int                    `json:"period" yaml:"period"`
	Proposals            []string               `json:"proposals" yaml:"proposals"`
	Metadata             map[string]interface{} `json:"metadata" yaml:"metadata"`
}

// SeedNonceRevelationOperationElem represents seed_nonce_revelation operation
type SeedNonceRevelationOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Level                int32                           `json:"level" yaml:"level"`
	Nonce                string                          `json:"nonce" yaml:"nonce"`
	Metadata             BalanceUpdatesOperationMetadata `json:"metadata" yaml:"metadata"`
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *SeedNonceRevelationOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// BalanceUpdatesOperationMetadata contains balance updates only
type BalanceUpdatesOperationMetadata struct {
	BalanceUpdates BalanceUpdates `json:"balance_updates" yaml:"balance_updates"`
}

// InlinedEndorsement corresponds to $inlined.endorsement
type InlinedEndorsement struct {
	Branch     string                     `json:"branch" yaml:"branch"`
	Operations InlinedEndorsementContents `json:"operations" yaml:"operations"`
	Signature  string                     `json:"signature" yaml:"signature"`
}

// InlinedEndorsementContents corresponds to $inlined.endorsement.contents
type InlinedEndorsementContents struct {
	Kind  string `json:"endorsement" yaml:"endorsement"`
	Level int    `json:"level" yaml:"level"`
}

// DoubleEndorsementEvidenceOperationElem represents double_endorsement_evidence operation
type DoubleEndorsementEvidenceOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Operation1           InlinedEndorsement              `json:"op1" yaml:"op1"`
	Operation2           InlinedEndorsement              `json:"op2" yaml:"op2"`
	Metadata             BalanceUpdatesOperationMetadata `json:"metadata" yaml:"metadata"`
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *DoubleEndorsementEvidenceOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// DoubleBakingEvidenceOperationElem represents double_baking_evidence operation
type DoubleBakingEvidenceOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	BlockHeader1         RawBlockHeader                  `json:"bh1" yaml:"bh1"`
	BlockHeader2         RawBlockHeader                  `json:"bh2" yaml:"bh2"`
	Metadata             BalanceUpdatesOperationMetadata `json:"metadata" yaml:"metadata"`
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *DoubleBakingEvidenceOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// ActivateAccountOperationElem represents activate_account operation
type ActivateAccountOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	PKH                  string                          `json:"pkh" yaml:"pkh"`
	Secret               string                          `json:"secret" yaml:"secret"`
	Metadata             BalanceUpdatesOperationMetadata `json:"metadata" yaml:"metadata"`
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *ActivateAccountOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// RevealOperationElem represents a reveal operation
type RevealOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Source               string                  `json:"source" yaml:"source"`
	Fee                  *BigInt                 `json:"fee" yaml:"fee"`
	Counter              *BigInt                 `json:"counter" yaml:"counter"`
	GasLimit             *BigInt                 `json:"gas_limit" yaml:"gas_limit"`
	StorageLimit         *BigInt                 `json:"storage_limit" yaml:"storage_limit"`
	PublicKey            string                  `json:"public_key" yaml:"public_key"`
	Metadata             RevealOperationMetadata `json:"metadata" yaml:"metadata"`
}

// OperationFee implements OperationWithFee
func (el *RevealOperationElem) OperationFee() *big.Int {
	if el.Fee != nil {
		return &el.Fee.Int
	}
	return big.NewInt(0)
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *RevealOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// RevealOperationMetadata represents a reveal operation metadata
type RevealOperationMetadata DelegationOperationMetadata

// OriginationOperationElem represents a origination operation
type OriginationOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Source               string                       `json:"source" yaml:"source"`
	Fee                  *BigInt                      `json:"fee" yaml:"fee"`
	Counter              *BigInt                      `json:"counter" yaml:"counter"`
	GasLimit             *BigInt                      `json:"gas_limit" yaml:"gas_limit"`
	StorageLimit         *BigInt                      `json:"storage_limit" yaml:"storage_limit"`
	ManagerPubKey        string                       `json:"managerPubkey" yaml:"managerPubkey"`
	Balance              *BigInt                      `json:"balance" yaml:"balance"`
	Spendable            *bool                        `json:"spendable,omitempty" yaml:"spendable,omitempty"`
	Delegatable          *bool                        `json:"delegatable,omitempty" yaml:"delegatable,omitempty"`
	Delegate             string                       `json:"delegate,omitempty" yaml:"delegate,omitempty"`
	Script               *ScriptedContracts           `json:"script,omitempty" yaml:"script,omitempty"`
	Metadata             OriginationOperationMetadata `json:"metadata" yaml:"metadata"`
}

// OperationFee implements OperationWithFee
func (el *OriginationOperationElem) OperationFee() *big.Int {
	if el.Fee != nil {
		return &el.Fee.Int
	}
	return big.NewInt(0)
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *OriginationOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// ScriptedContracts corresponds to $scripted.contracts
type ScriptedContracts struct {
	Code    map[string]interface{} `json:"code" yaml:"code"`
	Storage map[string]interface{} `json:"storage" yaml:"storage"`
}

// OriginationOperationMetadata represents a origination operation metadata
type OriginationOperationMetadata struct {
	BalanceUpdates  BalanceUpdates             `json:"balance_updates" yaml:"balance_updates"`
	OperationResult OriginationOperationResult `json:"operation_result" yaml:"operation_result"`
}

// OriginationOperationResult represents a origination operation result
type OriginationOperationResult struct {
	Status              string         `json:"status" yaml:"status"`
	BalanceUpdates      BalanceUpdates `json:"balance_updates,omitempty" yaml:"balance_updates,omitempty"`
	OriginatedContracts []string       `json:"originated_contracts,omitempty" yaml:"originated_contracts,omitempty"`
	ConsumedGas         *BigInt        `json:"consumed_gas,omitempty" yaml:"consumed_gas,omitempty"`
	StorageSize         *BigInt        `json:"storage_size,omitempty" yaml:"storage_size,omitempty"`
	PaidStorageSizeDiff *BigInt        `json:"paid_storage_size_diff,omitempty" yaml:"paid_storage_size_diff,omitempty"`
	Errors              Errors         `json:"errors,omitempty" yaml:"errors,omitempty"`
}

// DelegationOperationElem represents a delegation operation
type DelegationOperationElem struct {
	GenericOperationElem `yaml:",inline"`
	Source               string                      `json:"source" yaml:"source"`
	Fee                  *BigInt                     `json:"fee" yaml:"fee"`
	Counter              *BigInt                     `json:"counter" yaml:"counter"`
	GasLimit             *BigInt                     `json:"gas_limit" yaml:"gas_limit"`
	StorageLimit         *BigInt                     `json:"storage_limit" yaml:"storage_limit"`
	ManagerPubKey        string                      `json:"managerPubkey" yaml:"managerPubkey"`
	Balance              *BigInt                     `json:"balance" yaml:"balance"`
	Spendable            *bool                       `json:"spendable,omitempty" yaml:"spendable,omitempty"`
	Delegatable          *bool                       `json:"delegatable,omitempty" yaml:"delegatable,omitempty"`
	Delegate             string                      `json:"delegate,omitempty" yaml:"delegate,omitempty"`
	Script               *ScriptedContracts          `json:"script,omitempty" yaml:"script,omitempty"`
	Metadata             DelegationOperationMetadata `json:"metadata" yaml:"metadata"`
}

// OperationFee implements OperationWithFee
func (el *DelegationOperationElem) OperationFee() *big.Int {
	if el.Fee != nil {
		return &el.Fee.Int
	}
	return big.NewInt(0)
}

// BalanceUpdates implements BalanceUpdateOperation
func (el *DelegationOperationElem) BalanceUpdates() BalanceUpdates {
	return el.Metadata.BalanceUpdates
}

// DelegationOperationMetadata represents a delegation operation metadata
type DelegationOperationMetadata struct {
	BalanceUpdates  BalanceUpdates            `json:"balance_updates" yaml:"balance_updates"`
	OperationResult DelegationOperationResult `json:"operation_result" yaml:"operation_result"`
}

// DelegationOperationResult represents a delegation operation result
type DelegationOperationResult struct {
	Status string `json:"status" yaml:"status"`
	Errors Errors `json:"errors" yaml:"errors"`
}

// BalanceUpdate is a variable structure depending on the Kind field
type BalanceUpdate interface {
	BalanceUpdateKind() string
}

// GenericBalanceUpdate holds the common values among all BalanceUpdatesType variants
type GenericBalanceUpdate struct {
	Kind   string `json:"kind" yaml:"kind"`
	Change int64  `json:"change,string" yaml:"change"`
}

// BalanceUpdateKind returns the BalanceUpdateType's Kind field
func (g *GenericBalanceUpdate) BalanceUpdateKind() string {
	return g.Kind
}

// ContractBalanceUpdate is a BalanceUpdatesType variant for Kind=contract
type ContractBalanceUpdate struct {
	GenericBalanceUpdate `yaml:",inline"`
	Contract             string `json:"contract" yaml:"contract"`
}

// FreezerBalanceUpdate is a BalanceUpdatesType variant for Kind=freezer
type FreezerBalanceUpdate struct {
	GenericBalanceUpdate `yaml:",inline"`
	Category             string `json:"category" yaml:"category"`
	Delegate             string `json:"delegate" yaml:"delegate"`
	Level                int    `json:"level" yaml:"level"`
}

// BalanceUpdates is a list of balance update operations
type BalanceUpdates []BalanceUpdate

// UnmarshalJSON implements json.Unmarshaler
func (b *BalanceUpdates) UnmarshalJSON(data []byte) error {
	var raw []json.RawMessage

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	*b = make(BalanceUpdates, len(raw))

opLoop:
	for i, r := range raw {
		var tmp GenericBalanceUpdate
		if err := json.Unmarshal(r, &tmp); err != nil {
			return err
		}

		switch tmp.Kind {
		case "contract":
			(*b)[i] = &ContractBalanceUpdate{}

		case "freezer":
			(*b)[i] = &FreezerBalanceUpdate{}

		default:
			(*b)[i] = &tmp
			continue opLoop
		}

		if err := json.Unmarshal(r, (*b)[i]); err != nil {
			return err
		}
	}

	return nil
}

// Operation represents an operation included into block
type Operation struct {
	Protocol  string            `json:"protocol" yaml:"protocol"`
	ChainID   string            `json:"chain_id" yaml:"chain_id"`
	Hash      string            `json:"hash" yaml:"hash"`
	Branch    string            `json:"branch" yaml:"branch"`
	Contents  OperationElements `json:"contents" yaml:"contents"`
	Signature string            `json:"signature" yaml:"signature"`
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
	Error Errors `json:"error" yaml:"error"`
}

// OperationWithErrorAlt is a heterogeneously encoded OperationWithError with hash as a first array member.
// See OperationAlt for details
type OperationWithErrorAlt OperationWithError

// UnmarshalJSON implements json.Unmarshaler
func (o *OperationWithErrorAlt) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &o.Hash, (*OperationWithError)(o))
}

var (
	_ BalanceUpdatesOperation = &EndorsementOperationElem{}
	_ BalanceUpdatesOperation = &TransactionOperationElem{}
	_ BalanceUpdatesOperation = &SeedNonceRevelationOperationElem{}
	_ BalanceUpdatesOperation = &DoubleEndorsementEvidenceOperationElem{}
	_ BalanceUpdatesOperation = &DoubleBakingEvidenceOperationElem{}
	_ BalanceUpdatesOperation = &ActivateAccountOperationElem{}
	_ BalanceUpdatesOperation = &RevealOperationElem{}
	_ BalanceUpdatesOperation = &OriginationOperationElem{}
	_ BalanceUpdatesOperation = &DelegationOperationElem{}

	_ OperationWithFee = &TransactionOperationElem{}
	_ OperationWithFee = &RevealOperationElem{}
	_ OperationWithFee = &OriginationOperationElem{}
	_ OperationWithFee = &DelegationOperationElem{}
)
