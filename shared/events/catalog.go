package events

const (
	TransactionPosted        = "TransactionPosted"
	TransactionReversed      = "TransactionReversed"
	UserActivated            = "UserActivated"
	KycTierChanged           = "KycTierChanged"
	IqubContributionRecorded = "IqubContributionRecorded"
	IqubContributionMissed   = "IqubContributionMissed"
	IqubCycleClosed          = "IqubCycleClosed"
	LoanDecisioned           = "LoanDecisioned"
	LoanDisbursed            = "LoanDisbursed"
	LoanRepaid               = "LoanRepaid"
	FraudAlertRaised         = "FraudAlertRaised"
	AgentFloatLow            = "AgentFloatLow"
	SettlementConfirmed      = "SettlementConfirmed"
	PaymentCompleted         = "PaymentCompleted"
	PremiumPaid              = "PremiumPaid"
	ClaimFiled               = "ClaimFiled"
	ClaimSettled             = "ClaimSettled"
	AgentApproved            = "AgentApproved"
	DisputeResolved          = "DisputeResolved"
	PaymentReversed          = "PaymentReversed"
	CashInRecorded           = "CashInRecorded"
	CashOutRecorded          = "CashOutRecorded"
)

// Topic returns the Kafka topic for a given event type.
// Naming: domain.event-name (e.g. ledger.transaction-posted)
func Topic(eventType string) string {
	switch eventType {
	case TransactionPosted:
		return "ledger.transaction-posted"
	case TransactionReversed:
		return "ledger.transaction-reversed"
	case UserActivated:
		return "identity.user-activated"
	case KycTierChanged:
		return "identity.kyc-tier-changed"
	case IqubContributionRecorded:
		return "iqub.contribution-recorded"
	case IqubContributionMissed:
		return "iqub.contribution-missed"
	case IqubCycleClosed:
		return "iqub.cycle-closed"
	case LoanDecisioned:
		return "loans.decisioned"
	case LoanDisbursed:
		return "loans.disbursed"
	case LoanRepaid:
		return "loans.repaid"
	case FraudAlertRaised:
		return "fraud.alert-raised"
	case AgentFloatLow:
		return "agent.float-low"
	case SettlementConfirmed:
		return "branch.settlement-confirmed"
	case PaymentCompleted:
		return "payments.completed"
	case PaymentReversed:
		return "payments.reversed"
	case PremiumPaid:
		return "iddir.premium-paid"
	case ClaimFiled:
		return "iddir.claim-filed"
	case ClaimSettled:
		return "iddir.claim-settled"
	case AgentApproved:
		return "branch.agent-approved"
	case DisputeResolved:
		return "branch.dispute-resolved"
	case CashInRecorded:
		return "agent.cash-in-recorded"
	case CashOutRecorded:
		return "agent.cash-out-recorded"
	default:
		return "unknown." + eventType
	}
}
