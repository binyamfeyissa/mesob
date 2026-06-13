export interface Envelope {
  event_id: string;
  type: string;
  version: number;
  occurred_at: string;
  aggregate_id: string;
  payload: unknown;
}

export const EventTypes = {
  TransactionPosted: "TransactionPosted",
  TransactionReversed: "TransactionReversed",
  UserActivated: "UserActivated",
  KycTierChanged: "KycTierChanged",
  IqubContributionRecorded: "IqubContributionRecorded",
  IqubContributionMissed: "IqubContributionMissed",
  IqubCycleClosed: "IqubCycleClosed",
  LoanDecisioned: "LoanDecisioned",
  LoanDisbursed: "LoanDisbursed",
  LoanRepaid: "LoanRepaid",
  FraudAlertRaised: "FraudAlertRaised",
  AgentFloatLow: "AgentFloatLow",
  SettlementConfirmed: "SettlementConfirmed",
  PaymentCompleted: "PaymentCompleted",
  PremiumPaid: "PremiumPaid",
  ClaimFiled: "ClaimFiled",
  ClaimSettled: "ClaimSettled",
  AgentApproved: "AgentApproved",
  DisputeResolved: "DisputeResolved",
  PaymentReversed: "PaymentReversed",
  CashInRecorded: "CashInRecorded",
  CashOutRecorded: "CashOutRecorded",
} as const;

export type EventType = typeof EventTypes[keyof typeof EventTypes];

export interface MetaBlock {
  request_id: string;
  trace_id: string;
  next_cursor?: string;
}

export interface SuccessResponse<T> {
  data: T;
  meta: MetaBlock;
}

export interface ErrorBlock {
  code: string;
  message: string;
  details?: { field: string; message: string }[];
}

export interface ErrorResponse {
  error: ErrorBlock;
  meta: MetaBlock;
}
