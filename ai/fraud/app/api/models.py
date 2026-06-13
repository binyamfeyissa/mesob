from pydantic import BaseModel
from typing import List, Optional


class ScreenRequest(BaseModel):
    user_id: str
    type: str              # P2P, MERCHANT, BILL, CASH_IN, CASH_OUT
    amount_minor: int
    counterparty: str
    channel: str           # APP, USSD, AGENT


class ScreenResponse(BaseModel):
    decision: str          # ALLOW | REVIEW | BLOCK
    risk_score: float
    rules_hit: List[str]


class AlertRecord(BaseModel):
    alert_id: str
    severity: str
    rule: str
    transaction_id: str
    status: str            # OPEN


class DispositionRequest(BaseModel):
    disposition: str       # CLEAR | CONFIRM | ESCALATE_SAR
    note: str
