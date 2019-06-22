package webrtc

import "fmt"

// ICECandidatePair represents an ICE Candidate pair
type ICECandidatePair struct {
	ObjectID string
	Local    *ICECandidate
	Remote   *ICECandidate
}

func newICECandidatePairObjectID(localID, remoteID string) string {
	return fmt.Sprintf("%s-%s", localID, remoteID)
}

func (p *ICECandidatePair) String() string {
	return fmt.Sprintf("(local) %s <-> (remote) %s", p.Local, p.Remote)
}

// NewICECandidatePair returns an initialized *ICECandidatePair
// for the given pair of ICECandidate instances
func NewICECandidatePair(local, remote *ICECandidate) *ICECandidatePair {
	objectID := newICECandidatePairObjectID(local.ObjectID, remote.ObjectID)
	return &ICECandidatePair{
		ObjectID: objectID,
		Local:    local,
		Remote:   remote,
	}
}
