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
	paymentCollection  *mongo.Collection
	jobCollection      *mongo.Collection
	workflowCollection *mongo.Collection
	costCollection     *mongo.Collection
}

func NewAdminService() *AdminService {
	db := database.GetDatabase()
	return &AdminService{
		userCollection:     db.Collection("users"),
		paymentCollection:  db.Collection("payments"),
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

// GetCostStats returns cost analytics based on successful payments
func (as *AdminService) GetCostStats(params *models.AnalyticsParams) (*models.CostAnalytics, error) {
	ctx := context.Background()

	// Calculate date range
	var startDate, endDate time.Time
	if params.Period > 0 {
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -params.Period)
	} else if params.StartDate != "" && params.EndDate != "" {
		var err error
		startDate, err = time.Parse("2006-01-02", params.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start date format: %w", err)
		}
		endDate, err = time.Parse("2006-01-02", params.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end date format: %w", err)
		}
		endDate = endDate.Add(24 * time.Hour) // Include the end date
	} else {
		// Default to last 30 days
		endDate = time.Now()
		startDate = endDate.AddDate(0, 0, -30)
	}

	// Get all successful payments for overall stats
	totalSuccessfulPayments, err := as.getSuccessfulPaymentCount(ctx, time.Time{}, time.Time{})
	if err != nil {
		return nil, fmt.Errorf("failed to get total successful payments: %w", err)
	}

	// Get successful payments for the specified period
	periodSuccessfulPayments, err := as.getSuccessfulPaymentCount(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get period successful payments: %w", err)
	}

	// Calculate costs (each successful payment = 5,000,000 VND)
	const paymentAmount = 5000000
	totalCost := int64(totalSuccessfulPayments * paymentAmount)
	periodCost := int64(periodSuccessfulPayments * paymentAmount)

	// Split costs (you can adjust these ratios as needed)
	// Let's assume 60% is "execution cost" (revenue) and 40% is "infra cost" (fees, processing, etc.)
	overall := models.CostOverall{
		TotalCost:     totalCost,
		ExecutionCost: int64(float64(totalCost) * 0.6),  // 60% execution (revenue)
		InfraCost:     int64(float64(totalCost) * 0.4),  // 40% infra (costs)
	}

	period := models.CostPeriod{
		TotalCost:     periodCost,
		ExecutionCost: int64(float64(periodCost) * 0.6),
		InfraCost:     int64(float64(periodCost) * 0.4),
	}

	// Get daily payment amounts for the period
	dailyCosts, err := as.getDailyPaymentAmounts(ctx, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to get daily payment amounts: %w", err)
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

// Helper methods for payment calculations
func (as *AdminService) getSuccessfulPaymentCount(ctx context.Context, startDate, endDate time.Time) (int, error) {
	// First, let's count users who have owned: true (successful payments)
	filter := bson.M{"owned": true}
	
	// If date range is specified, add date filter
	if !startDate.IsZero() && !endDate.IsZero() {
		filter["updated_at"] = bson.M{
			"$gte": startDate,
			"$lt":  endDate,
		}
	}

	count, err := as.userCollection.CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}

	return int(count), nil
}

func (as *AdminService) getDailyPaymentAmounts(ctx context.Context, startDate, endDate time.Time) ([]models.DailyCost, error) {
	// Aggregate successful payments by day
	pipeline := []bson.M{
		{
			"$match": bson.M{
				"owned": true,
				"updated_at": bson.M{
					"$gte": startDate,
					"$lt":  endDate,
				},
			},
		},
		{
			"$group": bson.M{
				"_id": bson.M{
					"$dateToString": bson.M{
						"format": "%Y-%m-%d",
						"date":   "$updated_at",
					},
				},
				"count": bson.M{"$sum": 1},
			},
		},
		{
			"$project": bson.M{
				"_id":    1,
				"amount": bson.M{"$multiply": []interface{}{"$count", 5000000}}, // 5M VND per payment
			},
		},
		{
			"$sort": bson.M{"_id": 1},
		},
	}

	cursor, err := as.userCollection.Aggregate(ctx, pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var dailyCosts []models.DailyCost
	for cursor.Next(ctx) {
		var result models.DailyCost
		if err := cursor.Decode(&result); err != nil {
			continue // Skip invalid records
		}
		dailyCosts = append(dailyCosts, result)
	}

	// If no data found, return empty array instead of nil
	if dailyCosts == nil {
		dailyCosts = []models.DailyCost{}
	}

	return dailyCosts, nil
}
