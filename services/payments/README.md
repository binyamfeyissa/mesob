# Payments Service
P2P transfers, merchant payments, bill pay. Fail-closed fraud screening.
Owns: merchants, billers, payment_refs (mesob_payments DB).
Emits: PaymentCompleted, PaymentReversed.
Critical: If FraudClient unavailable → decline (never pass through).
