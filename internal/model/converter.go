package model

import (
	"github.com/google/uuid"
	pb "github.com/Raisondetr3/checklist-db-service/pkg/pb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TaskToProto(task *Task) *pb.Task {
	if task == nil {
		return nil
	}
	
	return &pb.Task{
		Id:          task.ID.String(),
		Title:       task.Title,
		Description: task.Description,
		Completed:   task.Completed,
		CreatedAt:   timestamppb.New(task.CreatedAt),
		UpdatedAt:   timestamppb.New(task.UpdatedAt),
	}
}

func TaskFromProto(protoTask *pb.Task) (*Task, error) {
	if protoTask == nil {
		return nil, nil
	}

	id, err := uuid.Parse(protoTask.Id)
	if err != nil {
		return nil, err
	}

	return &Task{
		ID:          id,
		Title:       protoTask.Title,
		Description: protoTask.Description,
		Completed:   protoTask.Completed,
		CreatedAt:   protoTask.CreatedAt.AsTime(),
		UpdatedAt:   protoTask.UpdatedAt.AsTime(),
	}, nil
}

func TasksToProto(tasks []*Task) []*pb.Task {
	if tasks == nil {
		return nil
	}

	protoTasks := make([]*pb.Task, len(tasks))
	for i, task := range tasks {
		protoTasks[i] = TaskToProto(task)
	}
	return protoTasks
}

func CreateTaskRequestFromProto(req *pb.CreateTaskRequest) (title, description string) {
	if req == nil {
		return "", ""
	}
	return req.Title, req.Description
}

func UpdateTaskRequestFromProto(req *pb.UpdateTaskRequest) (title, description *string, completed *bool) {
	if req == nil {
		return nil, nil, nil
	}

	if req.Title != nil {
		title = req.Title
	}
	if req.Description != nil {
		description = req.Description
	}
	if req.Completed != nil {
		completed = req.Completed
	}

	return title, description, completed
}

func GetTaskRequestFromProto(req *pb.GetTaskRequest) (uuid.UUID, error) {
	if req == nil {
		return uuid.Nil, nil
	}
	return uuid.Parse(req.Id)
}

func DeleteTaskRequestFromProto(req *pb.DeleteTaskRequest) (uuid.UUID, error) {
	if req == nil {
		return uuid.Nil, nil
	}
	return uuid.Parse(req.Id)
}


