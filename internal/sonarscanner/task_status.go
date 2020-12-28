package sonarscanner

import "fmt"

type TaskStatus int

const (
	TaskStatusUndefined  TaskStatus = iota
	TaskStatusPending    TaskStatus = iota
	TaskStatusInProgress TaskStatus = iota
	TaskStatusSuccess    TaskStatus = iota
	TaskStatusFailed     TaskStatus = iota
	TaskStatusCancelled  TaskStatus = iota
)

const (
	taskStatusUndefinedStr  = "UNDEFINED"
	taskStatusPendingStr    = "PENDING"
	taskStatusInProgressStr = "IN_PROGRESS"
	taskStatusSuccessStr    = "SUCCESS"
	taskStatusFailedStr     = "FAILED"
	taskStatusCancelledStr  = "CANCELLED"
)

func (status TaskStatus) String() string {
	switch status {
	case TaskStatusPending:
		return taskStatusPendingStr
	case TaskStatusInProgress:
		return taskStatusInProgressStr
	case TaskStatusSuccess:
		return taskStatusSuccessStr
	case TaskStatusFailed:
		return taskStatusFailedStr
	case TaskStatusCancelled:
		return taskStatusCancelledStr
	default:
		return taskStatusUndefinedStr
	}
}

func parseTaskStatus(status string) (TaskStatus, error) {
	switch status {
	case taskStatusPendingStr:
		return TaskStatusPending, nil
	case taskStatusInProgressStr:
		return TaskStatusInProgress, nil
	case taskStatusSuccessStr:
		return TaskStatusSuccess, nil
	case taskStatusCancelledStr:
		return TaskStatusCancelled, nil
	case taskStatusFailedStr:
		return TaskStatusFailed, nil
	default:
		return TaskStatusUndefined, fmt.Errorf(
			"unexpected task status '%s'",
			status,
		)
	}
}
