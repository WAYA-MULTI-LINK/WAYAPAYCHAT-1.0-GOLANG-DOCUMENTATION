package wayapay

import "strings"

// CollectionStatusCode is a recognised value of CollectionStatus.Status. Parse
// a raw API string with ParseCollectionStatus; unrecognised values map to
// CollectStatusUnknown.
type CollectionStatusCode string

// Collection status codes. The typed const names are prefixed to avoid clashing
// with the raw payout status string consts (StatusSuccess, StatusFailed, ...).
const (
	CollectStatusUnknown CollectionStatusCode = "UNKNOWN"

	// In flight (non-terminal): keep polling; don't refund or retry.
	CollectStatusInitiated  CollectionStatusCode = "INITIATED"
	CollectStatusPending    CollectionStatusCode = "PENDING"
	CollectStatusProcessing CollectionStatusCode = "PROCESSING"
	CollectStatusApproved   CollectionStatusCode = "APPROVED"
	CollectStatusPartial    CollectionStatusCode = "PARTIAL"

	// Terminal: funds confirmed.
	CollectStatusSuccessful CollectionStatusCode = "SUCCESSFUL"
	CollectStatusRefunded   CollectionStatusCode = "REFUNDED"

	// Terminal: customer not debited — no fulfilment.
	CollectStatusFailed        CollectionStatusCode = "FAILED"
	CollectStatusDeclined      CollectionStatusCode = "DECLINED"
	CollectStatusRejected      CollectionStatusCode = "REJECTED"
	CollectStatusAbandoned     CollectionStatusCode = "ABANDONED"
	CollectStatusExpired       CollectionStatusCode = "EXPIRED"
	CollectStatusCancelled     CollectionStatusCode = "CANCELLED"
	CollectStatusCustomerError CollectionStatusCode = "CUSTOMER_ERROR"
	CollectStatusFraudError    CollectionStatusCode = "FRAUD_ERROR"

	// Terminal: outcome unknown — reconcile, don't refund unilaterally.
	CollectStatusTimeout     CollectionStatusCode = "TIMEOUT"
	CollectStatusError       CollectionStatusCode = "ERROR"
	CollectStatusSystemError CollectionStatusCode = "SYSTEM_ERROR"
	CollectStatusBankError   CollectionStatusCode = "BANK_ERROR"
)

// CollectionOutcome is the action a merchant should take for a collection
// status.
type CollectionOutcome string

const (
	// CollectionInFlight means keep polling; don't refund or retry.
	CollectionInFlight CollectionOutcome = "IN_FLIGHT"
	// CollectionSucceeded means funds confirmed — fulfil the order.
	CollectionSucceeded CollectionOutcome = "SUCCEEDED"
	// CollectionRefunded means a previously-successful transaction was refunded.
	CollectionRefunded CollectionOutcome = "REFUNDED"
	// CollectionNotDebited means the customer was not debited — do not fulfil.
	CollectionNotDebited CollectionOutcome = "NOT_DEBITED"
	// CollectionIndeterminate means the outcome is unknown — reconcile; don't
	// refund unilaterally.
	CollectionIndeterminate CollectionOutcome = "INDETERMINATE"
)

// ParseCollectionStatus parses a raw collection status string. The input is
// trimmed and upper-cased; unrecognised values return CollectStatusUnknown.
func ParseCollectionStatus(s string) CollectionStatusCode {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "INITIATED":
		return CollectStatusInitiated
	case "PENDING":
		return CollectStatusPending
	case "PROCESSING":
		return CollectStatusProcessing
	case "APPROVED":
		return CollectStatusApproved
	case "PARTIAL":
		return CollectStatusPartial
	case "SUCCESSFUL":
		return CollectStatusSuccessful
	case "REFUNDED":
		return CollectStatusRefunded
	case "FAILED":
		return CollectStatusFailed
	case "DECLINED":
		return CollectStatusDeclined
	case "REJECTED":
		return CollectStatusRejected
	case "ABANDONED":
		return CollectStatusAbandoned
	case "EXPIRED":
		return CollectStatusExpired
	case "CANCELLED":
		return CollectStatusCancelled
	case "CUSTOMER_ERROR":
		return CollectStatusCustomerError
	case "FRAUD_ERROR":
		return CollectStatusFraudError
	case "TIMEOUT":
		return CollectStatusTimeout
	case "ERROR":
		return CollectStatusError
	case "SYSTEM_ERROR":
		return CollectStatusSystemError
	case "BANK_ERROR":
		return CollectStatusBankError
	default:
		return CollectStatusUnknown
	}
}

