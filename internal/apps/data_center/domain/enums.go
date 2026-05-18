package domain

const (
	AppCode = "data-center"

	StatusPending    = "pending"
	StatusProcessing = "processing"
	StatusSuccess    = "success"
	StatusWarning    = "warning"
	StatusFailed     = "failed"

	AnomalyAIStatusPending   = "pending"
	AnomalyAIStatusAnalyzing = "analyzing"
	AnomalyAIStatusSuccess   = "success"
	AnomalyAIStatusFailed    = "failed"

	AnomalyTaskStatusNone       = "none"
	AnomalyTaskStatusGenerated  = "generated"
	AnomalyTaskStatusProcessing = "processing"
	AnomalyTaskStatusCompleted  = "completed"

	AnomalyReviewStatusNone     = "none"
	AnomalyReviewStatusPending  = "pending"
	AnomalyReviewStatusReviewed = "reviewed"

	AnomalyStatusPending    = "pending"
	AnomalyStatusConfirmed  = "confirmed"
	AnomalyStatusProcessing = "processing"
	AnomalyStatusIgnored    = "ignored"
	AnomalyStatusClosed     = "closed"

	TaskStatusPending    = "pending"
	TaskStatusProcessing = "processing"
	TaskStatusCompleted  = "completed"
	TaskStatusOverdue    = "overdue"
	TaskStatusClosed     = "closed"

	ReviewConclusionEffective   = "effective"
	ReviewConclusionWeak        = "weak"
	ReviewConclusionIneffective = "ineffective"
	ReviewConclusionFollowUp    = "follow_up"
)

var TaskStatusFlow = map[string][]string{
	TaskStatusPending:    {TaskStatusProcessing, TaskStatusClosed},
	TaskStatusProcessing: {TaskStatusCompleted, TaskStatusClosed},
	TaskStatusCompleted:  {TaskStatusClosed},
	TaskStatusOverdue:    {TaskStatusProcessing, TaskStatusCompleted, TaskStatusClosed},
}

func CanTransitTaskStatus(from string, to string) bool {
	if from == to {
		return true
	}
	for _, candidate := range TaskStatusFlow[from] {
		if candidate == to {
			return true
		}
	}
	return false
}
