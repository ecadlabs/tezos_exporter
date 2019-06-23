package tezos

// Ballot holds information about a Tezos ballot
type Ballot struct {
	PKH    string `json:"pkh"`
	Ballot string `json:"ballot"`
}

// BallotListing holds information about a Tezos delegate and his voting weight in rolls
type BallotListing struct {
	PKH   string `json:"pkh"`
	Rolls int64  `json:"rolls"`
}

// Ballots holds summary data about a voting period
type Ballots struct {
	Yay  int64 `json:"yay"`
	Nay  int64 `json:"nay"`
	Pass int64 `json:"pass"`
}

// Proposal holds summary data about a proposal and his number of supporter
type Proposal struct {
	ProposalHash   string
	SupporterCount int
}

// PeriodKind contains information about tezos voting period kind
type PeriodKind string

// IsProposal return true if period kind is proposal
func (p PeriodKind) IsProposal() bool {
	return p == "proposal"
}

// IsTestingVote return true if period kind is testing vote
func (p PeriodKind) IsTestingVote() bool {
	return p == "testing_vote"
}

// IsTesting return true if period kind is testing
func (p PeriodKind) IsTesting() bool {
	return p == "testing"
}

// IsPromotionVote true if period kind is promotion vote
func (p PeriodKind) IsPromotionVote() bool {
	return p == "promotion_vote"
}
