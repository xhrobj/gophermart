package model

// AccrualStatus описывает статус расчета начисления во внешнем сервисе accrual.
type AccrualStatus string

const (
	// AccrualStatusRegistered - промежуточный статус:
	// заказ зарегистрирован, но начисление еще не рассчитано.
	AccrualStatusRegistered AccrualStatus = "REGISTERED"

	// AccrualStatusProcessing - промежуточный статус:
	// расчет начисления находится в процессе.
	AccrualStatusProcessing AccrualStatus = "PROCESSING"

	// AccrualStatusInvalid - финальный статус: заказ не принят к расчету.
	AccrualStatusInvalid AccrualStatus = "INVALID"

	// AccrualStatusProcessed - финальный статус: расчет начисления завершен.
	AccrualStatusProcessed AccrualStatus = "PROCESSED"
)

// AccrualResult описывает результат расчета начисления для заказа.
type AccrualResult struct {
	Order  string
	Status AccrualStatus

	// Accrual содержит сумму начисления в копейках.
	Accrual int64
}