// Outcome maps a collection status to the action a merchant should take.
// Unknown maps to CollectionIndeterminate — reconcile rather than guess.
func (c CollectionStatusCode) Outcome() CollectionOutcome {
	switch c {
	case CollectStatusInitiated, CollectStatusPending, CollectStatusProcessing,
		CollectStatusApproved, CollectStatusPartial:
		return CollectionInFlight
	case CollectStatusSuccessful:
		return CollectionSucceeded
	case CollectStatusRefunded:
		return CollectionRefunded
	case CollectStatusFailed, CollectStatusDeclined, CollectStatusRejected,
		CollectStatusAbandoned, CollectStatusExpired, CollectStatusCancelled,
		CollectStatusCustomerError, CollectStatusFraudError:
		return CollectionNotDebited
	default:
		// Timeout / Error / SystemError / BankError / Unknown
		return CollectionIndeterminate
	}
}

// IsTerminal reports whether the status will no longer change. Non-terminal
// statuses should be polled.
func (c CollectionStatusCode) IsTerminal() bool {
	switch c {
	case CollectStatusInitiated, CollectStatusPending, CollectStatusProcessing,
		CollectStatusApproved, CollectStatusPartial, CollectStatusUnknown:
		return false
	default:
		return true
	}
}

// PayoutStatusCode is a recognised value of PayoutStatus.Status. Parse a raw
// API string with ParsePayoutStatus; unrecognised values map to
// PayoutStatusUnknown.
type PayoutStatusCode string

// Payout status codes. Names are prefixed to avoid clashing with the raw payout
// status string consts (StatusSuccess, StatusFailed, StatusProcessing).
const (
	PayoutStatusUnknown  PayoutStatusCode = "UNKNOWN"
	PayoutStatusPending  PayoutStatusCode = "PENDING"
	PayoutStatusSuccess  PayoutStatusCode = "SUCCESS"
	PayoutStatusReversed PayoutStatusCode = "REVERSED"
)

// PayoutOutcome is the action a merchant should take for a payout status.
type PayoutOutcome string

const (
	// PayoutReconciling means the terminal outcome is not yet recorded — keep
	// reconciling.
	PayoutReconciling PayoutOutcome = "RECONCILING"
	// PayoutSucceeded means the payout completed — funds delivered.
	PayoutSucceeded PayoutOutcome = "SUCCEEDED"
	// PayoutReversed means the payout failed/reversed — the merchant wallet was
	// re-credited.
	PayoutReversed PayoutOutcome = "REVERSED"
)

// ParsePayoutStatus parses a raw payout status string. The input is trimmed and
// upper-cased; unrecognised values return PayoutStatusUnknown.
func ParsePayoutStatus(s string) PayoutStatusCode {
	switch strings.ToUpper(strings.TrimSpace(s)) {
	case "PENDING":
		return PayoutStatusPending
	case "SUCCESS":
		return PayoutStatusSuccess
	case "REVERSED":
		return PayoutStatusReversed
	default:
		return PayoutStatusUnknown
	}
}

// Outcome maps a payout status to the action a merchant should take. Unknown
// maps to PayoutReconciling — reconcile rather than guess.
func (p PayoutStatusCode) Outcome() PayoutOutcome {
	switch p {
	case PayoutStatusSuccess:
		return PayoutSucceeded
	case PayoutStatusReversed:
		return PayoutReversed
	default:
		// Pending / Unknown
		return PayoutReconciling
	}
}

// IsTerminal reports whether the status will no longer change. Non-terminal
// statuses should be reconciled.
func (p PayoutStatusCode) IsTerminal() bool {
	switch p {
	case PayoutStatusPending, PayoutStatusUnknown:
		return false
	default:
		return true
	}
}
