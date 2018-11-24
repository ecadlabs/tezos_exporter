package tezos

import (
	"context"
	"encoding/json"
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
	Incoming         bool             `json:"incoming"`
	PeerID           string           `json:"peer_id"`
	IDPoint          NetworkAddress   `json:"id_point"`
	RemoteSocketPort uint16           `json:"remote_socket_port"`
	Versions         []NetworkVersion `json:"versions"`
	Private          bool             `json:"private"`
	LocalMetadata    NetworkMetadata  `json:"local_metadata"`
	RemoteMetadata   NetworkMetadata  `json:"remote_metadata"`
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

type networkPointWithAddress NetworkPoint

func (n *networkPointWithAddress) UnmarshalJSON(data []byte) error {
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

// GetNetworkStats returns current network stats https://tezos.gitlab.io/betanet/api/rpc.html#get-network-stat
func (s *Service) GetNetworkStats(ctx context.Context) (*NetworkStats, error) {
	req, err := s.Client.NewRequest(ctx, http.MethodGet, "/network/stat", nil)
	if err != nil {
		return nil, err
	}

	var stats NetworkStats
	if err := s.Client.Do(req, &stats); err != nil {
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
	if err := s.Client.Do(req, &conns); err != nil {
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
	if err := s.Client.Do(req, &peers); err != nil {
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
	if err := s.Client.Do(req, &peer); err != nil {
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
	if err := s.Client.Do(req, &banned); err != nil {
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
	if err := s.Client.Do(req, &log); err != nil {
		return nil, err
	}

	return log, err
}

// MonitorNetworkPeerLog monitor network events related to a given peer.
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

	var points []*networkPointWithAddress
	if err := s.Client.Do(req, &points); err != nil {
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
	if err := s.Client.Do(req, &point); err != nil {
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
	if err := s.Client.Do(req, &banned); err != nil {
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
	if err := s.Client.Do(req, &log); err != nil {
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

type bigInt big.Int

func (z *bigInt) UnmarshalJSON(data []byte) error {
	var s string
	if err := json.Unmarshal(data, &s); err != nil {
		return err
	}

	return (*big.Int)(z).UnmarshalText([]byte(s))
}

// GetDelegateBalance returns a delegate's balance http://tezos.gitlab.io/mainnet/api/rpc.html#get-block-id-context-delegates-pkh-balance
func (s *Service) GetDelegateBalance(ctx context.Context, chainID string, blockID string, pkh string) (*big.Int, error) {
	u := "/chains/" + chainID + "/blocks/" + blockID + "/context/delegates/" + pkh + "/balance"
	req, err := s.Client.NewRequest(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil, err
	}

	var balance bigInt
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

	var balance bigInt
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
