package tezos

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"time"
)

// HexBytes represents bytes as a JSON string of hexadecimal digits
type HexBytes []byte

// UnmarshalText umarshalls a hex string to bytes
func (hb *HexBytes) UnmarshalText(data []byte) error {
	dst := make([]byte, hex.DecodedLen(len(data)))
	if _, err := hex.Decode(dst, data); err != nil {
		return err
	}
	*hb = dst
	return nil
}

// MonitorBlock holds information about block returned by monitor heads endpoint
type MonitorBlock struct {
	Level          int        `json:"level"`
	Proto          int        `json:"proto"`
	Predecessor    string     `json:"predecessor"`
	Timestamp      time.Time  `json:"timestamp"`
	ValidationPass int        `json:"validation_pass"`
	OperationsHash string     `json:"operations_hash"`
	Fitness        []HexBytes `json:"fitness"`
	Context        string     `json:"context"`
	Priority       int        `json:"priority"`
	Hash           string     `json:"hash"`
	ProtocolData   string     `json:"protocol_data"`
}

// RawBlockHeader is a part of the Tezos block data
type RawBlockHeader struct {
	Level            int        `json:"level"`
	Proto            int        `json:"proto"`
	Predecessor      string     `json:"predecessor"`
	Timestamp        time.Time  `json:"timestamp"`
	ValidationPass   int        `json:"validation_pass"`
	OperationsHash   string     `json:"operations_hash"`
	Fitness          []HexBytes `json:"fitness"`
	Context          string     `json:"context"`
	Priority         int        `json:"priority"`
	ProofOfWorkNonce HexBytes   `json:"proof_of_work_nonce"`
	SeedNonceHash    string     `json:"seed_nonce_hash"`
	Signature        string     `json:"signature"`
}

// TestChainStatus is a variable structure depending on the Status field
type TestChainStatus interface {
	TestChainStatus() string
}

// GenericTestChainStatus holds the common values among all TestChainStatus variants
type GenericTestChainStatus struct {
	Status string `json:"status"`
}

// TestChainStatus gets the TestChainStatus's Status field
func (t *GenericTestChainStatus) TestChainStatus() string {
	return t.Status
}

// NotRunningTestChainStatus is a TestChainStatus variant for Status=not_running
type NotRunningTestChainStatus struct {
	GenericTestChainStatus
}

// ForkingTestChainStatus is a TestChainStatus variant for Status=forking
type ForkingTestChainStatus struct {
	GenericTestChainStatus
	Protocol   string `json:"protocol"`
	Expiration string `json:"expiration"`
}

// RunningTestChainStatus is a TestChainStatus variant for Status=running
type RunningTestChainStatus struct {
	GenericTestChainStatus
	ChainID    string `json:"chain_id"`
	Genesis    string `json:"genesis"`
	Protocol   string `json:"protocol"`
	Expiration string `json:"expiration"`
}

// MaxOperationListLength is a part of the BlockHeaderMetadata
type MaxOperationListLength struct {
	MaxSize int `json:"max_size"`
	MaxOp   int `json:"max_op"`
}

// BlockHeaderMetadataLevel is a part of BlockHeaderMetadata
type BlockHeaderMetadataLevel struct {
	Level                int  `json:"level"`
	LevelPosition        int  `json:"level_position"`
	Cycle                int  `json:"cycle"`
	CyclePosition        int  `json:"cycle_position"`
	VotingPeriod         int  `json:"voting_period"`
	VotingPeriodPosition int  `json:"voting_period_position"`
	ExpectedCommitment   bool `json:"expected_commitment"`
}

// BlockHeaderMetadata is a part of the Tezos block data
type BlockHeaderMetadata struct {
	Protocol               string                    `json:"protocol"`
	NextProtocol           string                    `json:"next_protocol"`
	TestChainStatus        TestChainStatus           `json:"-"`
	MaxOperationsTTL       int                       `json:"max_operations_ttl"`
	MaxOperationDataLength int                       `json:"max_operation_data_length"`
	MaxBlockHeaderLength   int                       `json:"max_block_header_length"`
	MaxOperationListLength []*MaxOperationListLength `json:"max_operation_list_length"`
	Baker                  string                    `json:"baker"`
	Level                  BlockHeaderMetadataLevel  `json:"level"`
	VotingPeriodKind       string                    `json:"voting_period_kind"`
	NonceHash              string                    `json:"nonce_hash"`
	ConsumedGas            big.Int                   `json:"-"`
	Deactivated            []string                  `json:"deactivated"`
	BalanceUpdates         BalanceUpdates            `json:"balance_updates"`
}

func unmarshalTestChainStatus(data []byte) (TestChainStatus, error) {
	var tmp GenericTestChainStatus
	if err := json.Unmarshal(data, &tmp); err != nil {
		return nil, err
	}

	var v TestChainStatus

	switch tmp.Status {
	case "not_running":
		v = &NotRunningTestChainStatus{}
	case "forking":
		v = &ForkingTestChainStatus{}
	case "running":
		v = &RunningTestChainStatus{}

	default:
		return nil, fmt.Errorf("Unknown TestChainStatus.Status: %v", tmp.Status)
	}

	if err := json.Unmarshal(data, v); err != nil {
		return nil, err
	}

	return v, nil
}

// UnmarshalJSON unmarshals the BlockHeaderMetadata JSON
func (bhm *BlockHeaderMetadata) UnmarshalJSON(data []byte) error {
	type suppressJSONUnmarshaller BlockHeaderMetadata
	if err := json.Unmarshal(data, (*suppressJSONUnmarshaller)(bhm)); err != nil {
		return err
	}

	var tmp struct {
		TestChainStatus json.RawMessage `json:"test_chain_status"`
		ConsumedGas     bigIntStr       `json:"consumed_gas"`
	}

	if err := json.Unmarshal(data, &tmp); err != nil {
		return err
	}

	tcs, err := unmarshalTestChainStatus(tmp.TestChainStatus)
	if err != nil {
		return err
	}

	bhm.TestChainStatus = tcs
	bhm.ConsumedGas = big.Int(tmp.ConsumedGas)

	return nil
}

// Block holds information about a Tezos block
type Block struct {
	Protocol   string              `json:"protocol"`
	ChainID    string              `json:"chain_id"`
	Hash       string              `json:"hash"`
	Header     RawBlockHeader      `json:"header"`
	Metadata   BlockHeaderMetadata `json:"metadata"`
	Operations [][]*Operation      `json:"operations"`
}
