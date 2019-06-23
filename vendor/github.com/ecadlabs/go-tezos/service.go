package tezos

import (
	"context"
	"fmt"
	"math/big"
	"net/http"
	"net/url"
	"time"
)

// Service implements fetching of information from Tezos nodes via JSON.
type Service struct {
	Client *RPCClient
}

// NetworkStats models global network bandwidth totals and usage in B/s.
type NetworkStats struct {
	TotalBytesSent int64 `json:"total_sent,string"`
	TotalBytesRecv int64 `json:"total_recv,string"`
	CurrentInflow  int64 `json:"current_inflow"`
	CurrentOutflow int64 `json:"current_outflow"`
}

// NetworkConnection models detailed information for one network connection.
type NetworkConnection struct {
	Incoming         bool              `json:"incoming"`
	PeerID           string            `json:"peer_id"`
	IDPoint          NetworkAddress    `json:"id_point"`
	RemoteSocketPort uint16            `json:"remote_socket_port"`
	Versions         []*NetworkVersion `json:"versions"`
	Private          bool              `json:"private"`
	LocalMetadata    NetworkMetadata   `json:"local_metadata"`
	RemoteMetadata   NetworkMetadata   `json:"remote_metadata"`
}

// NetworkAddress models a point's address and port.
type NetworkAddress struct {
	Addr string `json:"addr"`
	Port uint16 `json:"port"`
}

// NetworkVersion models a network-layer version of a node.
type NetworkVersion struct {
	Name  string `json:"name"`
	Major uint16 `json:"major"`
	Minor uint16 `json:"minor"`
}

// NetworkMetadata models metadata of a node.
type NetworkMetadata struct {
	DisableMempool bool `json:"disable_mempool"`
	PrivateNode    bool `json:"private_node"`
}

// BootstrappedBlock represents bootstrapped block stream message
type BootstrappedBlock struct {
	Block     string    `json:"block"`
	Timestamp time.Time `json:"timestamp"`
}

// NetworkConnectionTimestamp represents peer address with timestamp added
type NetworkConnectionTimestamp struct {
	NetworkAddress
	Timestamp time.Time
}

// UnmarshalJSON implements json.Unmarshaler
func (n *NetworkConnectionTimestamp) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &n.NetworkAddress, &n.Timestamp)
}

// NetworkPeer represents peer info
type NetworkPeer struct {
	PeerID                    string                      `json:"-"`
	Score                     int64                       `json:"score"`
	Trusted                   bool                        `json:"trusted"`
	ConnMetadata              *NetworkMetadata            `json:"conn_metadata"`
	State                     string                      `json:"state"`
	ReachableAt               *NetworkAddress             `json:"reachable_at"`
	Stat                      NetworkStats                `json:"stat"`
	LastEstablishedConnection *NetworkConnectionTimestamp `json:"last_established_connection"`
	LastSeen                  *NetworkConnectionTimestamp `json:"last_seen"`
	LastFailedConnection      *NetworkConnectionTimestamp `json:"last_failed_connection"`
	LastRejectedConnection    *NetworkConnectionTimestamp `json:"last_rejected_connection"`
	LastDisconnection         *NetworkConnectionTimestamp `json:"last_disconnection"`
	LastMiss                  *NetworkConnectionTimestamp `json:"last_miss"`
}

// networkPeerWithID is a heterogeneously encoded NetworkPeer with ID as a first array member
// See OperationAlt for details
type networkPeerWithID NetworkPeer

func (n *networkPeerWithID) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &n.PeerID, (*NetworkPeer)(n))
}

// NetworkPeerLogEntry represents peer log entry
type NetworkPeerLogEntry struct {
	NetworkAddress
	Kind      string    `json:"kind"`
	Timestamp time.Time `json:"timestamp"`
}

// NetworkPoint represents network point info
type NetworkPoint struct {
	Address                   string            `json:"-"`
	Trusted                   bool              `json:"trusted"`
	GreylistedUntil           time.Time         `json:"greylisted_until"`
	State                     NetworkPointState `json:"state"`
	P2PPeerID                 string            `json:"p2p_peer_id"`
	LastFailedConnection      time.Time         `json:"last_failed_connection"`
	LastRejectedConnection    *IDTimestamp      `json:"last_rejected_connection"`
	LastEstablishedConnection *IDTimestamp      `json:"last_established_connection"`
	LastDisconnection         *IDTimestamp      `json:"last_disconnection"`
	LastSeen                  *IDTimestamp      `json:"last_seen"`
	LastMiss                  time.Time         `json:"last_miss"`
}

