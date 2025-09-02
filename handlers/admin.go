package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"jinzmedia-atmt/auth"
	"jinzmedia-atmt/models"
	"jinzmedia-atmt/services"
)

type AdminHandlers struct {
	adminService *services.AdminService
}

func NewAdminHandlers() *AdminHandlers {
	return &AdminHandlers{
		adminService: services.NewAdminService(),
	}
}

// Login handles admin authentication
func (h *AdminHandlers) Login(w http.ResponseWriter, r *http.Request) {
	var req models.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	// Basic validation
	if req.Email == "" || req.Password == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Email and password are required")
		return
	}

	authService := auth.NewAuthService()
	response, err := authService.Login(r.Context(), &req)
	if err != nil {
		if err == auth.ErrInvalidCredentials {
			writeErrorResponse(w, http.StatusUnauthorized, "Invalid email or password")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Login failed")
		return
	}

	// Check if user has admin role
	user := response.User
	if user.Role != "admin" && user.Role != "super" {
		writeErrorResponse(w, http.StatusForbidden, "Admin access required")
		return
	}

	log.Printf("ADMIN LOGIN: User %s logged in successfully", user.Email)
	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"token": response.Token,
		},
	})
}

// GetDashboardStats returns aggregated dashboard statistics
func (h *AdminHandlers) GetDashboardStats(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("ADMIN DASHBOARD: User %s requesting dashboard stats", user.Email)

	stats, err := h.adminService.GetDashboardStats()
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get dashboard stats: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get dashboard stats")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetWorkflowStats returns workflow analytics
func (h *AdminHandlers) GetWorkflowStats(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := extractAnalyticsParams(r)
	log.Printf("ADMIN ANALYTICS: User %s requesting workflow stats with params %+v", user.Email, params)

	stats, err := h.adminService.GetWorkflowStats(params)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get workflow stats: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get workflow stats")
		return
	}

	// Check if CSV export is requested
	if r.URL.Query().Get("export") == "csv" {
		h.exportWorkflowStatsCSV(w, stats)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetJobStats returns job analytics
func (h *AdminHandlers) GetJobStats(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := extractAnalyticsParams(r)
	log.Printf("ADMIN ANALYTICS: User %s requesting job stats with params %+v", user.Email, params)

	stats, err := h.adminService.GetJobStats(params)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get job stats: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get job stats")
		return
	}

	// Check if CSV export is requested
	if r.URL.Query().Get("export") == "csv" {
		h.exportJobStatsCSV(w, stats)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetCostStats returns cost analytics
func (h *AdminHandlers) GetCostStats(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := extractAnalyticsParams(r)
	log.Printf("ADMIN ANALYTICS: User %s requesting cost stats with params %+v", user.Email, params)

	stats, err := h.adminService.GetCostStats(params)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get cost stats: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get cost stats")
		return
	}

	// Check if CSV export is requested
	if r.URL.Query().Get("export") == "csv" {
		h.exportCostStatsCSV(w, stats)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    stats,
	})
}

// GetJobs returns paginated jobs list
func (h *AdminHandlers) GetJobs(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	params := extractJobsParams(r)
	log.Printf("ADMIN JOBS: User %s requesting jobs list with params %+v", user.Email, params)

	jobs, err := h.adminService.GetJobs(params)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get jobs: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get jobs")
		return
	}

	// Check if CSV export is requested
	if r.URL.Query().Get("export") == "csv" {
		h.exportJobsCSV(w, jobs.Items)
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    jobs,
	})
}

// GetJob returns job details by ID
func (h *AdminHandlers) GetJob(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	jobID := chi.URLParam(r, "id")
	if jobID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Job ID is required")
		return
	}

	log.Printf("ADMIN JOB: User %s requesting job details for ID %s", user.Email, jobID)

	job, err := h.adminService.GetJobByID(jobID)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get job %s: %v", jobID, err)
		if err.Error() == "job not found" {
			writeErrorResponse(w, http.StatusNotFound, "Job not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get job")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    job,
	})
}

// GetWorkflows returns workflows list
func (h *AdminHandlers) GetWorkflows(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	log.Printf("ADMIN WORKFLOWS: User %s requesting workflows list", user.Email)

	workflows, err := h.adminService.GetWorkflows()
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to get workflows: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to get workflows")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    workflows,
	})
}

