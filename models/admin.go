package models

import (
	"time"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Analytics Models
type AnalyticsParams struct {
	Period    int    `json:"period"`
	StartDate string `json:"startDate"`
	EndDate   string `json:"endDate"`
}

type JobsParams struct {
	Page     int    `json:"page"`
	PageSize int    `json:"pageSize"`
	Status   string `json:"status"`
	Search   string `json:"search"`
	Sort     string `json:"sort"`
}

// Dashboard Stats
type DashboardStats struct {
	Users           UserStats           `json:"users"`
	Workflows       WorkflowStats       `json:"workflows"`
	Jobs            JobStats            `json:"jobs"`
	RecentActivity  RecentActivity      `json:"recentActivity"`
}

type UserStats struct {
	TotalUsers    int `json:"totalUsers"`
	ActiveUsers   int `json:"activeUsers"`
	VerifiedUsers int `json:"verifiedUsers"`
	AdminUsers    int `json:"adminUsers"`
}

type WorkflowStats struct {
	TotalWorkflows  int `json:"totalWorkflows"`
	ActiveWorkflows int `json:"activeWorkflows"`
	FailedWorkflows int `json:"failedWorkflows"`
}

type JobStats struct {
	TotalJobs   int `json:"totalJobs"`
	RecentJobs  int `json:"recentJobs"`
	RecentSuccess int `json:"recentSuccess"`
}

type RecentActivity struct {
	Users []RecentUser `json:"users"`
	Jobs  []RecentJob  `json:"jobs"`
}

type RecentUser struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Name      string            `json:"name" bson:"full_name"`
	Email     string            `json:"email" bson:"email"`
	CreatedAt time.Time         `json:"createdAt" bson:"created_at"`
}

type RecentJob struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	Workflow  string            `json:"workflow" bson:"workflow"`
	Status    string            `json:"status" bson:"status"`
	CreatedAt time.Time         `json:"createdAt" bson:"created_at"`
}

// Analytics Response Models
type WorkflowAnalytics struct {
	Overall        WorkflowOverall    `json:"overall"`
	Period         WorkflowPeriod     `json:"period"`
	DailyWorkflows []DailyWorkflow    `json:"dailyWorkflows"`
}

type WorkflowOverall struct {
	TotalWorkflows  int `json:"totalWorkflows"`
	ActiveWorkflows int `json:"activeWorkflows"`
	FailedWorkflows int `json:"failedWorkflows"`
}

type WorkflowPeriod struct {
	TotalWorkflows  int `json:"totalWorkflows"`
	ActiveWorkflows int `json:"activeWorkflows"`
	FailedWorkflows int `json:"failedWorkflows"`
}

type DailyWorkflow struct {
	ID     string `json:"_id" bson:"_id"`
	Count  int    `json:"count" bson:"count"`
	Failed int    `json:"failed" bson:"failed"`
}

type JobAnalytics struct {
	Overall   JobOverall   `json:"overall"`
	Period    JobPeriod    `json:"period"`
	DailyJobs []DailyJob   `json:"dailyJobs"`
}

type JobOverall struct {
	TotalJobs   int `json:"totalJobs"`
	SuccessJobs int `json:"successJobs"`
	FailedJobs  int `json:"failedJobs"`
	QueuedJobs  int `json:"queuedJobs"`
}

type JobPeriod struct {
	TotalJobs   int `json:"totalJobs"`
	SuccessJobs int `json:"successJobs"`
	FailedJobs  int `json:"failedJobs"`
	QueuedJobs  int `json:"queuedJobs"`
}

type DailyJob struct {
	ID      string `json:"_id" bson:"_id"`
	Count   int    `json:"count" bson:"count"`
	Success int    `json:"success" bson:"success"`
	Failed  int    `json:"failed" bson:"failed"`
	Queued  int    `json:"queued" bson:"queued"`
}

type CostAnalytics struct {
	Overall    CostOverall   `json:"overall"`
	Period     CostPeriod    `json:"period"`
	DailyCosts []DailyCost   `json:"dailyCosts"`
}

type CostOverall struct {
	TotalCost     int64 `json:"totalCost"`
	InfraCost     int64 `json:"infraCost"`
	ExecutionCost int64 `json:"executionCost"`
}

type CostPeriod struct {
	TotalCost     int64 `json:"totalCost"`
	InfraCost     int64 `json:"infraCost"`
	ExecutionCost int64 `json:"executionCost"`
}

type DailyCost struct {
	ID     string `json:"_id" bson:"_id"`
	Amount int64  `json:"amount" bson:"amount"`
}

// Job Models
type Job struct {
	ID         primitive.ObjectID `json:"_id" bson:"_id"`
	Workflow   string            `json:"workflow" bson:"workflow"`
	Status     string            `json:"status" bson:"status"`
	DurationMs int64             `json:"durationMs" bson:"duration_ms"`
	Logs       []string          `json:"logs,omitempty" bson:"logs,omitempty"`
	CreatedAt  time.Time         `json:"createdAt" bson:"created_at"`
	UpdatedAt  time.Time         `json:"updatedAt" bson:"updated_at"`
}

type JobsList struct {
	Items []Job `json:"items"`
	Total int   `json:"total"`
}

// Workflow Models
type Workflow struct {
	ID     primitive.ObjectID   `json:"_id" bson:"_id"`
	Name   string              `json:"name" bson:"name"`
	Steps  []WorkflowStep      `json:"steps" bson:"steps"`
	Active bool                `json:"active" bson:"active"`
	CreatedAt time.Time        `json:"createdAt" bson:"created_at"`
	UpdatedAt time.Time        `json:"updatedAt" bson:"updated_at"`
}

type WorkflowStep struct {
	Type   string                 `json:"type" bson:"type"`
	Config map[string]interface{} `json:"config" bson:"config"`
}

type WorkflowsList struct {
	Items []Workflow `json:"items"`
	Total int        `json:"total"`
}

type CreateWorkflowRequest struct {
	Name   string         `json:"name"`
	Steps  []WorkflowStep `json:"steps"`
	Active bool           `json:"active"`
}

type UpdateWorkflowRequest struct {
	Name   *string         `json:"name,omitempty"`
	Steps  *[]WorkflowStep `json:"steps,omitempty"`
	Active *bool           `json:"active,omitempty"`
}
