package service

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/yc-alpha/admin/ent"
	"github.com/yc-alpha/logger"
)

// TenantHTTPHandler HTTP租户处理器
type TenantHTTPHandler struct {
	tenantService *SimpleTenantService
}

// NewTenantHTTPHandler 创建HTTP租户处理器
func NewTenantHTTPHandler(client *ent.Client) *TenantHTTPHandler {
	return &TenantHTTPHandler{
		tenantService: NewSimpleTenantService(client),
	}
}

// HTTPCreateTenantRequest HTTP创建租户请求
type HTTPCreateTenantRequest struct {
	Name     string `json:"name"`
	OwnerID  int64  `json:"owner_id"`
	Type     string `json:"type"`
	ParentID *int64 `json:"parent_id,omitempty"`
}

// HTTPCreateTenantResponse HTTP创建租户响应
type HTTPCreateTenantResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Tenant  *ent.Tenant `json:"tenant,omitempty"`
}

// TenantListResponse HTTP租户列表响应
type TenantListResponse struct {
	Success bool          `json:"success"`
	Message string        `json:"message"`
	Tenants []*ent.Tenant `json:"tenants"`
	Total   int           `json:"total"`
}

// TenantStatisticsResponse HTTP租户统计响应
type TenantStatisticsResponse struct {
	Success bool           `json:"success"`
	Message string         `json:"message"`
	Stats   map[string]int `json:"statistics"`
}

// CreateTenant HTTP创建租户
func (h *TenantHTTPHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req HTTPCreateTenantRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// 验证请求参数
	if req.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}

	if req.Type == "" {
		req.Type = "NORMAL" // 默认为普通租户
	}

	// 创建租户
	createdTenant, err := h.tenantService.CreateTenant(r.Context(), req.Name, req.OwnerID, req.Type)
	if err != nil {
		logger.Errorf("创建租户失败: %v", err)
		response := HTTPCreateTenantResponse{
			Success: false,
			Message: "创建租户失败: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := HTTPCreateTenantResponse{
		Success: true,
		Message: "租户创建成功",
		Tenant:  createdTenant,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListRootTenants HTTP获取根租户列表
func (h *TenantHTTPHandler) ListRootTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	tenants, err := h.tenantService.GetRootTenants(r.Context())
	if err != nil {
		logger.Errorf("获取根租户列表失败: %v", err)
		response := TenantListResponse{
			Success: false,
			Message: "获取根租户列表失败: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TenantListResponse{
		Success: true,
		Message: "查询成功",
		Tenants: tenants,
		Total:   len(tenants),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// ListSubTenants HTTP获取子租户列表
func (h *TenantHTTPHandler) ListSubTenants(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中获取父租户ID
	parentIDStr := r.URL.Query().Get("parent_id")
	if parentIDStr == "" {
		http.Error(w, "parent_id is required", http.StatusBadRequest)
		return
	}

	parentID, err := strconv.ParseInt(parentIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid parent_id", http.StatusBadRequest)
		return
	}

	tenants, err := h.tenantService.GetSubTenants(r.Context(), parentID)
	if err != nil {
		logger.Errorf("获取子租户列表失败: %v", err)
		response := TenantListResponse{
			Success: false,
			Message: "获取子租户列表失败: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TenantListResponse{
		Success: true,
		Message: "查询成功",
		Tenants: tenants,
		Total:   len(tenants),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTenantStatistics HTTP获取租户统计信息
func (h *TenantHTTPHandler) GetTenantStatistics(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.tenantService.GetTenantStatistics(r.Context())
	if err != nil {
		logger.Errorf("获取租户统计信息失败: %v", err)
		response := TenantStatisticsResponse{
			Success: false,
			Message: "获取租户统计信息失败: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := TenantStatisticsResponse{
		Success: true,
		Message: "查询成功",
		Stats:   stats,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetTenantByID HTTP根据ID获取租户
func (h *TenantHTTPHandler) GetTenantByID(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// 从URL路径中获取租户ID
	tenantIDStr := r.URL.Query().Get("id")
	if tenantIDStr == "" {
		http.Error(w, "id is required", http.StatusBadRequest)
		return
	}

	tenantID, err := strconv.ParseInt(tenantIDStr, 10, 64)
	if err != nil {
		http.Error(w, "Invalid id", http.StatusBadRequest)
		return
	}

	tenant, err := h.tenantService.GetTenantByID(r.Context(), tenantID)
	if err != nil {
		logger.Errorf("获取租户失败: %v", err)
		response := HTTPCreateTenantResponse{
			Success: false,
			Message: "获取租户失败: " + err.Error(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := HTTPCreateTenantResponse{
		Success: true,
		Message: "查询成功",
		Tenant:  tenant,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
