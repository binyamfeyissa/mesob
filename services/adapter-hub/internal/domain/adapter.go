package domain

type AdapterMode string

const (
	ModeDemo AdapterMode = "DEMO"
	ModeLive AdapterMode = "LIVE"
)

type BreakerState string

const (
	BreakerClosed BreakerState = "CLOSED"
	BreakerOpen   BreakerState = "OPEN"
)

type AdapterStatus struct {
	Adapter string
	Mode    AdapterMode
	Breaker BreakerState
	Healthy bool
}

type NIDVerifyResult struct {
	Verified   bool
	MatchScore float64
	Mode       AdapterMode
}

type MFIOriginateResult struct {
	FacilityID string
	Status     string // DISBURSED
	Mode       AdapterMode
}
