package services

import (
	"context"
	"fmt"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"jinzmedia-atmt/database"
	"jinzmedia-atmt/models"
)

type AdminService struct {
	userCollection     *mongo.Collection
	jobCollection      *mongo.Collection
	workflowCollection *mongo.Collection
	costCollection     *mongo.Collection
}

func NewAdminService() *AdminService {
	db := database.GetDatabase()
	return &AdminService{
		userCollection:     db.Collection("users"),
		jobCollection:      db.Collection("jobs"),
		workflowCollection: db.Collection("workflows"),
		costCollection:     db.Collection("costs"),
	}
}

// GetDashboardStats returns aggregated dashboard statistics
func (as *AdminService) GetDashboardStats() (*models.DashboardStats, error) {
	ctx := context.Background()

	// Get user stats
	userStats, err := as.getUserStats(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get user stats: %w", err)
	}

	// Get workflow stats (mock data for now)
	workflowStats := models.WorkflowStats{
		TotalWorkflows:  50,
		ActiveWorkflows: 45,
		FailedWorkflows: 5,
	}

	// Get job stats (mock data for now)
	jobStats := models.JobStats{
		TotalJobs:     500,
		RecentJobs:    50,
		RecentSuccess: 45,
	}

	// Get recent activity
	recentActivity, err := as.getRecentActivity(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get recent activity: %w", err)
	}

	return &models.DashboardStats{
		Users:          *userStats,
		Workflows:      workflowStats,
		Jobs:           jobStats,
		RecentActivity: *recentActivity,
	}, nil
}

// GetWorkflowStats returns workflow analytics
func (as *AdminService) GetWorkflowStats(params *models.AnalyticsParams) (*models.WorkflowAnalytics, error) {
	// Mock data for now - in production, you'd query your workflow database
	overall := models.WorkflowOverall{
		TotalWorkflows:  1200,
		ActiveWorkflows: 950,
		FailedWorkflows: 50,
	}

	period := models.WorkflowPeriod{
		TotalWorkflows:  300,
		ActiveWorkflows: 240,
		FailedWorkflows: 10,
	}

	// Generate mock daily data
	dailyWorkflows := []models.DailyWorkflow{
		{ID: "2024-03-01", Count: 12, Failed: 1},
		{ID: "2024-03-02", Count: 15, Failed: 0},
		{ID: "2024-03-03", Count: 18, Failed: 2},
	}

	return &models.WorkflowAnalytics{
		Overall:        overall,
		Period:         period,
		DailyWorkflows: dailyWorkflows,
	}, nil
}

// GetJobStats returns job analytics
func (as *AdminService) GetJobStats(params *models.AnalyticsParams) (*models.JobAnalytics, error) {
	// Mock data for now - in production, you'd query your job database
	overall := models.JobOverall{
		TotalJobs:   5500,
		SuccessJobs: 5000,
		FailedJobs:  200,
		QueuedJobs:  300,
	}

	period := models.JobPeriod{
		TotalJobs:   800,
		SuccessJobs: 740,
		FailedJobs:  20,
		QueuedJobs:  40,
	}

	// Generate mock daily data
	dailyJobs := []models.DailyJob{
		{ID: "2024-03-01", Count: 40, Success: 37, Failed: 1, Queued: 2},
		{ID: "2024-03-02", Count: 45, Success: 42, Failed: 2, Queued: 1},
		{ID: "2024-03-03", Count: 50, Success: 46, Failed: 1, Queued: 3},
	}

	return &models.JobAnalytics{
		Overall:   overall,
		Period:    period,
		DailyJobs: dailyJobs,
	}, nil
}

