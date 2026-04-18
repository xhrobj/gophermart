package service

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/xhrobj/gophermart/internal/model"
	"github.com/xhrobj/gophermart/internal/repository"
)

var (
	ErrInvalidOrderInput  = errors.New("invalid order input")
	ErrInvalidOrderNumber = errors.New("invalid order number")
)

// OrderService описывает операции с заказами пользователя.
type OrderService interface {
	// UploadOrder загружает номер заказа пользователя в систему.
	//
	// Для успешно разобранной бизнес-операции возвращает один из статусов: ACCEPTED, DUPLICATE или CONFLICT.
	// Некорректный номер заказа возвращается как ошибка.
	UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error)

	// ListOrders возвращает список заказов пользователя.
	ListOrders(ctx context.Context, userID int64) ([]model.Order, error)
}

type orderService struct {
	orderRepo repository.OrderRepository
}

// NewOrderService создаёт сервис работы с заказами пользователя.
func NewOrderService(orderRepo repository.OrderRepository) OrderService {
	return &orderService{
		orderRepo: orderRepo,
	}
}

func (s *orderService) UploadOrder(ctx context.Context, userID int64, orderNumber string) (model.UploadOrderResult, error) {
	orderNumber = strings.TrimSpace(orderNumber)

	if orderNumber == "" {
		return model.UploadOrderResult{}, ErrInvalidOrderInput
	}

	if !(isDigitsOnly(orderNumber) && isValidLuhn(orderNumber)) {
		return model.UploadOrderResult{}, ErrInvalidOrderNumber
	}

	existingOrder, err := s.orderRepo.FindByNumber(ctx, orderNumber)
	if err == nil {
		return buildUploadOrderResult(userID, existingOrder), nil
	}

	if !errors.Is(err, repository.ErrOrderNotFound) {
		return model.UploadOrderResult{}, fmt.Errorf("find order by number: %w", err)
	}

	createdOrder, err := s.orderRepo.Create(ctx, userID, orderNumber)
	if err == nil {
		// -> заказ создан
		return model.UploadOrderResult{
			Status: model.UploadOrderAccepted,
			Order:  createdOrder,
		}, nil
	}

	if errors.Is(err, repository.ErrOrderAlreadyExists) {
		existingOrder, findErr := s.orderRepo.FindByNumber(ctx, orderNumber)
		if findErr != nil {
			return model.UploadOrderResult{}, fmt.Errorf("find order by number after create conflict: %w", findErr)
		}

		return buildUploadOrderResult(userID, existingOrder), nil
	}

	return model.UploadOrderResult{}, fmt.Errorf("create order: %w", err)
}

func (s *orderService) ListOrders(ctx context.Context, userID int64) ([]model.Order, error) {
	orders, err := s.orderRepo.ListByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list orders by user id: %w", err)
	}

	return orders, nil
}

func buildUploadOrderResult(userID int64, order model.Order) model.UploadOrderResult {
	if order.UserID == userID {
		return model.UploadOrderResult{
			Status: model.UploadOrderDuplicate,
			Order:  order,
		}
	}

	return model.UploadOrderResult{
		Status: model.UploadOrderConflict,
		Order:  order,
	}
}

func isDigitsOnly(s string) bool {
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return false
		}
	}

	return true
}

func isValidLuhn(number string) bool {
	sum := 0
	double := false

	for i := len(number) - 1; i >= 0; i-- {
		digit := int(number[i] - '0')

		if double {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}

		sum += digit
		double = !double
	}

	return sum%10 == 0
}
