# ussd

USSD session engine for Mesob Wallet. Handles carrier gateway callbacks via FSM.

**Port**: 8010  
**Auth**: PARTNER·SIGNED  
**State store**: Redis (session TTL ≤180s)  
**Languages**: am, om, ti, en

## Local run
`MESOB_USSD_REDIS_URL=redis://localhost:6379 go run ./cmd/server`
