package service

import (
	"context"
	"time"

	"github.com/Raisondetr3/checklist-db-service/internal/errors"
	"github.com/Raisondetr3/checklist-db-service/internal/model"
	"github.com/Raisondetr3/checklist-db-service/internal/repository"
	"github.com/Raisondetr3/checklist-db-service/pkg/logger"
	pb "github.com/Raisondetr3/checklist-db-service/pkg/pb"
	"github.com/google/uuid"
)

type TaskService interface {
	CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error)
	GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error)
	UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.TaskResponse, error)
	DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error)
	ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error)
}

type taskService struct {
	taskRepo repository.TaskRepository
}

func NewTaskService(taskRepo repository.TaskRepository) TaskService {
	return &taskService{
		taskRepo: taskRepo,
	}
}

func (s *taskService) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	start := time.Now()
	operation := "CreateTask"

	if req.Title == "" {
		logger.LogError(ctx, errors.ErrTitleNotSpecified, operation)
		return nil, errors.ErrTitleNotSpecified.ToGRPCStatus()
	}

	title, description := model.CreateTaskRequestFromProto(req)
	task := model.NewTask(title, description)

	savedTask, err := s.taskRepo.Create(ctx, task)
	duration := time.Since(start)

	if err != nil {
		serviceErr := errors.WrapRepositoryError(err)
		logger.LogTaskOperation(ctx, operation, task.ID.String(), duration, serviceErr)
		return nil, serviceErr.ToGRPCStatus()
	}

	logger.LogTaskOperation(ctx, operation, savedTask.ID.String(), duration, nil)

	return &pb.TaskResponse{
		Task: model.TaskToProto(savedTask),
	}, nil
}

func (s *taskService) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
	start := time.Now()
	operation := "GetTask"

	id, err := model.GetTaskRequestFromProto(req)
	if err != nil {
		logger.LogError(ctx, errors.ErrInvalidTaskId, operation)
		return nil, errors.ErrInvalidTaskId.ToGRPCStatus()
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	duration := time.Since(start)

	if err != nil {
		serviceErr := errors.WrapRepositoryError(err)
		logger.LogTaskOperation(ctx, operation, id.String(), duration, serviceErr)
		return nil, serviceErr.ToGRPCStatus()
	}

	logger.LogTaskOperation(ctx, operation, task.ID.String(), duration, nil)

	return &pb.TaskResponse{
		Task: model.TaskToProto(task),
	}, nil
}

func (s *taskService) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.TaskResponse, error) {
	start := time.Now()
	operation := "UpdateTask"

	id, err := uuid.Parse(req.Id)
	if err != nil {
		logger.LogError(ctx, errors.ErrInvalidTaskId, operation)
		return nil, errors.ErrInvalidTaskId.ToGRPCStatus()
	}

	task, err := s.taskRepo.GetByID(ctx, id)
	if err != nil {
		duration := time.Since(start)
		serviceErr := errors.WrapRepositoryError(err)
		logger.LogTaskOperation(ctx, operation, id.String(), duration, serviceErr)
		return nil, serviceErr.ToGRPCStatus()
	}

	title, description, completed := model.UpdateTaskRequestFromProto(req)
	task.Update(title, description, completed)

	updatedTask, err := s.taskRepo.Update(ctx, task)
	duration := time.Since(start)

	if err != nil {
		serviceErr := errors.WrapRepositoryError(err)
		logger.LogTaskOperation(ctx, operation, task.ID.String(), duration, serviceErr)
		return nil, serviceErr.ToGRPCStatus()
	}

	logger.LogTaskOperation(ctx, operation, updatedTask.ID.String(), duration, nil)

	return &pb.TaskResponse{
		Task: model.TaskToProto(updatedTask),
	}, nil
}

func (s *taskService) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	start := time.Now()
	operation := "DeleteTask"

	id, err := model.DeleteTaskRequestFromProto(req)
	if err != nil {
		logger.LogError(ctx, errors.ErrInvalidTaskId, operation)
		return nil, errors.ErrInvalidTaskId.ToGRPCStatus()
	}

	err = s.taskRepo.DeleteByID(ctx, id)
	duration := time.Since(start)

	if err != nil {
		serviceErr := errors.WrapRepositoryError(err)
		logger.LogTaskOperation(ctx, operation, id.String(), duration, serviceErr)
		return nil, serviceErr.ToGRPCStatus()
	}

	logger.LogTaskOperation(ctx, operation, id.String(), duration, nil)

	return &pb.DeleteTaskResponse{
		Success: true,
	}, nil
}

func (s *taskService) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	start := time.Now()
	operation := "ListTasks"

	tasks, err := s.taskRepo.List(ctx)
	duration := time.Since(start)

	if err != nil {
		serviceErr := errors.WrapRepositoryError(err)
		logger.LogTaskOperation(ctx, operation, "", duration, serviceErr)
		return nil, serviceErr.ToGRPCStatus()
	}

	logger.LogTaskOperation(ctx, operation, "", duration, nil)

	return &pb.ListTasksResponse{
		Tasks: model.TasksToProto(tasks),
	}, nil
}