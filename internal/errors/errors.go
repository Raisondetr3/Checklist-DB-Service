package errors

import (
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type ServiceError struct {
	Code    codes.Code `json:"code"`
	Message string     `json:"message"`
	Time    time.Time  `json:"time"`
}

func NewServiceError(code codes.Code, message string) *ServiceError {
	return &ServiceError{
		Code:    code,
		Message: message,
		Time:    time.Now(),
	}
}

func (e *ServiceError) Error() string {
	return fmt.Sprintf("code: %s, message: %s, time: %s", 
		e.Code.String(), e.Message, e.Time.Format(time.RFC3339))
}

func (e *ServiceError) ToGRPCStatus() error {
	return status.Error(e.Code, e.Message)
}

var (
	ErrTitleNotSpecified = NewServiceError(codes.InvalidArgument, "title is required")
	ErrInvalidTaskId     = NewServiceError(codes.InvalidArgument, "invalid task id")
	ErrTaskNotFound      = NewServiceError(codes.NotFound, "task not found")
	ErrTaskAlreadyExists = NewServiceError(codes.AlreadyExists, "task already exists")
	ErrInternalError     = NewServiceError(codes.Internal, "internal server error")
)

func WrapRepositoryError(err error) *ServiceError {
	if err == nil {
		return nil
	}
	
	switch {
	case IsNotFoundError(err):
		return ErrTaskNotFound
	case IsConstraintViolationError(err):
		return ErrTaskAlreadyExists
	default:
		return NewServiceError(codes.Internal, fmt.Sprintf("repository error: %v", err))
	}
}

func IsNotFoundError(err error) bool {
	return err.Error() == "no rows in result set"
}

func IsConstraintViolationError(err error) bool {
	return false 
}