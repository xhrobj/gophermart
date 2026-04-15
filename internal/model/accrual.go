package model

type AccrualStatus string

const (
	AccrualStatusRegistered AccrualStatus = "REGISTERED"
	AccrualStatusProcessing AccrualStatus = "PROCESSING"
	AccrualStatusInvalid    AccrualStatus = "INVALID"
	AccrualStatusProcessed  AccrualStatus = "PROCESSED"
)

type AccrualResult struct {
	Order   string
	Status  AccrualStatus
	Accrual int64
}
