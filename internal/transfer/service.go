package transfer

import (
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Service struct {
	Repo *Repository
}

func (s *Service) TransferMoney(
	v *validator.Validator, userService *user.Service, fromUser *user.User, toUserEmail string,
	amount float64,
) (*Transfer, *user.User, error) {
	toUser, err := userService.Repo.GetByEmail(toUserEmail)
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
		return nil, nil, nil
	}

	fromUser, err = userService.TransferMoney(fromUser, toUser, transfer.Amount)
	if err != nil {
		return nil, nil, err
	}

	err = s.Repo.Insert(&transfer)
	if err != nil {
		return nil, nil, err
	}

	return &transfer, fromUser, nil
}
