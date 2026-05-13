package services

import (
	"fmt"

	"github.com/lh-khanhduy/banco_de_rata/pb"
	"github.com/lh-khanhduy/banco_de_rata/utils"
	"github.com/lh-khanhduy/banco_de_rata/validator"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
)

// USERS
func validateCreateUserRequest(req *pb.CreateUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	if err := validator.ValidateFullName(req.GetFullName()); err != nil {
		violations = append(violations, fieldViolation("full_name", err))
	}

	if err := validator.ValidateEmail(req.GetEmail()); err != nil {
		violations = append(violations, fieldViolation("email", err))
	}

	return violations
}

func validateLoginUserRequest(req *pb.LoginUserRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if err := validator.ValidateUsername(req.GetUsername()); err != nil {
		violations = append(violations, fieldViolation("username", err))
	}

	if err := validator.ValidatePassword(req.GetPassword()); err != nil {
		violations = append(violations, fieldViolation("password", err))
	}

	return violations
}

// ACCOUNTS
func validateCreateAccountRequest(req *pb.CreateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if !utils.IsSupportedCurrency(req.GetCurrency()) {
		err := fmt.Errorf("must be one of: USD, EUR, CAD")
		violations = append(violations, fieldViolation("currency", err))
	}

	return violations
}

func validateUpdateAccountRequest(req *pb.UpdateAccountRequest) (violations []*errdetails.BadRequest_FieldViolation) {
	if req.GetBalance() < 0 {
		err := fmt.Errorf("account balance cannot be less than 0")
		violations = append(violations, fieldViolation("balance", err))
	}

	return violations
}