// GetCostStats returns cost analytics
func (as *AdminService) GetCostStats(params *models.AnalyticsParams) (*models.CostAnalytics, error) {
	// Mock data for now - in production, you'd query your cost database
	overall := models.CostOverall{
		TotalCost:     12500000,
		InfraCost:     7000000,
		ExecutionCost: 5500000,
	}

	period := models.CostPeriod{
		TotalCost:     3200000,
		InfraCost:     1800000,
		ExecutionCost: 1400000,
	}

	// Generate mock daily data
	dailyCosts := []models.DailyCost{
		{ID: "2024-03-01", Amount: 120000},
		{ID: "2024-03-02", Amount: 135000},
		{ID: "2024-03-03", Amount: 110000},
	}

	return &models.CostAnalytics{
		Overall:    overall,
		Period:     period,
		DailyCosts: dailyCosts,
	}, nil
}

// GetJobs returns paginated jobs list
func (as *AdminService) GetJobs(params *models.JobsParams) (*models.JobsList, error) {
	// Mock data for now - in production, you'd query your job database
	jobs := []models.Job{
		{
			ID:         primitive.NewObjectID(),
			Workflow:   "Build",
			Status:     "success",
			DurationMs: 12500,
			CreatedAt:  time.Now().Add(-1 * time.Hour),
			UpdatedAt:  time.Now().Add(-30 * time.Minute),
		},
		{
			ID:         primitive.NewObjectID(),
			Workflow:   "Test",
			Status:     "failed",
			DurationMs: 8500,
			CreatedAt:  time.Now().Add(-2 * time.Hour),
			UpdatedAt:  time.Now().Add(-90 * time.Minute),
		},
		{
			ID:         primitive.NewObjectID(),
			Workflow:   "Deploy",
			Status:     "running",
			DurationMs: 0,
			CreatedAt:  time.Now().Add(-10 * time.Minute),
			UpdatedAt:  time.Now().Add(-5 * time.Minute),
		},
	}

	// Filter by status if provided
	if params.Status != "" {
		var filteredJobs []models.Job
		for _, job := range jobs {
			if job.Status == params.Status {
				filteredJobs = append(filteredJobs, job)
			}
		}
		jobs = filteredJobs
	}

	// Apply pagination
	total := len(jobs)
	start := (params.Page - 1) * params.PageSize
	end := start + params.PageSize
	
	if start >= total {
		jobs = []models.Job{}
	} else if end > total {
		jobs = jobs[start:]
	} else {
		jobs = jobs[start:end]
	}

	return &models.JobsList{
		Items: jobs,
		Total: total,
	}, nil
}

// GetJobByID returns job details by ID
func (as *AdminService) GetJobByID(jobID string) (*models.Job, error) {
	// Mock data for now - in production, you'd query your job database
	job := &models.Job{
		ID:         primitive.NewObjectID(),
		Workflow:   "Build",
		Status:     "success",
		DurationMs: 12500,
		Logs:       []string{"Starting build...", "Installing dependencies...", "Running tests...", "Build completed successfully"},
		CreatedAt:  time.Now().Add(-1 * time.Hour),
		UpdatedAt:  time.Now().Add(-30 * time.Minute),
	}

	return job, nil
}

// GetWorkflows returns workflows list
func (as *AdminService) GetWorkflows() (*models.WorkflowsList, error) {
	// Mock data for now - in production, you'd query your workflow database
	workflows := []models.Workflow{
		{
			ID:     primitive.NewObjectID(),
			Name:   "Build",
			Active: true,
			Steps: []models.WorkflowStep{
				{Type: "http", Config: map[string]interface{}{"url": "https://api.example.com/build"}},
				{Type: "script", Config: map[string]interface{}{"command": "npm run build"}},
			},
			CreatedAt: time.Now().Add(-7 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-1 * time.Hour),
		},
		{
			ID:     primitive.NewObjectID(),
			Name:   "Test",
			Active: true,
			Steps: []models.WorkflowStep{
				{Type: "script", Config: map[string]interface{}{"command": "npm test"}},
			},
			CreatedAt: time.Now().Add(-5 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-2 * time.Hour),
		},
		{
			ID:     primitive.NewObjectID(),
			Name:   "Deploy",
			Active: false,
			Steps: []models.WorkflowStep{
				{Type: "http", Config: map[string]interface{}{"url": "https://api.example.com/deploy"}},
			},
			CreatedAt: time.Now().Add(-3 * 24 * time.Hour),
			UpdatedAt: time.Now().Add(-3 * time.Hour),
		},
	}

	return &models.WorkflowsList{
		Items: workflows,
		Total: len(workflows),
	}, nil
}

