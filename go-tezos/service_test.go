package tezos

import (
	"context"
	"io/ioutil"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func timeMustUnmarshalText(text string) (t time.Time) {
	if err := t.UnmarshalText([]byte(text)); err != nil {
		panic(err)
	}
	return
}

func TestServiceGetMethods(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		get             func(s *Service) (interface{}, error)
		respFixture     string
		respInline      string
		respStatus      int
		respContentType string
		expectedPath    string
		expectedQuery   string
		expectedValue   interface{}
		expectedMethod  string
		errMsg          string
		errType         interface{}
	}{
		{
			get:             func(s *Service) (interface{}, error) { return s.GetNetworkStats(ctx) },
			respFixture:     "fixtures/network/stat.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			expectedValue: &NetworkStats{
				TotalBytesSent: 291690080,
				TotalBytesRecv: 532639553,
				CurrentInflow:  23596,
				CurrentOutflow: 14972,
			},
		},
		{
			get:             func(s *Service) (interface{}, error) { return s.GetNetworkConnections(ctx) },
			respFixture:     "fixtures/network/connections.json",
			respContentType: "application/json",
			expectedPath:    "/network/connections",
			expectedValue:   []*NetworkConnection{{Incoming: false, PeerID: "idt5qvkLiJ15rb6yJU1bjpGmdyYnPJ", IDPoint: NetworkAddress{Addr: "::ffff:34.253.64.43", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []*NetworkVersion{{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: false, LocalMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}, RemoteMetadata: NetworkMetadata{DisableMempool: false, PrivateNode: false}}, {Incoming: true, PeerID: "ids8VJTHEuyND6B8ahGgXPAJ3BDp1c", IDPoint: NetworkAddress{Addr: "::ffff:176.31.255.202", Port: 0x2604}, RemoteSocketPort: 0x2604, Versions: []*NetworkVersion{{Name: "TEZOS_ALPHANET_2018-07-31T16:22:39Z", Major: 0x0, Minor: 0x0}}, Private: true, LocalMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}, RemoteMetadata: NetworkMetadata{DisableMempool: true, PrivateNode: true}}},
		},
		{
			get:             func(s *Service) (interface{}, error) { return s.GetNetworkPeers(ctx, "") },
			respFixture:     "fixtures/network/peers.json",
			respContentType: "application/json",
			expectedPath:    "/network/peers",
			expectedValue:   []*NetworkPeer{{PeerID: "idrnHcGMrFxiYsmxf5Cqd6NhUTUU8X", ConnMetadata: &NetworkMetadata{}, State: "running", ReachableAt: &NetworkAddress{Addr: "::ffff:45.79.146.133", Port: 39732}, Stat: NetworkStats{TotalBytesSent: 4908012, TotalBytesRecv: 14560268, CurrentInflow: 66, CurrentOutflow: 177}, LastEstablishedConnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:45.79.146.133", Port: 39732}, Timestamp: timeMustUnmarshalText("2018-11-13T20:56:14Z")}, LastSeen: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:45.79.146.133", Port: 39732}, Timestamp: timeMustUnmarshalText("2018-11-13T20:56:14Z")}, LastRejectedConnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:45.79.146.133", Port: 39732}, Timestamp: timeMustUnmarshalText("2018-11-13T15:22:41Z")}, LastDisconnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:45.79.146.133", Port: 39732}, Timestamp: timeMustUnmarshalText("2018-11-13T18:04:12Z")}, LastMiss: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:45.79.146.133", Port: 39732}, Timestamp: timeMustUnmarshalText("2018-11-13T18:04:12Z")}}, {PeerID: "idsXeq1zboupwXXDdDDiWhBjimeJe3", State: "disconnected", LastEstablishedConnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.155.17.238", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-13T17:57:18Z")}, LastSeen: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.155.17.238", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-13T19:48:57Z")}, LastDisconnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.155.17.238", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-13T19:48:57Z")}, LastMiss: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.155.17.238", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-13T19:48:57Z")}}},
		},
		{
			get:             func(s *Service) (interface{}, error) { return s.GetNetworkPeer(ctx, "idtTZmNapGXAcfbnPoAcDz6J2xCHZZ") },
			respFixture:     "fixtures/network/peer.json",
			respContentType: "application/json",
			expectedPath:    "/network/peers/idtTZmNapGXAcfbnPoAcDz6J2xCHZZ",
			expectedValue:   &NetworkPeer{PeerID: "idtTZmNapGXAcfbnPoAcDz6J2xCHZZ", ConnMetadata: &NetworkMetadata{}, State: "running", ReachableAt: &NetworkAddress{Addr: "::ffff:104.248.233.63", Port: 9732}, Stat: NetworkStats{TotalBytesSent: 1196571, TotalBytesRecv: 1302211, CurrentInflow: 0, CurrentOutflow: 1}, LastEstablishedConnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.248.233.63", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-14T11:47:07Z")}, LastSeen: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.248.233.63", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-14T11:47:07Z")}, LastDisconnection: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.248.233.63", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-14T11:44:57Z")}, LastMiss: &NetworkConnectionTimestamp{NetworkAddress: NetworkAddress{Addr: "::ffff:104.248.233.63", Port: 9732}, Timestamp: timeMustUnmarshalText("2018-11-14T11:44:57Z")}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetNetworkPeerBanned(ctx, "idtTZmNapGXAcfbnPoAcDz6J2xCHZZ")
			},
			respInline:      "false",
			respContentType: "application/json",
			expectedPath:    "/network/peers/idtTZmNapGXAcfbnPoAcDz6J2xCHZZ/banned",
			expectedValue:   false,
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetNetworkPeerLog(ctx, "idrPSsREFE1MV1161ybEpaebFwgYWE")
			},
			respFixture:     "fixtures/network/peer_log.json",
			respContentType: "application/json",
			expectedPath:    "/network/peers/idrPSsREFE1MV1161ybEpaebFwgYWE/log",
			expectedValue:   []*NetworkPeerLogEntry{{NetworkAddress: NetworkAddress{Addr: "::ffff:13.81.43.51", Port: 9732}, Kind: "incoming_request", Timestamp: timeMustUnmarshalText("2018-11-13T15:35:17Z")}, {NetworkAddress: NetworkAddress{Addr: "::ffff:13.81.43.51", Port: 9732}, Kind: "connection_established", Timestamp: timeMustUnmarshalText("2018-11-13T15:35:19Z")}, {NetworkAddress: NetworkAddress{Addr: "::ffff:13.81.43.51", Port: 9732}, Kind: "external_disconnection", Timestamp: timeMustUnmarshalText("2018-11-13T18:02:51Z")}, {NetworkAddress: NetworkAddress{Addr: "::ffff:13.81.43.51", Port: 9732}, Kind: "incoming_request", Timestamp: timeMustUnmarshalText("2018-11-13T20:56:14Z")}, {NetworkAddress: NetworkAddress{Addr: "::ffff:13.81.43.51", Port: 9732}, Kind: "connection_established", Timestamp: timeMustUnmarshalText("2018-11-13T20:56:15Z")}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				ch := make(chan []*NetworkPeerLogEntry, 100)
				if err := s.MonitorNetworkPeerLog(ctx, "idsBATisQfJu7d6vCLY4CP66dKj7CQ", ch); err != nil {
					return nil, err
				}
				close(ch)

				var res [][]*NetworkPeerLogEntry
				for b := range ch {
					res = append(res, b)
				}
				return res, nil
			},
			respFixture:     "fixtures/network/peer_log.chunked",
			respContentType: "application/json",
			expectedPath:    "/network/peers/idsBATisQfJu7d6vCLY4CP66dKj7CQ/log",
			expectedValue:   [][]*NetworkPeerLogEntry{{&NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "incoming_request", Timestamp: timeMustUnmarshalText("2018-11-13T15:20:14Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "connection_established", Timestamp: timeMustUnmarshalText("2018-11-13T15:20:14Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "external_disconnection", Timestamp: timeMustUnmarshalText("2018-11-13T16:30:08Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "incoming_request", Timestamp: timeMustUnmarshalText("2018-11-13T16:39:20Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "connection_established", Timestamp: timeMustUnmarshalText("2018-11-13T16:39:20Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "external_disconnection", Timestamp: timeMustUnmarshalText("2018-11-13T19:48:58Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "incoming_request", Timestamp: timeMustUnmarshalText("2018-11-13T20:56:30Z")}, &NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "connection_established", Timestamp: timeMustUnmarshalText("2018-11-13T20:56:30Z")}}, {&NetworkPeerLogEntry{NetworkAddress: NetworkAddress{Addr: "::ffff:51.15.242.114", Port: 9732}, Kind: "external_disconnection", Timestamp: timeMustUnmarshalText("2018-11-13T22:25:07Z")}}},
		},
		{
			get:             func(s *Service) (interface{}, error) { return s.GetNetworkPoints(ctx, "") },
			respFixture:     "fixtures/network/points.json",
			respContentType: "application/json",
			expectedPath:    "/network/points",
			expectedValue:   []*NetworkPoint{{Address: "73.247.92.150:9732", Trusted: false, GreylistedUntil: timeMustUnmarshalText("2018-11-14T19:01:28Z"), State: NetworkPointState{EventKind: "disconnected"}, LastFailedConnection: timeMustUnmarshalText("2018-11-14T19:01:16Z"), LastMiss: timeMustUnmarshalText("2018-11-14T19:01:16Z")}, {Address: "40.119.159.28:9732", Trusted: false, GreylistedUntil: timeMustUnmarshalText("2018-11-14T16:24:57Z"), State: NetworkPointState{EventKind: "running", P2PPeerID: "ids496Ey2BKHVJYZdsk72XCwbZteTj"}, P2PPeerID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", LastFailedConnection: timeMustUnmarshalText("2018-11-14T11:47:13Z"), LastRejectedConnection: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T12:03:11Z")}, LastEstablishedConnection: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T16:48:56Z")}, LastDisconnection: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T16:23:57Z")}, LastSeen: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T16:48:56Z")}, LastMiss: timeMustUnmarshalText("2018-11-14T16:23:57Z")}},
		},
		{
			get:             func(s *Service) (interface{}, error) { return s.GetNetworkPoint(ctx, "40.119.159.28:9732") },
			respFixture:     "fixtures/network/point.json",
			respContentType: "application/json",
			expectedPath:    "/network/points/40.119.159.28:9732",
			expectedValue:   &NetworkPoint{Address: "40.119.159.28:9732", Trusted: false, GreylistedUntil: timeMustUnmarshalText("2018-11-14T16:24:57Z"), State: NetworkPointState{EventKind: "running", P2PPeerID: "ids496Ey2BKHVJYZdsk72XCwbZteTj"}, P2PPeerID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", LastFailedConnection: timeMustUnmarshalText("2018-11-14T11:47:13Z"), LastRejectedConnection: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T12:03:11Z")}, LastEstablishedConnection: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T16:48:56Z")}, LastDisconnection: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T16:23:57Z")}, LastSeen: &IDTimestamp{ID: "ids496Ey2BKHVJYZdsk72XCwbZteTj", Timestamp: timeMustUnmarshalText("2018-11-14T16:48:56Z")}, LastMiss: timeMustUnmarshalText("2018-11-14T16:23:57Z")},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetNetworkPointBanned(ctx, "40.119.159.28:9732")
			},
			respInline:      "false",
			respContentType: "application/json",
			expectedPath:    "/network/points/40.119.159.28:9732/banned",
			expectedValue:   false,
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetNetworkPointLog(ctx, "34.255.45.196:9732")
			},
			respFixture:     "fixtures/network/point_log.json",
			respContentType: "application/json",
			expectedPath:    "/network/points/34.255.45.196:9732/log",
			expectedValue:   []*NetworkPointLogEntry{{Kind: NetworkPointState{EventKind: "outgoing_request"}, Timestamp: timeMustUnmarshalText("2018-11-15T17:56:18Z")}, {Kind: NetworkPointState{EventKind: "accepting_request", P2PPeerID: "idrBJarh4t32gN9s52kxMWmeSi76Jk"}, Timestamp: timeMustUnmarshalText("2018-11-15T17:56:19Z")}, {Kind: NetworkPointState{EventKind: "rejecting_request", P2PPeerID: "idrBJarh4t32gN9s52kxMWmeSi76Jk"}, Timestamp: timeMustUnmarshalText("2018-11-15T17:56:19Z")}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				ch := make(chan []*NetworkPointLogEntry, 100)
				if err := s.MonitorNetworkPointLog(ctx, "80.214.69.170:9732", ch); err != nil {
					return nil, err
				}
				close(ch)

				var res [][]*NetworkPointLogEntry
				for b := range ch {
					res = append(res, b)
				}

				return res, nil
			},
			respFixture:     "fixtures/network/point_log.chunked",
			respContentType: "application/json",
			expectedPath:    "/network/points/80.214.69.170:9732/log",
			expectedValue:   [][]*NetworkPointLogEntry{{&NetworkPointLogEntry{Kind: NetworkPointState{EventKind: "outgoing_request"}, Timestamp: timeMustUnmarshalText("2018-11-15T18:00:39Z")}, &NetworkPointLogEntry{Kind: NetworkPointState{EventKind: "request_rejected"}, Timestamp: timeMustUnmarshalText("2018-11-15T18:00:49Z")}}, {&NetworkPointLogEntry{Kind: NetworkPointState{EventKind: "outgoing_request"}, Timestamp: timeMustUnmarshalText("2018-11-15T18:16:18Z")}}, {&NetworkPointLogEntry{Kind: NetworkPointState{EventKind: "request_rejected"}, Timestamp: timeMustUnmarshalText("2018-11-15T18:16:28Z")}}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return nil, s.ConnectToNetworkPoint(ctx, "80.214.69.170:9732", 10*time.Second)
			},
			respInline:      "{}",
			respContentType: "application/json",
			expectedPath:    "/network/points/80.214.69.170:9732",
			expectedMethod:  "PUT",
			expectedQuery:   "timeout=10.000000",
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetDelegateBalance(ctx, "main", "head", "tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5")
			},
			respFixture:     "fixtures/block/delegate_balance.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/context/delegates/tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5/balance",
			expectedValue:   big.NewInt(13490453135591),
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetContractBalance(ctx, "main", "head", "tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5")
			},
			respFixture:     "fixtures/block/contract_balance.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/context/contracts/tz3WXYtyDUNL91qfiCJtVUX746QpNv5i5ve5/balance",
			expectedValue:   big.NewInt(4700354460878),
		},
		{
			get: func(s *Service) (interface{}, error) {
				ch := make(chan *BootstrappedBlock, 100)
				if err := s.MonitorBootstrapped(ctx, ch); err != nil {
					return nil, err
				}
				close(ch)

				var res []*BootstrappedBlock
				for b := range ch {
					res = append(res, b)
				}
				return res, nil
			},
			respFixture:     "fixtures/monitor/bootstrapped.chunked",
			respContentType: "application/json",
			expectedPath:    "/monitor/bootstrapped",
			expectedValue: []*BootstrappedBlock{
				{Block: "BLgz6z8w5bYtn2AAEmsfMD3aH9o8SUnVygUpVUsCe6dkRpEt5Qy", Timestamp: timeMustUnmarshalText("2018-09-17T00:46:12Z")},
				{Block: "BLc3Y6zsb7PT6QnScu8VKcUPGkCoeCLPWLVTQoQjk5QQ7pbmHs5", Timestamp: timeMustUnmarshalText("2018-09-17T00:46:42Z")},
				{Block: "BKiqiXgqAPHX4bRzk2p1jEKHijaxLPdcQi8hqVfGhBwngcticEk", Timestamp: timeMustUnmarshalText("2018-09-17T00:48:32Z")},
			},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetMempoolPendingOperations(ctx, "main")
			},
			respFixture:     "fixtures/block/pending_operations.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/mempool/pending_operations",
			expectedValue:   &MempoolOperations{Applied: []*Operation{{Hash: "opLHEC3xm8qPRP9g44oBpB45RzRVUoMX1NsX75sKKtNvA8pvSm2", Branch: "BMLvebSvhTyZ7GG2vykV8hpGEc8egzcwn9fc3JJKrtCk8FssT9M", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 208806}}, Signature: "sigtTW5Y3xQaTKo5vEiqr8zG4YnPv7GbVbUgo7XYw7UZduz9jvdxzFbKUmftKFsFGH1UEZBbxyhyH5DLUUMh5KrQ3MENzUwC"}, {Hash: "ooSEFHRfArRSjeWhHhcmBa5aL2E3MqsN1HucCm3xiR2gLuzGSYN", Branch: "BMLvebSvhTyZ7GG2vykV8hpGEc8egzcwn9fc3JJKrtCk8FssT9M", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 208806}}, Signature: "sigeVFaHCGk9S6P9MhNNyZjHMcfPgYZw5cTwejtbGDEZdp58XKcxVkP3CFCKiPHesiEDqCxvrPGHZUpQLNmmqaSgrmv1ePNZ"}}, Refused: []*OperationWithErrorAlt{}, BranchRefused: []*OperationWithErrorAlt{}, BranchDelayed: []*OperationWithErrorAlt{{Operation: Operation{Protocol: "PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt", Hash: "oo1Z19oCkTWibLp7mJwFKP3UFVxuf6eV1iNWwJS7gZs8uZbrduS", Branch: "BMTSuKyFBhgmD7e3UDt9jLtjC2ftTUosTGEiiYc61Lu6F3xSkvJ", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 208804}}, Signature: "sigZXm4SGNcHwh5qsfjsFYmhSCwtimifq4EPje5rnJxvNDkymC2o3Yv8cJWgug3dDxiQWDexRDeBBu8Pf5qFxA6SckKypiau"}, Error: Errors{&GenericError{Kind: "temporary", ID: "proto.002-PsYLVpVv.operation.wrong_endorsement_predecessor"}}}, {Operation: Operation{Protocol: "PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt", Hash: "ooCaHemWe76uiBLDUXY2uhbhuiyLG7w7rqUFaJPxr7v56z6DVPS", Branch: "BL1pULCBFDJkqDHmYqK8yrVM3mHQHi72JFg6dT5qJ96ncjDbPpn", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 208773}}, Signature: "sigpkWpkY25KDBo7YcaLYx5Q61ypcfFWXjXgvbMG6uFrnStboCxCoCnJbDNri7CGzad35zLUvXCVxu2uj4WBSPgfxsnGKUBn"}, Error: Errors{&GenericError{Kind: "temporary", ID: "proto.002-PsYLVpVv.operation.wrong_endorsement_predecessor"}}}}, Unprocessed: []*OperationAlt{}},
		},
		// Handling 5xx errors from the Tezos node with RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetNetworkStats(ctx)
			},
			respStatus:      500,
			respFixture:     "fixtures/error.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			errMsg:          `tezos: kind = "permanent", id = "proto.002-PsYLVpVv.context.storage_error"`,
			errType:         (*rpcError)(nil),
		},
		// Handling 5xx errors from the Tezos node with empty RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetNetworkStats(ctx)
			},
			respStatus:      500,
			respFixture:     "fixtures/empty_error.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			errMsg:          `tezos: empty error response`,
			errType:         (*plainError)(nil),
		},
		// Handling 5xx errors from the Tezos node with malformed RPC error information.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetNetworkStats(ctx)
			},
			respStatus:      500,
			respFixture:     "fixtures/malformed_error.json",
			respContentType: "application/json",
			expectedPath:    "/network/stat",
			errMsg:          `tezos: error decoding RPC error: invalid character ',' looking for beginning of value`,
			errType:         (*plainError)(nil),
		},
		// Handling unexpected response status codes.
		{
			get: func(s *Service) (interface{}, error) {
				// Doesn't matter which Get* method we call here, as long as it calls RPCClient.Get
				// in the implementation.
				return s.GetNetworkStats(ctx)
			},
			respStatus:   404,
			respFixture:  "fixtures/empty.json",
			expectedPath: "/network/stat",
			errMsg:       `tezos: HTTP status 404`,
			errType:      (*httpError)(nil),
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetInvalidBlocks(ctx, "main")
			},
			respFixture:     "fixtures/chains/invalid_blocks.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/invalid_blocks",
			expectedValue:   []*InvalidBlock{{Block: "BM31cpbqfXu3WNYLQ8Tch21tXjcnwbyFzvcqohHL1BSnkhnhzwp", Level: 42, Error: Errors{}}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetBlock(ctx, "main", "BLnoArJNPCyYFK2z3Mnomi36Jo3FwrjriJ6hvzgTJGYYDKEkDXm")
			},
			respFixture:     "fixtures/chains/block.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/BLnoArJNPCyYFK2z3Mnomi36Jo3FwrjriJ6hvzgTJGYYDKEkDXm",
			expectedValue:   &Block{Protocol: "PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt", ChainID: "NetXZUqeBjDnWde", Hash: "BLnoArJNPCyYFK2z3Mnomi36Jo3FwrjriJ6hvzgTJGYYDKEkDXm", Header: RawBlockHeader{Level: 219133, Proto: 1, Predecessor: "BLNWdEensT9MFq8pkDwjHfGVFsV1reYUhVcMAVzq3LCMS1WdKZ8", Timestamp: timeMustUnmarshalText("2018-11-27T17:49:57Z"), ValidationPass: 4, OperationsHash: "LLoZamNeucV8tqPAcqJQYsNEsMwnCuL1xu1kJMiGFCx9MBVCGcWJF", Fitness: []HexBytes{{0x0}, {0x0, 0x0, 0x0, 0x0, 0x0, 0x5a, 0x12, 0x5f}}, Context: "CoW5zHjWVHfUAbSgzqnZ938eDXG37P9oJVn3Lb3NyQJBheUDvdVf", ProofOfWorkNonce: HexBytes{0x7d, 0x94, 0x95, 0x82, 0xfe, 0x2, 0x48, 0x62}, Signature: "sigktdiZpdykWEjgeTB3N1qFJ5bsh3SxVNB8wc5FAutbJPG7puWQAPrxwL6BZPJVKLRj2uLnCw54Akx4KA48DS5Jg8tthCLY"}, Metadata: BlockHeaderMetadata{Protocol: "PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt", NextProtocol: "PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt", TestChainStatus: &NotRunningTestChainStatus{GenericTestChainStatus: GenericTestChainStatus{Status: "not_running"}}, MaxOperationsTTL: 60, MaxOperationDataLength: 16384, MaxBlockHeaderLength: 238, MaxOperationListLength: []*MaxOperationListLength{{MaxSize: 32768, MaxOp: 32}}, Baker: "tz3gN8NTLNLJg5KRsUU47NHNVHbdhcFXjjaB", Level: BlockHeaderMetadataLevel{Level: 219133, LevelPosition: 219132, Cycle: 106, CyclePosition: 2044, VotingPeriod: 6, VotingPeriodPosition: 22524, ExpectedCommitment: false}, VotingPeriodKind: "proposal", ConsumedGas: &BigInt{}, Deactivated: []string{}, BalanceUpdates: BalanceUpdates{&ContractBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: "contract", Change: -512000000}, Contract: "tz3gN8NTLNLJg5KRsUU47NHNVHbdhcFXjjaB"}, &FreezerBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: "freezer", Change: 512000000}, Category: "deposits", Delegate: "tz3gN8NTLNLJg5KRsUU47NHNVHbdhcFXjjaB", Level: 106}}}, Operations: [][]*Operation{{&Operation{Protocol: "PsYLVpVvgbLhAhoqAkMFUo6gudkJ9weNXhUYCiLDzcUpFpkk8Wt", ChainID: "NetXZUqeBjDnWde", Hash: "opEatwYFvwuUM2aEa9cUU1ofMzsi46bYwiUhPLENXpLkjpps4Xq", Branch: "BLNWdEensT9MFq8pkDwjHfGVFsV1reYUhVcMAVzq3LCMS1WdKZ8", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 219132, Metadata: EndorsementOperationMetadata{BalanceUpdates: BalanceUpdates{&ContractBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: "contract", Change: -128000000}, Contract: "tz1SfH1vxAt2TTZV7mpsN79uGas5LHhV8epq"}, &FreezerBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: "freezer", Change: 128000000}, Category: "deposits", Delegate: "tz1SfH1vxAt2TTZV7mpsN79uGas5LHhV8epq", Level: 106}, &FreezerBalanceUpdate{GenericBalanceUpdate: GenericBalanceUpdate{Kind: "freezer", Change: 2000000}, Category: "rewards", Delegate: "tz1SfH1vxAt2TTZV7mpsN79uGas5LHhV8epq", Level: 106}}, Delegate: "tz1SfH1vxAt2TTZV7mpsN79uGas5LHhV8epq", Slots: []int{18, 16}}}}, Signature: "sigS3d9wfEFuChEqLetCxf4G8QYAjWL7ND3F8amMPVPDS2RwQqkeKU9hbrEXk7GG7U2aPcWkTA3uTdNzz4gkAb8jSy8hUc51"}}, {}, {}, {}}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				ch := make(chan *BlockInfo, 100)
				if err := s.MonitorHeads(ctx, "main", ch); err != nil {
					return nil, err
				}
				close(ch)

				var res []*BlockInfo
				for b := range ch {
					res = append(res, b)
				}
				return res, nil
			},
			respFixture:     "fixtures/monitor/heads.chunked",
			respContentType: "application/json",
			expectedPath:    "/monitor/heads/main",
			expectedValue: []*BlockInfo{
				{Hash: "BKq199p1Hm1phfJ4DhuRjB6yBSJnDNG8sgMSnja9pXR96T2Hyy1", Timestamp: timeMustUnmarshalText("2019-04-10T22:37:08Z"), OperationsHash: "LLobC6LA4T2STTa3D77YDuDsrw6xEY8DakpkvR9kd7DL9HpvchUtb", Level: 390397, Context: "CoUiJrzomxKms5eELzgpULo2iyf7dJAqW3gEBnFE7WHv3cy9pfVE", Predecessor: "BKihh4Bd3nAypX5bZtYy7xoxQDRbygkoyjB9w171exm2mbXHQWj", Proto: 3, ProtocolData: "000000000003bcf5f72d00320dffeb51c154077ce7dd2af6057f0370485a738345d3cb5c722db6df6ddb9b48c4e7a4282a3b994bca1cc52f6b95c889f23906e1d4e3e20203e171ff924004", ValidationPass: 4, Fitness: []HexBytes{{0x0}, {0x0, 0x0, 0x0, 0x0, 0x0, 0x5a, 0x12, 0x5f}}},
			},
		},
		{
			get: func(s *Service) (interface{}, error) {
				ch := make(chan []*Operation, 100)
				if err := s.MonitorMempoolOperations(ctx, "main", "", ch); err != nil {
					return nil, err
				}
				close(ch)

				var res []*Operation
				for b := range ch {
					res = append(res, b...)
				}
				return res, nil
			},
			respFixture:     "fixtures/monitor/mempool_operations.chunked",
			respContentType: "application/json",
			expectedPath:    "/chains/main/mempool/monitor_operations",
			expectedValue:   []*Operation{{Protocol: "Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd", Branch: "BKvSZMWpcDc9RkKg11sQ5oRDyHrMDiKX5RmTdU455XnPHuYZWRS", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 489922}}, Signature: "sigbdfHsA4XHTB3ToUMzRRAYmSJBCvJ52jdE7SrFp7BD3jUnd9sVBdzytHKTD6ygy343jRjJvc4E8kuZRiEqUdExH333RaqP"}, {Protocol: "Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd", Branch: "BKvSZMWpcDc9RkKg11sQ5oRDyHrMDiKX5RmTdU455XnPHuYZWRS", Contents: OperationElements{&EndorsementOperationElem{GenericOperationElem: GenericOperationElem{Kind: "endorsement"}, Level: 489922}}, Signature: "sigk5ep31BR1gSFSD37aiiAbT2azciyBdBaZD8Xp4Ef1NCT37L9ggucZySHhrNEnmqKZSRq5LKq5MJDVhj4tKmP1z8GqmY5j"}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetBallotList(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/ballot_list.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/ballot_list",
			expectedValue:   []*Ballot{{PKH: "tz3e75hU4EhDU3ukyJueh5v6UvEHzGwkg3yC", Ballot: "yay"}, {PKH: "tz1iEWcNL383qiDJ3Q3qt5W2T4aSKUbEU4An", Ballot: "nay"}, {PKH: "tz3bvNMQ95vfAYtG8193ymshqjSvmxiCUuR5", Ballot: "pass"}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetBallots(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/ballots.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/ballots",
			expectedValue:   &Ballots{Yay: 26776, Nay: 11, Pass: 19538},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetBallotListings(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/listings.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/listings",
			expectedValue:   []*BallotListing{{PKH: "tz1KfCukgwoU32Z4or88467mMM3in5smtv8k", Rolls: 5}, {PKH: "tz1KfEsrtDaA1sX7vdM4qmEPWuSytuqCDp5j", Rolls: 307}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetProposals(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/proposals.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/proposals",
			expectedValue:   []*Proposal{{ProposalHash: "Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd", SupporterCount: 11}},
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetCurrentProposals(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/current_proposal.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/current_proposal",
			expectedValue:   "Pt24m4xiPbLDhVgVfABUjirbmda3yohdN82Sp9FeuAXJ4eV9otd",
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetCurrentQuorum(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/current_quorum.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/current_quorum",
			expectedValue:   8000,
		},
		{
			get: func(s *Service) (interface{}, error) {
				return s.GetCurrentPeriodKind(ctx, "main", "head")
			},
			respFixture:     "fixtures/votes/current_period_kind.json",
			respContentType: "application/json",
			expectedPath:    "/chains/main/blocks/head/votes/current_period_kind",
			expectedValue:   PeriodKind("testing_vote"),
		},
	}

	for _, test := range tests {
		// Start a test HTTP server that responds as specified in the test case parameters.
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			require.Equal(t, test.expectedPath, r.URL.Path)

			if test.expectedQuery != "" {
				require.Equal(t, test.expectedQuery, r.URL.RawQuery)
			}

			m := test.expectedMethod
			if m == "" {
				m = http.MethodGet
			}
			require.Equal(t, m, r.Method)

			var buf []byte
			if test.respInline != "" {
				buf = []byte(test.respInline)
			} else {
				var err error
				buf, err = ioutil.ReadFile(test.respFixture)
				require.NoError(t, err, "error reading fixture %q", test.respFixture)
			}

			if test.respContentType != "" {
				w.Header().Set("Content-Type", test.respContentType)
			}

			if test.respStatus != 0 {
				w.WriteHeader(test.respStatus)
			}
			_, err := w.Write(buf)
			require.NoError(t, err, "error writing HTTP response")
		}))

		c, err := NewRPCClient(srv.URL)
		require.NoError(t, err, "error creating client")

		s := &Service{Client: c}

		value, err := test.get(s)

		if test.errType != nil {
			require.IsType(t, test.errType, err)
		}

		if test.errMsg == "" {
			require.NoError(t, err, "error getting value")
			require.Equal(t, test.expectedValue, value, "unexpected value")
		} else {
			require.EqualError(t, err, test.errMsg, "unexpected error string")
		}

		srv.Close()
	}
}
