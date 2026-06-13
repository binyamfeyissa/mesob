package domain

type KYCLimits struct {
	Tier         int8
	PerTxnMinor  int64
	DailyMinor   int64
	BalanceMinor int64
}
