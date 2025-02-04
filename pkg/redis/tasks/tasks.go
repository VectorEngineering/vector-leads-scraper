package tasks

import (
	"encoding/json"
	"time"

	"github.com/Vector/vector-leads-scraper/pkg/redis/priorityqueue"
)

type TaskType string

// Additional task types extending the existing ones in types.go
const (
	TypeLeadProcess    TaskType = "lead:process"
	TypeLeadValidate   TaskType = "lead:validate"
	TypeLeadEnrich     TaskType = "lead:enrich"
	TypeReportGenerate TaskType = "report:generate"
	TypeDataExport     TaskType = "data:export"
	TypeDataImport     TaskType = "data:import"
	TypeDataCleanup    TaskType = "data:cleanup"
	TypeScrapeGMaps    TaskType = "scrape:gmaps"  // done
	TypeEmailExtract   TaskType = "extract:email" // done
	TypeHealthCheck    TaskType = "health:check"
	TypeConnectionTest TaskType = "connection:test"
)

func (t TaskType) String() string {
	return string(t)
}

func DefaultTaskTypes() []string {
	return []string{
		TypeEmailExtract.String(),
		TypeScrapeGMaps.String(),
		TypeHealthCheck.String(),
		TypeConnectionTest.String(),
		TypeLeadProcess.String(),
		TypeLeadValidate.String(),
		TypeLeadEnrich.String(),
		TypeReportGenerate.String(),
		TypeDataExport.String(),
		TypeDataImport.String(),
		TypeDataCleanup.String(),
	}
}

// TaskPriorityLevel represents numeric priority levels for internal queue management
type TaskPriorityLevel int

const (
	// Priority Levels (higher number = higher priority)
	PriorityLevelLow      TaskPriorityLevel = 1
	PriorityLevelNormal   TaskPriorityLevel = 2
	PriorityLevelHigh     TaskPriorityLevel = 3
	PriorityLevelCritical TaskPriorityLevel = 4
)

// TaskConfig represents the configuration for a specific task type
type TaskConfig struct {
	Type        TaskType
	Priority    TaskPriorityLevel
	MaxRetries  int
	Timeout     time.Duration
	Description string
	// Queue determines which subscription tier can process this task
	Queue priorityqueue.SubscriptionType
}

// TaskConfigs maps task types to their configurations
var TaskConfigs = map[TaskType]TaskConfig{
	TypeEmailExtract: {
		Type:        TypeEmailExtract,
		Priority:    PriorityLevelHigh,
		MaxRetries:  3,
		Timeout:     5 * time.Minute,
		Description: "Extract and validate email addresses",
		Queue:       priorityqueue.SubscriptionEnterprise,
	},
	TypeScrapeGMaps: {
		Type:        TypeScrapeGMaps,
		Priority:    PriorityLevelNormal,
		MaxRetries:  5,
		Timeout:     15 * time.Minute,
		Description: "Scrape leads from Google Maps",
		Queue:       priorityqueue.SubscriptionPro,
	},
	TypeLeadProcess: {
		Type:        TypeLeadProcess,
		Priority:    PriorityLevelNormal,
		MaxRetries:  3,
		Timeout:     10 * time.Minute,
		Description: "Process and transform lead data",
		Queue:       priorityqueue.SubscriptionPro,
	},
	TypeLeadValidate: {
		Type:        TypeLeadValidate,
		Priority:    PriorityLevelNormal,
		MaxRetries:  3,
		Timeout:     5 * time.Minute,
		Description: "Validate lead data",
		Queue:       priorityqueue.SubscriptionPro,
	},
	TypeLeadEnrich: {
		Type:        TypeLeadEnrich,
		Priority:    PriorityLevelLow,
		MaxRetries:  3,
		Timeout:     10 * time.Minute,
		Description: "Enrich lead data with additional information",
		Queue:       priorityqueue.SubscriptionEnterprise,
	},
	TypeReportGenerate: {
		Type:        TypeReportGenerate,
		Priority:    PriorityLevelNormal,
		MaxRetries:  2,
		Timeout:     15 * time.Minute,
		Description: "Generate reports from lead data",
		Queue:       priorityqueue.SubscriptionPro,
	},
	TypeDataExport: {
		Type:        TypeDataExport,
		Priority:    PriorityLevelLow,
		MaxRetries:  3,
		Timeout:     30 * time.Minute,
		Description: "Export data to external systems",
		Queue:       priorityqueue.SubscriptionFree,
	},
	TypeDataImport: {
		Type:        TypeDataImport,
		Priority:    PriorityLevelNormal,
		MaxRetries:  3,
		Timeout:     30 * time.Minute,
		Description: "Import data from external systems",
		Queue:       priorityqueue.SubscriptionFree,
	},
	TypeDataCleanup: {
		Type:        TypeDataCleanup,
		Priority:    PriorityLevelLow,
		MaxRetries:  2,
		Timeout:     20 * time.Minute,
		Description: "Clean up old or invalid data",
		Queue:       priorityqueue.SubscriptionFree,
	},
	TypeHealthCheck: {
		Type:        TypeHealthCheck,
		Priority:    PriorityLevelCritical,
		MaxRetries:  5,
		Timeout:     1 * time.Minute,
		Description: "System health check",
		Queue:       priorityqueue.SubscriptionEnterprise,
	},
}

// GetTaskConfig returns the configuration for a given task type
func GetTaskConfig(taskType TaskType) (TaskConfig, bool) {
	config, exists := TaskConfigs[taskType]
	return config, exists
}

// TaskPayload represents the base interface for all task payloads
type TaskPayload interface {
	Validate() error
}

// ErrInvalidPayload represents a validation error in the task payload
type ErrInvalidPayload struct {
	Field   string
	Message string
}

func (e ErrInvalidPayload) Error() string {
	return "invalid payload: " + e.Field + " - " + e.Message
}

// NewTask creates a new task with the given type and payload
func NewTask(taskType string, payload TaskPayload) ([]byte, error) {
	if err := payload.Validate(); err != nil {
		return nil, err
	}

	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return payloadBytes, nil
}

// ParsePayload parses the payload bytes into the appropriate struct based on task type
func ParsePayload(taskType TaskType, payloadBytes []byte) (interface{}, error) {
	var payload interface{}

	switch taskType {
	case TypeEmailExtract:
		p := &EmailPayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		payload = p
	case TypeScrapeGMaps:
		p := &ScrapePayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		payload = p
	case TypeLeadProcess:
		p := &LeadProcessPayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	case TypeLeadValidate:
		p := &LeadValidatePayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	case TypeLeadEnrich:
		p := &LeadEnrichPayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	case TypeReportGenerate:
		p := &ReportGeneratePayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	case TypeDataExport:
		p := &DataExportPayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	case TypeDataImport:
		p := &DataImportPayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	case TypeDataCleanup:
		p := &DataCleanupPayload{}
		if err := json.Unmarshal(payloadBytes, p); err != nil {
			return nil, err
		}
		if err := p.Validate(); err != nil {
			return nil, err
		}
		payload = p
	default:
		return nil, ErrInvalidPayload{Field: "task_type", Message: "unsupported task type"}
	}

	return payload, nil
}
