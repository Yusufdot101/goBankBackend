package transfer

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type TransferRepo interface {
	Insert(transfer *Transfer) error
}

type UserService interface {
	TransferMoney(fromUser, toUser *user.User, amount float64) (*user.User, error)
	GetUserByEmail(email string) (*user.User, error)
}

type Service struct {
	Repo        TransferRepo
	UserService UserService
}

func (s *Service) TransferMoney(
	v *validator.Validator, fromUser *user.User, toUserEmail string, amount float64,
) (*Transfer, *user.User, error) {
	toUser, err := s.UserService.GetUserByEmail(toUserEmail)
	if err != nil {
		return nil, nil, err
	}

	transfer := Transfer{
		CreatedAd:  time.Now(),
		FromUserID: fromUser.ID,
		ToUserID:   toUser.ID,
		Amount:     amount,
	}

	if ValidateTransfer(v, &transfer, fromUser); !v.IsValid() {
		return nil, nil, validator.ErrFailedValidation
	}

	fromUser, err = s.UserService.TransferMoney(fromUser, toUser, transfer.Amount)
	if err != nil {
		return nil, nil, err
	}

	err = s.Repo.Insert(&transfer)
	if err != nil {
		return nil, nil, err
	}

	return &transfer, fromUser, nil
}