// networkPointAlt is a heterogeneously encoded NetworkPoint with address as a first array member
// See OperationAlt for details
type networkPointAlt NetworkPoint

func (n *networkPointAlt) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &n.Address, (*NetworkPoint)(n))
}

// NetworkPointState represents point state
type NetworkPointState struct {
	EventKind string `json:"event_kind"`
	P2PPeerID string `json:"p2p_peer_id"`
}

// IDTimestamp represents peer id with timestamp
type IDTimestamp struct {
	ID        string
	Timestamp time.Time
}

// UnmarshalJSON implements json.Unmarshaler
func (i *IDTimestamp) UnmarshalJSON(data []byte) error {
	return unmarshalHeterogeneousJSONArray(data, &i.ID, &i.Timestamp)
}

// NetworkPointLogEntry represents point's log entry
type NetworkPointLogEntry struct {
	Kind      NetworkPointState `json:"kind"`
	Timestamp time.Time         `json:"timestamp"`
}

// MempoolOperations represents mempool operations
type MempoolOperations struct {
	Applied       []*Operation             `json:"applied"`
	Refused       []*OperationWithErrorAlt `json:"refused"`
	BranchRefused []*OperationWithErrorAlt `json:"branch_refused"`
	BranchDelayed []*OperationWithErrorAlt `json:"branch_delayed"`
	Unprocessed   []*OperationAlt          `json:"unprocessed"`
}

// InvalidBlock represents invalid block hash along with the errors that led to it being declared invalid
type InvalidBlock struct {
	Block string `json:"block"`
	Level int    `json:"level"`
	Error Errors `json:"error"`
}

type proposalsRPCResponse = [][]interface{}

// Just suppress UnmarshalJSON
type bigIntStr big.Int

func (z *bigIntStr) UnmarshalText(data []byte) error {
	return (*big.Int)(z).UnmarshalText(data)
}

func (z *bigIntStr) MarshalJSON() ([]byte, error) {
	return (*big.Int)(z).MarshalText()
}

func (z *bigIntStr) Int64() int64 {
	return (*big.Int)(z).Int64()
}

// GetNetworkStats returns current network stats https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat
func (s *Service) GetNetworkStats(ctx context.Context) (*NetworkStats, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/stat", nil)
	if err != nil {
		return nil, err
	}

	var stats NetworkStats
	if err = s.Client.Do(req, &stats); err != nil {
		return nil, err
	}
	return &stats, err
}

// GetNetworkConnections returns all network connections http://tezos.gitlab.io/mainnet/api/rpc.html#get-network-connections
func (s *Service) GetNetworkConnections(ctx context.Context) ([]*NetworkConnection, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/connections", nil)
	if err != nil {
		return nil, err
	}

	var conns []*NetworkConnection
	if err = s.Client.Do(req, &conns); err != nil {
		return nil, err
	}
	return conns, err
}

// GetNetworkPeers returns the list the peers the node ever met.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers
func (s *Service) GetNetworkPeers(ctx context.Context, filter string) ([]*NetworkPeer, error) {
	u := url.URL{
		Path: "/network/peers",
	}

	if filter != "" {
		q := url.Values{
			"filter": []string{filter},
		}
		u.RawQuery = q.Encode()
	}

	req, err := s.Client.NewRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var peers []*networkPeerWithID
	if err = s.Client.Do(req, &peers); err != nil {
		return nil, err
	}

	ret := make([]*NetworkPeer, len(peers))
	for i, p := range peers {
		ret[i] = (*NetworkPeer)(p)
	}

	return ret, err
}

// GetNetworkPeer returns details about a given peer.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id
func (s *Service) GetNetworkPeer(ctx context.Context, peerID string) (*NetworkPeer, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID, nil)
	if err != nil {
		return nil, err
	}

	var peer NetworkPeer
	if err = s.Client.Do(req, &peer); err != nil {
		return nil, err
	}
	peer.PeerID = peerID

	return &peer, err
}

