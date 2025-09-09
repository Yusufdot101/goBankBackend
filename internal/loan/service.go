package loan

import (
	"math"
	"time"

	"github.com/Yusufdot101/goBankBackend/internal/user"
	"github.com/Yusufdot101/goBankBackend/internal/validator"
)

type Repo interface {
	Insert(*Loan) error
	GetByID(loanID, userID int64) (*Loan, error)
	InsertDeletion(loan *LoanDeletion) error
	MakePaymentTx(loanID, userID int64, payment, totalOwed float64) (*Loan, error)
	DeleteLoan(loanID, debtorID int64) error
}

type UserService interface {
	GetUser(userID int64) (*user.User, error)
	UpdateUser(
		userID int64, userName, userEmail string, userPasswordHash []byte,
		userAccountBalance float64, userActivated bool) (*user.User, error)
}

type Service struct {
	Repo        Repo
	UserService UserService
}

func (s *Service) GetLoan(
	u *user.User, amount, dailyInterestRate float64,
) error {
	loan := Loan{
		UserID:            u.ID,
		Amount:            amount,
		Action:            "took",
		DailyInterestRate: dailyInterestRate,
		RemainingAmount:   amount,
		LastUpdatedAt:     time.Now(),
	}

	err := s.Repo.Insert(&loan)
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) MakePayment(
	v *validator.Validator, loanID, userID int64, payment float64,
) (*Loan, error) {
	if payment <= 0 {
		v.AddError("amount", "must be more than 0")
		return nil, validator.ErrFailedValidation
	}
	loan, err := s.Repo.GetByID(loanID, userID)
	if err != nil {
		return nil, err
	}

	if loan.RemainingAmount == 0 {
		v.AddError("loan", "is already paid off")
		return nil, validator.ErrFailedValidation
	}

	// get the user
	u, err := s.UserService.GetUser(userID)
	if err != nil {
		return nil, err
	}

	// check if he has enough funds
	if u.AccountBalance < payment {
		v.AddError("account_balance", "insufficient funds")
		return nil, validator.ErrFailedValidation
	}

	// get the time since last payment was made, we use LastUpdatedAt instead of created_at to
	// avoid over-charging in partial payments.
	elapsedTimeDays := time.Since(loan.LastUpdatedAt).Hours() / 24
	interest := elapsedTimeDays * (loan.RemainingAmount * (loan.DailyInterestRate / 100))
	totalOwed := loan.RemainingAmount + interest

	loan, err = s.Repo.MakePaymentTx(loan.ID, userID, payment, totalOwed)
	if err != nil {
		return nil, err
	}

	loanPayment := Loan{
		UserID:            loan.UserID,
		Amount:            math.Min(payment, totalOwed),
		Action:            "paid",
		RemainingAmount:   loan.RemainingAmount,
		DailyInterestRate: loan.DailyInterestRate,
		LastUpdatedAt:     loan.LastUpdatedAt,
	}

	err = s.Repo.Insert(&loanPayment)
	if err != nil {
		return nil, err
	}

	// deduct the payment from the users account
	u.AccountBalance -= loanPayment.Amount
	_, err = s.UserService.UpdateUser(
		u.ID, u.Name, u.Email, u.Password.Hash, u.AccountBalance, u.Activated,
	)
	if err != nil {
		return nil, err
	}

	return &loanPayment, nil
}

func (s *Service) DeleteLoan(
	v *validator.Validator, loanID, debtorID, deletedByID int64, reason string,
) (*LoanDeletion, error) {
	loanDeletion := &LoanDeletion{Reason: reason}
	if ValidateLoanDeletion(v, loanDeletion); !v.IsValid() {
		return nil, validator.ErrFailedValidation
	}

	loan, err := s.Repo.GetByID(loanID, debtorID)
	if err != nil {
		return nil, err
	}

	loanDeletion.LoanCreatedAt = loan.CreatedAt
	loanDeletion.LoanLastUpdatedAt = loan.LastUpdatedAt
	loanDeletion.LoanID = loan.ID
	loanDeletion.DebtorID = loan.UserID
	loanDeletion.DeletedByID = deletedByID
	loanDeletion.Amount = loan.Amount
	loanDeletion.RemainingAmount = loan.RemainingAmount
	loanDeletion.DailyInterestRate = loan.DailyInterestRate
	loanDeletion.Reason = reason

	// first record the deletion before deleting the loan because we could get an error before
	// recording it after removing the loan from the system. if we get an error removing the loan
	// but the deletion is recorded we could handle it as the loan with the id exists and deal with
	// it. we retry 5 times to record the entry
	err = nil // clean the err var before
	for range 5 {
		err = s.Repo.InsertDeletion(loanDeletion)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond) // wait for 100ms before retrying.
	}
	if err != nil {
		return nil, err
	}

	// delete the actual loan. we retry 5 times
	err = nil
	for range 5 {
		err = s.Repo.DeleteLoan(loanID, debtorID)
		if err == nil {
			break
		}
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, err
	}

	return loanDeletion, nil
}
