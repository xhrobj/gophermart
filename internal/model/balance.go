package model

type Balance struct {
	// Сумма всех начислений [минус] сумма всех списаний (в копейках).
	Current int64

	// Сумма всех списаний (в копейках).
	Withdrawn int64
}