// BanNetworkPeer blacklists the given peer.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-ban
func (s *Service) BanNetworkPeer(ctx context.Context, peerID string) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/ban", nil)
	if err != nil {
		return err
	}

	if err := s.Client.Do(req, nil); err != nil {
		return err
	}
	return nil
}

// TrustNetworkPeer used to trust a given peer permanently: the peer cannot be blocked (but its host IP still can).
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-trust
func (s *Service) TrustNetworkPeer(ctx context.Context, peerID string) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/trust", nil)
	if err != nil {
		return err
	}

	if err := s.Client.Do(req, nil); err != nil {
		return err
	}
	return nil
}

// GetNetworkPeerBanned checks if a given peer is blacklisted or greylisted.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-banned
func (s *Service) GetNetworkPeerBanned(ctx context.Context, peerID string) (bool, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/banned", nil)
	if err != nil {
		return false, err
	}

	var banned bool
	if err = s.Client.Do(req, &banned); err != nil {
		return false, err
	}

	return banned, err
}

// GetNetworkPeerLog monitors network events related to a given peer.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-log
func (s *Service) GetNetworkPeerLog(ctx context.Context, peerID string) ([]*NetworkPeerLogEntry, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/log", nil)
	if err != nil {
		return nil, err
	}

	var log []*NetworkPeerLogEntry
	if err = s.Client.Do(req, &log); err != nil {
		return nil, err
	}

	return log, err
}

// MonitorNetworkPeerLog monitors network events related to a given peer.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-log
func (s *Service) MonitorNetworkPeerLog(ctx context.Context, peerID string, results chan<- []*NetworkPeerLogEntry) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/peers/"+peerID+"/log?monitor", nil)
	if err != nil {
		return err
	}

	return s.Client.Do(req, results)
}

// GetNetworkPoints returns list the pool of known `IP:port` used for establishing P2P connections.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-points
func (s *Service) GetNetworkPoints(ctx context.Context, filter string) ([]*NetworkPoint, error) {
	u := url.URL{
		Path: "/network/points",
	}

	if filter != "" {
		q := url.Values{
			"filter": []string{filter},
		}
		u.RawQuery = q.Encode()
	}

	req, err := s.Client.NewRequest(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}

	var points []*networkPointAlt
	if err = s.Client.Do(req, &points); err != nil {
		return nil, err
	}

	ret := make([]*NetworkPoint, len(points))
	for i, p := range points {
		ret[i] = (*NetworkPoint)(p)
	}

	return ret, err
}

// GetNetworkPoint returns details about a given `IP:addr`.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-points-point
func (s *Service) GetNetworkPoint(ctx context.Context, address string) (*NetworkPoint, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/points/"+address, nil)
	if err != nil {
		return nil, err
	}

	var point NetworkPoint
	if err = s.Client.Do(req, &point); err != nil {
		return nil, err
	}
	point.Address = address

	return &point, err
}

// ConnectToNetworkPoint used to connect to a peer.
// https://tezos.gitlab.io/mainnet/api/rpc.html#put-network-points-point
func (s *Service) ConnectToNetworkPoint(ctx context.Context, address string, timeout time.Duration) error {
	u := url.URL{
		Path: "/network/points/" + address,
	}

	if timeout > 0 {
		q := url.Values{
			"timeout": []string{fmt.Sprintf("%f", float64(timeout)/float64(time.Second))},
		}
		u.RawQuery = q.Encode()
	}

	req, err := s.Client.NewRequest(ctx, http.MethodPut, u.String(), &struct{}{})
	if err != nil {
		return err
	}

	if err := s.Client.Do(req, nil); err != nil {
		return err
	}

	return nil
}

// BanNetworkPoint blacklists the given address.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-points-point-ban
func (s *Service) BanNetworkPoint(ctx context.Context, address string) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/points/"+address+"/ban", nil)
	if err != nil {
		return err
	}

	if err := s.Client.Do(req, nil); err != nil {
		return err
	}
	return nil
}

// TrustNetworkPoint used to trust a given address permanently. Connections from this address can still be closed on authentication if the peer is blacklisted or greylisted.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-points-point-trust
func (s *Service) TrustNetworkPoint(ctx context.Context, address string) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/points/"+address+"/trust", nil)
	if err != nil {
		return err
	}

	if err := s.Client.Do(req, nil); err != nil {
		return err
	}
	return nil
}

