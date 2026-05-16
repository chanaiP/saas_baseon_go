package services

import (
	"context"
	"net/url"
	"strings"
	"time"

	"saas_baseon_go/internal/apps/integration_center/repositories"
	"saas_baseon_go/internal/infrastructure/persistence/postgres/models"
)

type APICallLogRecorder interface {
	RecordAPICallLog(ctx context.Context, entry APICallLogEntry) error
}

type noopAPICallLogRecorder struct{}

func (noopAPICallLogRecorder) RecordAPICallLog(context.Context, APICallLogEntry) error {
	return nil
}

type RepositoryAPICallLogRecorder struct {
	repo *repositories.Repository
}

func NewRepositoryAPICallLogRecorder(repo *repositories.Repository) APICallLogRecorder {
	if repo == nil {
		return noopAPICallLogRecorder{}
	}
	return RepositoryAPICallLogRecorder{repo: repo}
}

func (r RepositoryAPICallLogRecorder) RecordAPICallLog(ctx context.Context, entry APICallLogEntry) error {
	if r.repo == nil {
		return nil
	}
	log := BuildAPICallLog(entry, time.Now())
	return r.repo.CreateAPICallLog(ctx, log)
}

func BuildAPICallLog(entry APICallLogEntry, calledAt time.Time) models.IntegrationAPICallLog {
	if strings.TrimSpace(entry.RequestID) == "" {
		entry.RequestID = gatewayRequestID(RequestMeta{})
	}
	if strings.TrimSpace(entry.TraceID) == "" {
		entry.TraceID = entry.RequestID
	}
	if strings.TrimSpace(entry.CallType) == "" {
		entry.CallType = "third_party_api"
	}
	safeEndpoint := strings.TrimSpace(entry.Endpoint)
	if safeEndpoint != "" {
		if parsed, err := url.Parse(safeEndpoint); err == nil {
			parsed.RawQuery = ""
			parsed.Fragment = ""
			safeEndpoint = parsed.String()
		}
	}
	errorMessage := redactSensitiveText(entry.ErrorMessage)
	log := models.IntegrationAPICallLog{
		TenantID:           entry.TenantID,
		TenantConnectionID: entry.TenantConnectionID,
		PlatformID:         entry.PlatformID,
		ProviderAppID:      entry.ProviderAppID,
		RequestID:          entry.RequestID,
		TraceID:            &entry.TraceID,
		CallType:           entry.CallType,
		Status:             normalizeAPICallLogStatus(entry.Status),
		DurationMS:         entry.DurationMS,
		CalledAt:           calledAt,
	}
	if strings.TrimSpace(entry.Method) != "" {
		log.Method = &entry.Method
	}
	if safeEndpoint != "" {
		log.Endpoint = &safeEndpoint
	}
	if entry.HTTPStatus > 0 {
		log.HTTPStatus = &entry.HTTPStatus
	}
	if entry.RequestDigest != "" {
		log.RequestDigest = &entry.RequestDigest
	}
	if entry.ResponseDigest != "" {
		log.ResponseDigest = &entry.ResponseDigest
	}
	if entry.ErrorCode != "" {
		log.ErrorCode = &entry.ErrorCode
	}
	if errorMessage != "" {
		log.ErrorMessage = &errorMessage
	}
	return log
}