// CreateWorkflow creates a new workflow
func (h *AdminHandlers) CreateWorkflow(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	var req models.CreateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	log.Printf("ADMIN WORKFLOW: User %s creating workflow %s", user.Email, req.Name)

	workflow, err := h.adminService.CreateWorkflow(&req)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to create workflow: %v", err)
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to create workflow")
		return
	}

	writeJSONResponse(w, http.StatusCreated, map[string]interface{}{
		"success": true,
		"data":    workflow,
	})
}

// UpdateWorkflow updates an existing workflow
func (h *AdminHandlers) UpdateWorkflow(w http.ResponseWriter, r *http.Request) {
	user := auth.GetUserFromContext(r.Context())
	if user == nil {
		writeErrorResponse(w, http.StatusUnauthorized, "User not authenticated")
		return
	}

	workflowID := chi.URLParam(r, "id")
	if workflowID == "" {
		writeErrorResponse(w, http.StatusBadRequest, "Workflow ID is required")
		return
	}

	var req models.UpdateWorkflowRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorResponse(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	log.Printf("ADMIN WORKFLOW: User %s updating workflow %s", user.Email, workflowID)

	workflow, err := h.adminService.UpdateWorkflow(workflowID, &req)
	if err != nil {
		log.Printf("ADMIN ERROR: Failed to update workflow %s: %v", workflowID, err)
		if err.Error() == "workflow not found" {
			writeErrorResponse(w, http.StatusNotFound, "Workflow not found")
			return
		}
		writeErrorResponse(w, http.StatusInternalServerError, "Failed to update workflow")
		return
	}

	writeJSONResponse(w, http.StatusOK, map[string]interface{}{
		"success": true,
		"data":    workflow,
	})
}

// Helper functions for parameter extraction
func extractAnalyticsParams(r *http.Request) *models.AnalyticsParams {
	query := r.URL.Query()
	
	params := &models.AnalyticsParams{}
	
	if period := query.Get("period"); period != "" {
		if p, err := strconv.Atoi(period); err == nil {
			params.Period = p
		}
	}
	
	params.StartDate = query.Get("startDate")
	params.EndDate = query.Get("endDate")
	
	return params
}

func extractJobsParams(r *http.Request) *models.JobsParams {
	query := r.URL.Query()
	
	params := &models.JobsParams{
		Page:     1,
		PageSize: 20,
		Sort:     "createdAt:desc",
	}
	
	if page := query.Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil && p > 0 {
			params.Page = p
		}
	}
	
	if pageSize := query.Get("pageSize"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil && ps > 0 && ps <= 100 {
			params.PageSize = ps
		}
	}
	
	params.Status = query.Get("status")
	params.Search = query.Get("search")
	
	if sort := query.Get("sort"); sort != "" {
		params.Sort = sort
	}
	
	return params
}

// CSV Export Methods
func (h *AdminHandlers) exportWorkflowStatsCSV(w http.ResponseWriter, stats *models.WorkflowAnalytics) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"workflow_stats.csv\"")
	
	csvData := "date,workflows,failed\n"
	for _, daily := range stats.DailyWorkflows {
		csvData += fmt.Sprintf("%s,%d,%d\n", daily.ID, daily.Count, daily.Failed)
	}
	w.Write([]byte(csvData))
}

func (h *AdminHandlers) exportJobStatsCSV(w http.ResponseWriter, stats *models.JobAnalytics) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"job_stats.csv\"")
	
	csvData := "date,total,success,failed,queued\n"
	for _, daily := range stats.DailyJobs {
		csvData += fmt.Sprintf("%s,%d,%d,%d,%d\n", daily.ID, daily.Count, daily.Success, daily.Failed, daily.Queued)
	}
	w.Write([]byte(csvData))
}

func (h *AdminHandlers) exportCostStatsCSV(w http.ResponseWriter, stats *models.CostAnalytics) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"cost_stats.csv\"")
	
	csvData := "date,amount\n"
	for _, daily := range stats.DailyCosts {
		csvData += fmt.Sprintf("%s,%d\n", daily.ID, daily.Amount)
	}
	w.Write([]byte(csvData))
}

func (h *AdminHandlers) exportJobsCSV(w http.ResponseWriter, jobs []models.Job) {
	w.Header().Set("Content-Type", "text/csv")
	w.Header().Set("Content-Disposition", "attachment; filename=\"jobs.csv\"")
	
	csvData := "id,workflow,status,duration_ms,created_at\n"
	for _, job := range jobs {
		csvData += fmt.Sprintf("%s,%s,%s,%d,%s\n", 
			job.ID.Hex(), job.Workflow, job.Status, job.DurationMs, job.CreatedAt.Format("2006-01-02 15:04:05"))
	}
	w.Write([]byte(csvData))
}
