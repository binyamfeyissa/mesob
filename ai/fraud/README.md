# fraud

Real-time fraud/AML screening service for Mesob Wallet.

**Port**: 9002  
**Auth**: INTERNAL mTLS  
**Screening**: Deterministic rules + Isolation Forest anomaly detection  
**Decisions**: ALLOW | REVIEW | BLOCK  
**Behaviour**: Fails closed — callers must decline if this service is unavailable.

## Rules
- High velocity (>10 txns/hr, >30 txns/24h)
- High volume (>50k ETB/hr)
- Structuring near 10,000 ETB threshold
- Fan-out counterparties (>5 unique/hr)
- Near-ceiling float (agent cash-in >=90% of ceiling)

## Local run
```bash
cd ai/fraud
pip install -e .
uvicorn app.main:app --reload --port 9002
```

## Tests
```bash
python -m pytest tests/
```