// GetNetworkPointBanned check is a given address is blacklisted or greylisted.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-points-point-banned
func (s *Service) GetNetworkPointBanned(ctx context.Context, address string) (bool, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/points/"+address+"/banned", nil)
	if err != nil {
		return false, err
	}

	var banned bool
	if err = s.Client.Do(req, &banned); err != nil {
		return false, err
	}

	return banned, err
}

// GetNetworkPointLog monitors network events related to an `IP:addr`.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-log
func (s *Service) GetNetworkPointLog(ctx context.Context, address string) ([]*NetworkPointLogEntry, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/points/"+address+"/log", nil)
	if err != nil {
		return nil, err
	}

	var log []*NetworkPointLogEntry
	if err = s.Client.Do(req, &log); err != nil {
		return nil, err
	}

	return log, err
}

// MonitorNetworkPointLog monitors network events related to an `IP:addr`.
// https://tezos.gitlab.io/mainnet/api/rpc.html#get-network-peers-peer-id-log
func (s *Service) MonitorNetworkPointLog(ctx context.Context, address string, results chan<- []*NetworkPointLogEntry) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/points/"+address+"/log?monitor", nil)
	if err != nil {
		return err
	}

	return s.Client.Do(req, results)
}

// GetDelegateBalance returns a delegate's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-delegates-pkh-balance
func (s *Service) GetDelegateBalance(ctx context.Context, chainID string, blockID string, pkh string) (*big.Int, error) {
	u := "/chains/" + chainID + "/blocks/" + blockID + "/context/delegates/" + pkh + "/balance"
	req, err := s.Client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var balance bigIntStr
	if err := s.Client.Do(req, &balance); err != nil {
		return nil, err
	}

	return (*big.Int)(&balance), nil
}

// GetContractBalance returns a contract's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-contracts-contract-id-balance
func (s *Service) GetContractBalance(ctx context.Context, chainID string, blockID string, contractID string) (*big.Int, error) {
	u := "/chains/" + chainID + "/blocks/" + blockID + "/context/contracts/" + contractID + "/balance"
	req, err := s.Client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var balance bigIntStr
	if err := s.Client.Do(req, &balance); err != nil {
		return nil, err
	}

	return (*big.Int)(&balance), nil
}

// GetBootstrapped reads from the bootstrapped blocks stream http://tezos.gitlab.io/mainnet/api/rpc.html#get-monitor-bootstrapped
func (s *Service) GetBootstrapped(ctx context.Context, results chan<- *BootstrappedBlock) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/monitor/bootstrapped", nil)
	if err != nil {
		return err
	}

	return s.Client.Do(req, results)
}

// GetMonitorHeads reads from the heads blocks stream https://tezos.gitlab.io/mainnet/api/rpc.html#get-monitor-heads-chain-id
func (s *Service) GetMonitorHeads(ctx context.Context, chainID string, results chan<- *MonitorBlock) error {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/monitor/heads/"+chainID, nil)
	if err != nil {
		return err
	}

	return s.Client.Do(req, results)
}

// GetMempoolPendingOperations returns mempool pending operations
func (s *Service) GetMempoolPendingOperations(ctx context.Context, chainID string) (*MempoolOperations, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/mempool/pending_operations", nil)
	if err != nil {
		return nil, err
	}

	var ops MempoolOperations
	if err := s.Client.Do(req, &ops); err != nil {
		return nil, err
	}

	return &ops, nil
}

// MonitorMempoolOperations monitors mempool pending operations.
// The connection is closed after every new block.
func (s *Service) MonitorMempoolOperations(ctx context.Context, chainID, filter string, results chan<- []*Operation) error {
	if filter == "" {
		filter = "applied"
	}

	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/mempool/monitor_operations?"+filter, nil)
	if err != nil {
		return err
	}

	return s.Client.Do(req, results)
}

// GetInvalidBlocks lists blocks that have been declared invalid along with the errors that led to them being declared invalid.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-chains-chain-id-invalid-blocks
func (s *Service) GetInvalidBlocks(ctx context.Context, chainID string) ([]*InvalidBlock, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/invalid_blocks", nil)
	if err != nil {
		return nil, err
	}

	var invalidBlocks []*InvalidBlock
	if err := s.Client.Do(req, &invalidBlocks); err != nil {
		return nil, err
	}

	return invalidBlocks, nil
}