// CreateWorkflow creates a new workflow
func (as *AdminService) CreateWorkflow(req *models.CreateWorkflowRequest) (*models.Workflow, error) {
	now := time.Now()
	workflow := &models.Workflow{
		ID:        primitive.NewObjectID(),
		Name:      req.Name,
		Steps:     req.Steps,
		Active:    req.Active,
		CreatedAt: now,
		UpdatedAt: now,
	}

	// In production, you'd save to database
	log.Printf("ADMIN SERVICE: Created workflow %s", workflow.Name)

	return workflow, nil
}

// UpdateWorkflow updates an existing workflow
func (as *AdminService) UpdateWorkflow(workflowID string, req *models.UpdateWorkflowRequest) (*models.Workflow, error) {
	// In production, you'd query and update the database
	workflow := &models.Workflow{
		ID:        primitive.NewObjectID(),
		Name:      "Updated Workflow",
		Active:    true,
		UpdatedAt: time.Now(),
	}

	if req.Name != nil {
		workflow.Name = *req.Name
	}
	if req.Active != nil {
		workflow.Active = *req.Active
	}
	if req.Steps != nil {
		workflow.Steps = *req.Steps
	}

	log.Printf("ADMIN SERVICE: Updated workflow %s", workflowID)

	return workflow, nil
}

// Private helper methods
func (as *AdminService) getUserStats(ctx context.Context) (*models.UserStats, error) {
	// Count total users
	totalUsers, err := as.userCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return nil, err
	}

	// Count active users (users who logged in within last 30 days)
	thirtyDaysAgo := time.Now().Add(-30 * 24 * time.Hour)
	activeUsers, err := as.userCollection.CountDocuments(ctx, bson.M{
		"last_login": bson.M{"$gte": thirtyDaysAgo},
	})
	if err != nil {
		return nil, err
	}

	// Count verified users (users with owned: true)
	verifiedUsers, err := as.userCollection.CountDocuments(ctx, bson.M{
		"owned": true,
	})
	if err != nil {
		return nil, err
	}

	// Count admin users
	adminUsers, err := as.userCollection.CountDocuments(ctx, bson.M{
		"role": bson.M{"$in": []string{"admin", "super"}},
	})
	if err != nil {
		return nil, err
	}

	return &models.UserStats{
		TotalUsers:    int(totalUsers),
		ActiveUsers:   int(activeUsers),
		VerifiedUsers: int(verifiedUsers),
		AdminUsers:    int(adminUsers),
	}, nil
}

func (as *AdminService) getRecentActivity(ctx context.Context) (*models.RecentActivity, error) {
	// Get recent users (last 10 registered)
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}).SetLimit(10)
	cursor, err := as.userCollection.Find(ctx, bson.M{}, opts)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var recentUsers []models.RecentUser
	for cursor.Next(ctx) {
		var user models.RecentUser
		if err := cursor.Decode(&user); err != nil {
			continue // Skip invalid records
		}
		recentUsers = append(recentUsers, user)
	}

	// Mock recent jobs data (in production, you'd query your jobs collection)
	recentJobs := []models.RecentJob{
		{
			ID:        primitive.NewObjectID(),
			Workflow:  "Build",
			Status:    "success",
			CreatedAt: time.Now().Add(-10 * time.Minute),
		},
		{
			ID:        primitive.NewObjectID(),
			Workflow:  "Test",
			Status:    "running",
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        primitive.NewObjectID(),
			Workflow:  "Deploy",
			Status:    "queued",
			CreatedAt: time.Now().Add(-2 * time.Minute),
		},
	}

	return &models.RecentActivity{
		Users: recentUsers,
		Jobs:  recentJobs,
	}, nil
}
