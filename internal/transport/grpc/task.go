package grpc

import (
	"context"

	pb "github.com/Raisondetr3/checklist-db-service/pkg/pb"
)

func (s *GRPCServer) CreateTask(ctx context.Context, req *pb.CreateTaskRequest) (*pb.TaskResponse, error) {
	return s.taskService.CreateTask(ctx, req)
}

func (s *GRPCServer) GetTask(ctx context.Context, req *pb.GetTaskRequest) (*pb.TaskResponse, error) {
	return s.taskService.GetTask(ctx, req)
}

func (s *GRPCServer) UpdateTask(ctx context.Context, req *pb.UpdateTaskRequest) (*pb.TaskResponse, error) {
	return s.taskService.UpdateTask(ctx, req)
}

func (s *GRPCServer) DeleteTask(ctx context.Context, req *pb.DeleteTaskRequest) (*pb.DeleteTaskResponse, error) {
	return s.taskService.DeleteTask(ctx, req)
}

func (s *GRPCServer) ListTasks(ctx context.Context, req *pb.ListTasksRequest) (*pb.ListTasksResponse, error) {
	return s.taskService.ListTasks(ctx, req)
}
