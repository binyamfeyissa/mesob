package domain

type CreditScore struct {
	ScoreID      string
	Score        int
	Tier         string
	CeilingMinor int64
	ModelVer     string
	Source       string
	Factors      []Factor
}

type Factor struct {
	Feature      string  `json:"feature"`
	Contribution float64 `json:"contribution"`
}
