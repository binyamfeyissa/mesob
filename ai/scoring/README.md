# scoring

Credit scoring service for Mesob Wallet.

**Port**: 9001  
**Auth**: INTERNAL mTLS  
**Fallback chain**: ML model → rules scorecard → INSUFFICIENT_HISTORY (422)

## Features used (no protected attributes)
- `iqub_punctuality` — ratio of on-time Iqub contributions
- `savings_velocity_minor` — rolling 90-day deposit volume
- `months_active` — account tenure
- `prior_loan_repaid` — binary: had and repaid a loan
- `prior_loan_defaulted` — binary: had a defaulted loan

## Local run
```bash
cd ai/scoring
pip install -e .
uvicorn app.main:app --reload --port 9001
```

## Tests
```bash
python -m pytest tests/
```