// GetBlock returns information about a Tezos block
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id
func (s *Service) GetBlock(ctx context.Context, chainID, blockID string) (*Block, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID, nil)
	if err != nil {
		return nil, err
	}

	var block Block
	if err := s.Client.Do(req, &block); err != nil {
		return nil, err
	}

	return &block, nil
}

// GetBallotList returns ballots casted so far during a voting period.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-ballot-list
func (s *Service) GetBallotList(ctx context.Context, chainID, blockID string) ([]*Ballot, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/ballot_list", nil)
	if err != nil {
		return nil, err
	}

	var ballots []*Ballot
	if err := s.Client.Do(req, &ballots); err != nil {
		return nil, err
	}

	return ballots, nil
}

// GetBallots returns sum of ballots casted so far during a voting period.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-ballots
func (s *Service) GetBallots(ctx context.Context, chainID, blockID string) (*Ballots, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/ballots", nil)
	if err != nil {
		return nil, err
	}

	var ballots Ballots
	if err := s.Client.Do(req, &ballots); err != nil {
		return nil, err
	}

	return &ballots, nil
}

// GetBallotListings returns a list of delegates with their voting weight, in number of rolls.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-listings
func (s *Service) GetBallotListings(ctx context.Context, chainID, blockID string) ([]*BallotListing, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/listings", nil)
	if err != nil {
		return nil, err
	}

	var listings []*BallotListing
	if err := s.Client.Do(req, &listings); err != nil {
		return nil, err
	}

	return listings, nil
}

// GetProposals returns a list of proposals with number of supporters.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-proposals
func (s *Service) GetProposals(ctx context.Context, chainID, blockID string) ([]*Proposal, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/proposals", nil)
	if err != nil {
		return nil, err
	}

	var proposalsResp proposalsRPCResponse
	if err := s.Client.Do(req, &proposalsResp); err != nil {
		return nil, err
	}

	proposals := make([]*Proposal, len(proposalsResp))

	for i, proposalResp := range proposalsResp {
		if len(proposalResp) == 2 {
			proposal := &Proposal{}
			if propHash, ok := proposalResp[0].(string); ok {
				proposal.ProposalHash = propHash
			} else {
				return nil, fmt.Errorf("Malformed request ProposalHash was expected to be a string but got: %v", proposalResp[0])
			}
			if supporters, ok := proposalResp[1].(float64); ok {
				proposal.SupporterCount = int(supporters)
			} else {
				return nil, fmt.Errorf("Malformed request SupporterCount was expected to be a float but got: %v", proposalResp[1])
			}
			proposals[i] = proposal
		} else {
			return nil, fmt.Errorf("Malformed request Proposal is expected to be tuple of size 2")
		}
	}

	return proposals, nil
}

// GetCurrentProposals returns the current proposal under evaluation.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-current-proposal
func (s *Service) GetCurrentProposals(ctx context.Context, chainID, blockID string) (string, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/current_proposal", nil)
	if err != nil {
		return "", err
	}

	var currentProposal string
	if err := s.Client.Do(req, &currentProposal); err != nil {
		return "", err
	}

	return currentProposal, nil
}

// GetCurrentQuorum returns the current expected quorum.
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-current-quorum
func (s *Service) GetCurrentQuorum(ctx context.Context, chainID, blockID string) (int, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/current_quorum", nil)
	if err != nil {
		return -1, err
	}

	var currentQuorum int
	if err := s.Client.Do(req, &currentQuorum); err != nil {
		return -1, err
	}

	return currentQuorum, nil
}

// GetCurrentPeriodKind returns the current period kind
// https://tezos.gitlab.io/alphanet/api/rpc.html#get-block-id-votes-current-period-kind
func (s *Service) GetCurrentPeriodKind(ctx context.Context, chainID, blockID string) (PeriodKind, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/chains/"+chainID+"/blocks/"+blockID+"/votes/current_period_kind", nil)
	if err != nil {
		return "", err
	}

	var periodKind PeriodKind
	if err := s.Client.Do(req, &periodKind); err != nil {
		return "", err
	}

	return periodKind, nil
}
