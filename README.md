GoAPIAnalyzer
tôi muốn tạo 1 dự án backend go. dự án này sẽ import 1 thư mục dự án backend go khác và nó sẽ phân tích ra toàn bộ api có trong dự án. sau đó người dùng có thể chọn bất kì 1 api để xem các func, type, struct... (các kiểu dữ liệu có trong go, DO GƯỜI DÙNG TẠO)

internal/infrastructure/router/account_routes.go:

```
package router

import (
	accountHandler "fluencybe/internal/app/handler/account"
	"fluencybe/pkg/middleware"

	"github.com/gin-gonic/gin"
)

// setupAccountRoutes configures all account-related routes
func (r *Router) setupAccountRoutes(
	userHandler *accountHandler.UserHandler,
	userDetailHandler *accountHandler.UserDetailHandler,
	developerHandler *accountHandler.DeveloperHandler,
) {
	api := r.Group("/v1")

	// User routes
	user := api.Group("/user")
	{
		// Registration and Authentication
		user.POST("/register", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.Register(c.Request.Context(), c.Writer, c.Request)
		}))
		user.POST("/register/social", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.SocialRegister(c.Request.Context(), c.Writer, c.Request)
		}))
		user.POST("/login", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.Login(c.Request.Context(), c.Writer, c.Request)
		}))
		user.POST("/login/social", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.SocialLogin(c.Request.Context(), c.Writer, c.Request)
		}))

		// User Profile Management
		user.GET("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetMyUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.GET("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.PUT("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.UpdateMyUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.UpdateUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.DELETE("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.DeleteMyUser(c.Request.Context(), c.Writer, c.Request)
		}))
		user.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.DeleteUser(c.Request.Context(), c.Writer, c.Request)
		}))

		// User Details Management
		userDetail := user.Group("/detail")
		{
			userDetail.POST("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
				userDetailHandler.Create(c.Request.Context(), c.Writer, c.Request)
			}))
			userDetail.GET("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
				userDetailHandler.GetByUserID(c.Request.Context(), c.Writer, c.Request)
			}))
			userDetail.PUT("", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
				userDetailHandler.Update(c.Request.Context(), c.Writer, c.Request)
			}))
		}

		// User Listing and Validation
		user.GET("/list", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetListUserWithPagination(c.Request.Context(), c.Writer, c.Request)
		}))
		user.GET("/search", middleware.UserOrDeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.GetByUsernameOrEmail(c.Request.Context(), c.Writer, c.Request)
		}))
		user.GET("/check-email", gin.HandlerFunc(func(c *gin.Context) {
			userHandler.CheckExitsGmail(c.Request.Context(), c.Writer, c.Request)
		}))

		user.POST("/connect-social", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.ConnectSocial(c.Request.Context(), c.Writer, c.Request)
		}))
		user.POST("/disconnect-social", middleware.UserAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			userHandler.DisconnectSocial(c.Request.Context(), c.Writer, c.Request)
		}))
	}

	// Developer routes
	developer := api.Group("/developer")
	{
		developer.POST("/register", gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.Register(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.POST("/login", gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.Login(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.GET("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.GetMyDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.GET("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.GetDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.PUT("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.UpdateMyDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.PUT("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.UpdateDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.DELETE("", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.DeleteMyDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.DELETE("/:id", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.DeleteDeveloper(c.Request.Context(), c.Writer, c.Request)
		}))
		developer.GET("/list", middleware.DeveloperAuthMiddleware(r.db), gin.HandlerFunc(func(c *gin.Context) {
			developerHandler.GetListDeveloperWithPagination(c.Request.Context(), c.Writer, c.Request)
		}))
	}
}

```

---

internal/app/handler/account/user_handler.go:

```
package account

import (
	"context"
	"encoding/json"
	accountDTO "fluencybe/internal/app/dto"
	accountModel "fluencybe/internal/app/model"
	accountService "fluencybe/internal/app/service/account"
	"fluencybe/internal/core/constants"
	"fluencybe/pkg/logger"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UserHandler struct {
	service *accountService.UserService
	logger  *logger.PrettyLogger
}

func NewUserHandler(service *accountService.UserService) *UserHandler {
	enabledLevels := make(map[logger.LogLevel]bool)
	enabledLevels[logger.LevelDebug] = true
	enabledLevels[logger.LevelInfo] = true
	enabledLevels[logger.LevelWarning] = true
	enabledLevels[logger.LevelError] = true

	return &UserHandler{
		service: service,
		logger:  logger.GetGlobalLogger(),
	}
}

func (h *UserHandler) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("register_start", nil, "Starting user registration process")

	var req accountDTO.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("register_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	user := &accountModel.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	h.logger.Debug("register_attempt", map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
	}, "Attempting to register new user")

	if err := h.service.Register(ctx, user); err != nil {
		h.logger.Error("register_failed", map[string]interface{}{
			"error":    err.Error(),
			"email":    req.Email,
			"username": req.Username,
		}, "User registration failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("register_success", map[string]interface{}{
		"user_id":  user.ID.String(),
		"email":    user.Email,
		"username": user.Username,
	}, "User registered successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:        user.ID,
		Email:     user.Email,
		Username:  user.Username,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	})
}

func (h *UserHandler) SocialRegister(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("social_register_start", nil, "Starting social user registration process")

	var req accountDTO.SocialRegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("social_register_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid request body",
		})
		return
	}

	user := &accountModel.User{
		ID:       uuid.New(),
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
	}

	// Set social ID based on type
	switch req.Type {
	case "google":
		user.GoogleID = req.SocialID
	case "facebook":
		user.FacebookID = req.SocialID
	default:
		h.logger.Error("social_register_invalid_type", map[string]interface{}{
			"type": req.Type,
		}, "Invalid social login type")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Invalid social login type",
		})
		return
	}

	h.logger.Debug("social_register_attempt", map[string]interface{}{
		"email":    req.Email,
		"username": req.Username,
		"type":     req.Type,
	}, "Attempting to register new social user")

	if err := h.service.Register(ctx, user); err != nil {
		h.logger.Error("social_register_failed", map[string]interface{}{
			"error":    err.Error(),
			"email":    req.Email,
			"username": req.Username,
		}, "Social user registration failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("social_register_success", map[string]interface{}{
		"user_id":  user.ID.String(),
		"email":    user.Email,
		"username": user.Username,
		"type":     req.Type,
	}, "Social user registered successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Username:   user.Username,
		GoogleID:   user.GoogleID,
		FacebookID: user.FacebookID,
		CreatedAt:  user.CreatedAt,
	})
}

func (h *UserHandler) Login(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("login_start", nil, "Starting user login process")

	var req accountDTO.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("login_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("login_attempt", map[string]interface{}{
		"email": req.Email,
	}, "Attempting user login")

	user, token, err := h.service.Login(ctx, req.Email, req.Password)
	if err != nil {
		h.logger.Error("login_failed", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
		}, "Login failed")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}

	h.logger.Info("login_success", map[string]interface{}{
		"email": req.Email,
	}, "User logged in successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.LoginResponse{
		ID:    user.ID,
		Token: token,
	})
}

func (h *UserHandler) SocialLogin(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("social_login_start", nil, "Starting social login process")

	var req accountDTO.SocialLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("social_login_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("social_login_attempt", map[string]interface{}{
		"email": req.Email,
		"type":  req.Type,
	}, "Attempting social login")

	user, token, err := h.service.SocialLogin(ctx, req.Email, req.SocialID, req.Type)
	if err != nil {
		h.logger.Error("social_login_failed", map[string]interface{}{
			"error": err.Error(),
			"email": req.Email,
			"type":  req.Type,
		}, "Social login failed")
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	h.logger.Info("social_login_success", map[string]interface{}{
		"email": req.Email,
		"type":  req.Type,
	}, "Social login successful")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.LoginResponse{
		ID:    user.ID,
		Token: token,
	})
}

func (h *UserHandler) GetUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get_user_start", nil, "Starting to get user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("get_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from URL parameter
	userID := ginCtx.Param("id")
	if userID == "" {
		h.logger.Error("get_user_invalid_id", nil, "User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("get_user_attempt", map[string]interface{}{
		"user_id": userID,
	}, "Attempting to get user")

	user, err := h.service.GetUser(ctx, userID)
	if err != nil {
		h.logger.Error("get_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to get user")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.logger.Info("get_user_success", map[string]interface{}{
		"user_id": userID,
	}, "User retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Username:   user.Username,
		GoogleID:   user.GoogleID,
		FacebookID: user.FacebookID,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	})
}

func (h *UserHandler) GetMyUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get_my_user_start", nil, "Starting to get my user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("get_my_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, ok := ginCtx.Get("user_id")
	if !ok {
		h.logger.Error("get_my_user_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("get_my_user_invalid_id_type", nil, "Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("get_my_user_attempt", map[string]interface{}{
		"user_id": userIDStr,
	}, "Attempting to get user details")

	user, err := h.service.GetUser(ctx, userIDStr)

	if err == nil {
		h.logger.Debug("get_my_user_data", map[string]interface{}{
			"user_id":     userIDStr,
			"google_id":   user.GoogleID,
			"facebook_id": user.FacebookID,
		}, "Retrieved user data")
	}
	if err != nil {
		h.logger.Error("get_my_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to get my user")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.logger.Info("get_my_user_success", map[string]interface{}{
		"user_id": userIDStr,
	}, "My user retrieved successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Username:   user.Username,
		GoogleID:   user.GoogleID,
		FacebookID: user.FacebookID,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	})
}

func (h *UserHandler) UpdateUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update_user_start", nil, "Starting to update user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("update_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from URL parameter
	userID := ginCtx.Param("id")
	if userID == "" {
		h.logger.Error("update_user_invalid_id", nil, "User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	var req accountDTO.FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("update_user_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("update_user_attempt", map[string]interface{}{
		"user_id": userID,
		"field":   req.Field,
		"value":   req.Value,
	}, "Attempting to update user")

	updates := make(map[string]interface{})
	updates[req.Field] = req.Value

	updatedUser, err := h.service.UpdateUser(ctx, userID, updates)
	if err != nil {
		h.logger.Error("update_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to update user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("update_user_success", map[string]interface{}{
		"user_id": userID,
	}, "User updated successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         updatedUser.ID,
		Email:      updatedUser.Email,
		Username:   updatedUser.Username,
		GoogleID:   updatedUser.GoogleID,
		FacebookID: updatedUser.FacebookID,
		CreatedAt:  updatedUser.CreatedAt,
	})
}

func (h *UserHandler) UpdateMyUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("update_my_user_start", nil, "Starting to update my user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("update_my_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, ok := ginCtx.Get("user_id")
	if !ok {
		h.logger.Error("update_my_user_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("update_my_user_invalid_id_type", nil, "Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var req accountDTO.FieldUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("update_my_user_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	h.logger.Debug("update_my_user_attempt", map[string]interface{}{
		"user_id": userIDStr,
		"field":   req.Field,
		"value":   req.Value,
	}, "Attempting to update my user")

	updates := make(map[string]interface{})
	updates[req.Field] = req.Value

	updatedUser, err := h.service.UpdateUser(ctx, userIDStr, updates)
	if err != nil {
		h.logger.Error("update_my_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to update my user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("update_my_user_success", map[string]interface{}{
		"user_id": userIDStr,
	}, "My user updated successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         updatedUser.ID,
		Email:      updatedUser.Email,
		Username:   updatedUser.Username,
		GoogleID:   updatedUser.GoogleID,
		FacebookID: updatedUser.FacebookID,
		CreatedAt:  updatedUser.CreatedAt,
	})
}

func (h *UserHandler) DeleteUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete_user_start", nil, "Starting to delete user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("delete_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from URL parameter
	userID := ginCtx.Param("id")
	if userID == "" {
		h.logger.Error("delete_user_invalid_id", nil, "User ID is required")
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("delete_user_attempt", map[string]interface{}{
		"user_id": userID,
	}, "Attempting to delete user")

	if err := h.service.DeleteUser(ctx, userID); err != nil {
		h.logger.Error("delete_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to delete user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("delete_user_success", map[string]interface{}{
		"user_id": userID,
	}, "User deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) DeleteMyUser(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("delete_my_user_start", nil, "Starting to delete my user")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("delete_my_user_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, ok := ginCtx.Get("user_id")
	if !ok {
		h.logger.Error("delete_my_user_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("delete_my_user_invalid_id_type", nil, "Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	h.logger.Debug("delete_my_user_attempt", map[string]interface{}{
		"user_id": userIDStr,
	}, "Attempting to delete my user")

	if err := h.service.DeleteUser(ctx, userIDStr); err != nil {
		h.logger.Error("delete_my_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to delete my user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("delete_my_user_success", map[string]interface{}{
		"user_id": userIDStr,
	}, "My user deleted successfully")

	w.WriteHeader(http.StatusNoContent)
}

func (h *UserHandler) GetListUserWithPagination(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("get_list_user_start", nil, "Starting to get list of users")

	// Get pagination parameters from query string
	page := r.URL.Query().Get("page")
	if page == "" {
		page = "1"
	}
	limit := r.URL.Query().Get("limit")
	if limit == "" {
		limit = "10"
	}

	h.logger.Debug("get_list_user_attempt", map[string]interface{}{
		"page":  page,
		"limit": limit,
	}, "Attempting to get list of users")

	users, total, err := h.service.GetUserList(ctx, page, limit)
	if err != nil {
		h.logger.Error("get_list_user_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to get list of users")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("get_list_user_success", map[string]interface{}{
		"total": total,
	}, "List of users retrieved successfully")

	// Convert users to response DTOs
	var userResponses []accountDTO.UserResponse
	for _, user := range users {
		userResponses = append(userResponses, accountDTO.UserResponse{
			ID:         user.ID,
			Email:      user.Email,
			Username:   user.Username,
			GoogleID:   user.GoogleID,
			FacebookID: user.FacebookID,
			CreatedAt:  user.CreatedAt,
			UpdatedAt:  user.UpdatedAt,
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"users": userResponses,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

func (h *UserHandler) CheckExitsGmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("check_email_start", nil, "Starting to check email existence")

	// Get email from query params
	email := r.URL.Query().Get("email")
	if email == "" {
		h.logger.Error("check_email_invalid_email", nil, "Email is required")
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Check if email exists
	exists, err := h.service.CheckEmailExists(ctx, email)
	if err != nil {
		h.logger.Error("check_email_failed", map[string]interface{}{
			"error": err.Error(),
			"email": email,
		}, "Failed to check email existence")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("check_email_success", map[string]interface{}{
		"email":  email,
		"exists": exists,
	}, "Email check completed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"exists": exists,
	})
}

func (h *UserHandler) ConnectSocial(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("connect_social_start", nil, "Starting to connect social account")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("connect_social_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, ok := ginCtx.Get("user_id")
	if !ok {
		h.logger.Error("connect_social_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("connect_social_invalid_id_type", nil, "Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var req struct {
		SocialID string `json:"social_id"`
		Type     string `json:"type"`
		Email    string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("connect_social_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate social type
	if req.Type != "google" && req.Type != "facebook" {
		h.logger.Error("connect_social_invalid_type", map[string]interface{}{
			"type": req.Type,
		}, "Invalid social type")
		http.Error(w, "Invalid social type. Must be 'google' or 'facebook'", http.StatusBadRequest)
		return
	}

	// Get current user to verify email
	user, err := h.service.GetUser(ctx, userIDStr)
	if err != nil {
		h.logger.Error("connect_social_get_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to get user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify email matches
	if user.Email != req.Email {
		h.logger.Error("connect_social_email_mismatch", map[string]interface{}{
			"user_email":   user.Email,
			"social_email": req.Email,
		}, "Email mismatch")
		http.Error(w, "Social account email does not match user email", http.StatusBadRequest)
		return
	}

	// Prepare updates based on social type
	updates := make(map[string]interface{})
	switch req.Type {
	case "google":
		updates["google_id"] = req.SocialID
	case "facebook":
		updates["facebook_id"] = req.SocialID
	}

	h.logger.Debug("connect_social_update", map[string]interface{}{
		"user_id": userIDStr,
		"type":    req.Type,
		"updates": updates,
	}, "Attempting to update user with social ID")

	// Update user
	updatedUser, err := h.service.UpdateUser(ctx, userIDStr, updates)

	h.logger.Debug("connect_social_update_result", map[string]interface{}{
		"user_id":     userIDStr,
		"type":        req.Type,
		"google_id":   updatedUser.GoogleID,
		"facebook_id": updatedUser.FacebookID,
	}, "User update result")
	if err != nil {
		h.logger.Error("connect_social_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
			"type":    req.Type,
		}, "Failed to connect social account")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("connect_social_success", map[string]interface{}{
		"user_id": userIDStr,
		"type":    req.Type,
	}, "Social account connected successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         updatedUser.ID,
		Email:      updatedUser.Email,
		Username:   updatedUser.Username,
		GoogleID:   updatedUser.GoogleID,
		FacebookID: updatedUser.FacebookID,
		CreatedAt:  updatedUser.CreatedAt,
		UpdatedAt:  updatedUser.UpdatedAt,
	})
}

func (h *UserHandler) GetByUsernameOrEmail(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("search_user_start", nil, "Starting to search user by username or email")

	// Get query from URL params
	query := r.URL.Query().Get("query")
	if query == "" {
		h.logger.Error("search_user_invalid_query", nil, "Query parameter is required")
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	h.logger.Debug("search_user_attempt", map[string]interface{}{
		"query": query,
	}, "Attempting to search user")

	user, err := h.service.GetByUsernameOrEmail(ctx, query)
	if err != nil {
		h.logger.Error("search_user_failed", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		}, "Failed to search user")
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	h.logger.Info("search_user_success", map[string]interface{}{
		"query": query,
		"found": user != nil,
	}, "User search completed successfully")

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         user.ID,
		Email:      user.Email,
		Username:   user.Username,
		GoogleID:   user.GoogleID,
		FacebookID: user.FacebookID,
		CreatedAt:  user.CreatedAt,
		UpdatedAt:  user.UpdatedAt,
	})
}

func (h *UserHandler) DisconnectSocial(ctx context.Context, w http.ResponseWriter, r *http.Request) {
	h.logger.Info("disconnect_social_start", nil, "Starting to disconnect social account")

	// Get gin context from request context
	ginCtx, ok := ctx.Value(constants.GinContextKey).(*gin.Context)
	if !ok {
		h.logger.Error("disconnect_social_invalid_context", nil, "Invalid context")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	// Get user ID from gin context
	userID, ok := ginCtx.Get("user_id")
	if !ok {
		h.logger.Error("disconnect_social_invalid_id", nil, "Unauthorized")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Convert userID to string
	userIDStr, ok := userID.(string)
	if !ok {
		h.logger.Error("disconnect_social_invalid_id_type", nil, "Internal Server Error")
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	var req struct {
		Type     string `json:"type"`
		Email    string `json:"email"`
		SocialID string `json:"social_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Error("disconnect_social_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate social type
	if req.Type != "google" && req.Type != "facebook" {
		h.logger.Error("disconnect_social_invalid_type", map[string]interface{}{
			"type": req.Type,
		}, "Invalid social type")
		http.Error(w, "Invalid social type. Must be 'google' or 'facebook'", http.StatusBadRequest)
		return
	}

	// Get current user to verify email
	user, err := h.service.GetUser(ctx, userIDStr)
	if err != nil {
		h.logger.Error("disconnect_social_get_user_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
		}, "Failed to get user")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Verify email matches
	if user.Email != req.Email {
		h.logger.Error("disconnect_social_email_mismatch", map[string]interface{}{
			"user_email":   user.Email,
			"social_email": req.Email,
		}, "Email mismatch")
		http.Error(w, "Social account email does not match user email", http.StatusBadRequest)
		return
	}

	// Verify social ID matches
	var currentSocialID string
	switch req.Type {
	case "google":
		currentSocialID = user.GoogleID
	case "facebook":
		currentSocialID = user.FacebookID
	}

	if currentSocialID != req.SocialID {
		h.logger.Error("disconnect_social_id_mismatch", map[string]interface{}{
			"current_id": currentSocialID,
			"social_id":  req.SocialID,
		}, "Social ID mismatch")
		http.Error(w, "Social account ID does not match", http.StatusBadRequest)
		return
	}

	// Prepare updates based on social type
	updates := make(map[string]interface{})
	switch req.Type {
	case "google":
		updates["google_id"] = nil
	case "facebook":
		updates["facebook_id"] = nil
	}

	// Update user
	updatedUser, err := h.service.UpdateUser(ctx, userIDStr, updates)
	if err != nil {
		h.logger.Error("disconnect_social_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userIDStr,
			"type":    req.Type,
		}, "Failed to disconnect social account")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	h.logger.Info("disconnect_social_success", map[string]interface{}{
		"user_id": userIDStr,
		"type":    req.Type,
	}, "Social account disconnected successfully")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(accountDTO.UserResponse{
		ID:         updatedUser.ID,
		Email:      updatedUser.Email,
		Username:   updatedUser.Username,
		GoogleID:   updatedUser.GoogleID,
		FacebookID: updatedUser.FacebookID,
		CreatedAt:  updatedUser.CreatedAt,
		UpdatedAt:  updatedUser.UpdatedAt,
	})
}

```

---

internal/app/model/account_model.go:

```
package model

import (
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type User struct {
	ID         uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email      string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username   string    `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Password   string    `gorm:"type:text" json:"-"`
	GoogleID   string    `gorm:"column:google_id;default:null" json:"google_id,omitempty"`
	FacebookID string    `gorm:"column:facebook_id;default:null" json:"facebook_id,omitempty"`
	CreatedAt  time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (u *User) GetGoogleID() *string {
	if u.GoogleID == "" {
		return nil
	}
	return &u.GoogleID
}

func (u *User) GetFacebookID() *string {
	if u.FacebookID == "" {
		return nil
	}
	return &u.FacebookID
}

type UserDetail struct {
	UserID        uuid.UUID      `json:"user_id"`
	Gender        string         `json:"gender,omitempty"`
	Age           int            `json:"age"`
	Interests     []string       `json:"interests"`
	LearningGoals []string       `json:"learning_goals"`
	Occupation    []string       `json:"occupation"`
	FacebookLink  sql.NullString `json:"facebook_link,omitempty"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`
}

func (ud *UserDetail) Validate() error {
	if ud.UserID == uuid.Nil {
		return errors.New("user_id is required")
	}
	if ud.Gender != "" && ud.Gender != "male" && ud.Gender != "female" && ud.Gender != "other" {
		return errors.New("gender must be 'male', 'female', or 'other'")
	}
	if ud.Age < 0 || ud.Age > 120 {
		return errors.New("age must be between 0 and 120")
	}
	return nil
}

type Developer struct {
	ID        uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
	Email     string    `gorm:"uniqueIndex;not null;size:255" json:"email"`
	Username  string    `gorm:"uniqueIndex;not null;size:255" json:"username"`
	Password  string    `gorm:"not null;type:text" json:"-"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

```

---

internal/app/dto/account_dto.go:

```
package dto

import (
	"time"

	"github.com/google/uuid"
)

//==============================================================================
// * =-=-=-=-=-=-=-=-=-=-=-=-=-= Account Management =-=-=-=-=-=-=-=-=-=-=-=-=-= *
//==============================================================================

//------------------------------------------------------------------------------
// * Authentication DTOs
//------------------------------------------------------------------------------

// ? Registration
type RegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=8,max=100"`
}

// ? Social Registration
type SocialRegisterRequest struct {
	Email    string `json:"email" validate:"required,email,max=255"`
	Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
	Password string `json:"password" validate:"required,min=8,max=100"`
	SocialID string `json:"social_id" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=google facebook"`
}

// ? Login
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

// ? Social Login
type SocialLoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	SocialID string `json:"social_id" validate:"required"`
	Type     string `json:"type" validate:"required,oneof=google facebook"`
}

type LoginResponse struct {
	ID    uuid.UUID `json:"id"`
	Token string    `json:"token"`
}

//------------------------------------------------------------------------------
// * User Profile DTOs
//------------------------------------------------------------------------------

// ? User Information
type UserResponse struct {
	ID         uuid.UUID `json:"id"`
	Email      string    `json:"email"`
	Username   string    `json:"username"`
	GoogleID   string    `json:"google_id,omitempty"`
	FacebookID string    `json:"facebook_id,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at,omitempty"`
}

// ? Profile Updates
type UpdateRequest struct {
	Email    *string `json:"email,omitempty" validate:"omitempty,email,max=255"`
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=50,alphanum"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=8,max=100"`
}

type FieldUpdateRequest struct {
	Field string `json:"field" validate:"required,oneof=username email password"`
	Value string `json:"value" validate:"required"`
}

//------------------------------------------------------------------------------
// * Error Handling DTOs
//------------------------------------------------------------------------------

type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details string `json:"details,omitempty"`
}

```

---

internal/app/service/account/user_service.go:

```
package account

import (
	"context"
	accountModel "fluencybe/internal/app/model"
	accountRepository "fluencybe/internal/app/repository/account"
	"fluencybe/internal/app/shared/apierrors"
	"fluencybe/internal/infrastructure/database/nrdb"
	"fluencybe/pkg/logger"
	"fluencybe/pkg/utils"
	"strconv"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	repo          *accountRepository.UserRepository
	logger        *logger.PrettyLogger
	leaderboardDB *nrdb.WrappedDB
	notebookDB    *nrdb.WrappedDB
	progressDB    *nrdb.WrappedDB
}

func NewUserService(
	repo *accountRepository.UserRepository,
	leaderboardDB *nrdb.WrappedDB,
	notebookDB *nrdb.WrappedDB,
	progressDB *nrdb.WrappedDB,
) *UserService {
	return &UserService{
		repo:          repo,
		logger:        logger.GetGlobalLogger(),
		leaderboardDB: leaderboardDB,
		notebookDB:    notebookDB,
		progressDB:    progressDB,
	}
}

func (s *UserService) hashPassword(password string) (string, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		s.logger.Error("hash_password_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to hash password")
		return "", apierrors.ErrServerError
	}
	return string(hashedPassword), nil
}

func (s *UserService) verifyPassword(hashedPassword, password string) error {
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password)); err != nil {
		return apierrors.NewAPIError("Sai mật khẩu")
	}
	return nil
}

func (s *UserService) Register(ctx context.Context, user *accountModel.User) error {
	s.logger.Info("register_start", map[string]interface{}{
		"email": user.Email,
	}, "Starting user registration")

	if user == nil || user.Email == "" || user.Username == "" {
		return apierrors.ErrInvalidInput
	}

	if user.Password != "" {
		hashedPassword, err := s.hashPassword(user.Password)
		if err != nil {
			return err
		}
		user.Password = hashedPassword
	}

	if err := s.repo.Create(ctx, user); err != nil {
		if err == accountRepository.ErrUserDuplicateEmail {
			return apierrors.ErrEmailExists
		}
		if err == accountRepository.ErrUserDuplicateUsername {
			return apierrors.ErrUsernameExists
		}
		if err == accountRepository.ErrUserDuplicateSocialID {
			return apierrors.ErrSocialAccountExists
		}
		return apierrors.ErrServerError
	}

	return nil
}

func (s *UserService) Login(ctx context.Context, email, password string) (*accountModel.User, string, error) {
	s.logger.Info("login_start", map[string]interface{}{
		"email": email,
	}, "Starting user login")

	if email == "" || password == "" {
		return nil, "", apierrors.ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == accountRepository.ErrUserNotFound {
			return nil, "", apierrors.NewAPIError("Email không tồn tại trên hệ thống")
		}
		return nil, "", apierrors.ErrServerError
	}

	if err := s.verifyPassword(user.Password, password); err != nil {
		return nil, "", err
	}

	token, err := utils.GenerateJWT(user.ID.String(), "user")
	if err != nil {
		s.logger.Error("login_token_generation_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to generate JWT token")
		return nil, "", apierrors.ErrServerError
	}

	return user, token, nil
}

func (s *UserService) SocialLogin(ctx context.Context, email, socialID, socialType string) (*accountModel.User, string, error) {
	s.logger.Info("social_login_start", map[string]interface{}{
		"email": email,
		"type":  socialType,
	}, "Starting social login process")

	if email == "" || socialID == "" || socialType == "" {
		return nil, "", apierrors.ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil && err != accountRepository.ErrUserNotFound {
		return nil, "", apierrors.ErrServerError
	}

	if err == nil {
		var storedSocialID string
		switch socialType {
		case "google":
			storedSocialID = user.GoogleID
		case "facebook":
			storedSocialID = user.FacebookID
		default:
			return nil, "", apierrors.NewAPIError("Invalid social login type")
		}

		if storedSocialID != socialID {
			return nil, "", apierrors.NewAPIError("Invalid social credentials")
		}
	}

	token, err := utils.GenerateJWT(user.ID.String(), "user")
	if err != nil {
		s.logger.Error("social_login_token_generation_failed", map[string]interface{}{
			"error": err.Error(),
		}, "Failed to generate JWT token")
		return nil, "", apierrors.ErrServerError
	}

	return user, token, nil
}

func (s *UserService) GetUser(ctx context.Context, id string) (*accountModel.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, apierrors.ErrInvalidInput
	}

	user, err := s.repo.Get(ctx, userID)
	if err != nil {
		if err == accountRepository.ErrUserNotFound {
			return nil, apierrors.ErrNotFound
		}
		return nil, apierrors.ErrServerError
	}

	return user, nil
}

func (s *UserService) UpdateUser(ctx context.Context, id string, updates map[string]interface{}) (*accountModel.User, error) {
	userID, err := uuid.Parse(id)
	if err != nil {
		return nil, apierrors.ErrInvalidInput
	}

	if len(updates) == 0 {
		return nil, apierrors.ErrInvalidInput
	}

	if password, ok := updates["password"].(string); ok {
		hashedPassword, err := s.hashPassword(password)
		if err != nil {
			return nil, err
		}
		updates["password"] = hashedPassword
	}

	if err := s.repo.Update(ctx, userID, updates); err != nil {
		if err == accountRepository.ErrUserDuplicateEmail {
			return nil, apierrors.ErrEmailExists
		}
		if err == accountRepository.ErrUserDuplicateUsername {
			return nil, apierrors.ErrUsernameExists
		}
		if err == accountRepository.ErrUserDuplicateSocialID {
			return nil, apierrors.ErrSocialAccountExists
		}
		return nil, apierrors.ErrServerError
	}

	updatedUser, err := s.repo.Get(ctx, userID)
	if err != nil {
		s.logger.Error("update_user_get_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to get updated user")
		return nil, apierrors.ErrServerError
	}

	if googleID, ok := updates["google_id"]; ok {
		if googleID == nil {
			updatedUser.GoogleID = ""
		} else if str, ok := googleID.(string); ok {
			updatedUser.GoogleID = str
		}
	}
	if facebookID, ok := updates["facebook_id"]; ok {
		if facebookID == nil {
			updatedUser.FacebookID = ""
		} else if str, ok := facebookID.(string); ok {
			updatedUser.FacebookID = str
		}
	}

	return updatedUser, nil
}

func (s *UserService) DeleteUser(ctx context.Context, id string) error {
	userID, err := uuid.Parse(id)
	if err != nil {
		return apierrors.ErrInvalidInput
	}

	// Delete from notebook database first using the handle_user_deletion function
	_, err = s.notebookDB.ExecContext(ctx, "SELECT handle_user_deletion($1)", userID)
	if err != nil {
		s.logger.Error("delete_user_notebook_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to delete user's notebook data")
		// Continue with other deletions even if notebook deletion fails
	}

	// Delete from leaderboard database
	var leaderboardDeleted bool
	err = s.leaderboardDB.QueryRowContext(ctx, "SELECT delete_user_leaderboard_data($1)", userID).Scan(&leaderboardDeleted)
	if err != nil {
		s.logger.Error("delete_user_leaderboard_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to delete user's leaderboard data")

		// Compensating transaction: Roll back notebookDB deletion if possible
		_, restoreErr := s.notebookDB.ExecContext(ctx, "SELECT handle_user_restore_deletion($1)", userID)
		if restoreErr != nil {
			s.logger.Error("rollback_notebook_deletion_failed", map[string]interface{}{
				"error":   restoreErr.Error(),
				"user_id": userID,
			}, "Failed to rollback notebook deletion after leaderboard delete error")
		}

		// Optionally return an error here if full consistency is required
		// return errors.New("user deletion failed on leaderboard DB; notebook deletion rolled back")
	}

	// Delete from progress database
	_, err = s.progressDB.ExecContext(ctx, "SELECT delete_user_lesson_progress($1)", userID)
	if err != nil {
		s.logger.Error("delete_user_progress_failed", map[string]interface{}{
			"error":   err.Error(),
			"user_id": userID,
		}, "Failed to delete user's progress data")
		// Continue with user deletion even if progress deletion fails
	}

	// Finally delete user from main database
	if err := s.repo.Delete(ctx, userID); err != nil {
		if err == accountRepository.ErrUserNotFound {
			return apierrors.ErrNotFound
		}
		return apierrors.ErrServerError
	}

	return nil
}

func (s *UserService) GetUserList(ctx context.Context, page string, limit string) ([]*accountModel.User, int64, error) {
	pageNum, err := strconv.Atoi(page)
	if err != nil || pageNum < 1 {
		return nil, 0, apierrors.ErrInvalidInput
	}

	limitNum, err := strconv.Atoi(limit)
	if err != nil || limitNum < 1 {
		return nil, 0, apierrors.ErrInvalidInput
	}

	offset := (pageNum - 1) * limitNum
	users, err := s.repo.GetList(ctx, limitNum, offset)
	if err != nil {
		return nil, 0, apierrors.ErrServerError
	}

	total, err := s.repo.GetTotalCount(ctx)
	if err != nil {
		return nil, 0, apierrors.ErrServerError
	}

	return users, total, nil
}

func (s *UserService) GetByUsernameOrEmail(ctx context.Context, query string) (*accountModel.User, error) {
	if query == "" {
		return nil, apierrors.ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, query)
	if err != nil && err != accountRepository.ErrUserNotFound {
		s.logger.Error("get_by_username_or_email_failed", map[string]interface{}{
			"error": err.Error(),
			"query": query,
		}, "Failed to search by email")
		return nil, apierrors.ErrServerError
	}

	if err == accountRepository.ErrUserNotFound {
		user, err = s.repo.GetByUsername(ctx, query)
		if err != nil {
			if err == accountRepository.ErrUserNotFound {
				return nil, apierrors.NewAPIError("User not found")
			}
			s.logger.Error("get_by_username_or_email_failed", map[string]interface{}{
				"error": err.Error(),
				"query": query,
			}, "Failed to search by username")
			return nil, apierrors.ErrServerError
		}
	}

	return user, nil
}

func (s *UserService) CheckEmailExists(ctx context.Context, email string) (bool, error) {
	if email == "" {
		return false, apierrors.ErrInvalidInput
	}

	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if err == accountRepository.ErrUserNotFound {
			return false, nil
		}
		return false, apierrors.ErrServerError
	}

	return user != nil, nil
}

pkg/logger/logger.go:
```

package logger

import (
"context"
"fmt"
"io"
"os"
"runtime"
"strings"
"sync"
"time"

    "github.com/fatih/color"
    "github.com/newrelic/go-agent/v3/newrelic"
    "go.uber.org/zap"

)

type LogLevel int

const (
LevelDebug LogLevel = iota
LevelInfo
LevelSuccess
LevelWarning
LevelError
LevelCritical
)

const (
ServiceAccount = "ACCOUNT"
ServiceCourse = "COURSE"
ServiceForm = "FORM"
ServiceQuestion = "QUESTION"
ServiceLeaderboard = "LEADERBOARD"
ServiceMatch = "MATCH"
ServiceNotebook = "NOTEBOOK"
ServiceProgress = "PROGRESS"
ServiceWiki = "WIKI"
ServiceOther = "OTHER"
)

type PrettyLogger struct {
level LogLevel
output io.Writer
service string
colorful bool
mu sync.Mutex
activeLevels map[LogLevel]bool
nrApp *newrelic.Application
zapLogger *zap.Logger
enableNRLog bool
ctx context.Context
}

func (l *PrettyLogger) WithContext(ctx context.Context) *PrettyLogger {
newLogger := \*l
newLogger.ctx = ctx
return &newLogger
}

type LogEntry struct {
Timestamp time.Time
Level LogLevel
EventCode string
Fields map[string]interface{}
Message string
}

func NewPrettyLogger(service string, level LogLevel, colorful bool, nrApp *newrelic.Application, enableNRLog bool) *PrettyLogger {
if service == "" {
service = ServiceOther
}
var zapLogger \*zap.Logger
if enableNRLog && nrApp != nil {
var err error
fmt.Fprintf(os.Stdout, "Initializing New Relic Zap logger integration...\n")
zapLogger, err = InitZapLogger(nrApp)
if err != nil {
fmt.Fprintf(os.Stderr, "Failed to initialize Zap logger: %v\n", err)
} else {
fmt.Fprintf(os.Stdout, "Successfully initialized New Relic Zap logger\n")
}
} else {
fmt.Fprintf(os.Stderr, "Skipping New Relic logger initialization: enableNRLog=%v, nrApp=%v\n", enableNRLog, nrApp != nil)
}
return &PrettyLogger{
level: level,
output: os.Stdout,
service: service,
colorful: colorful,
activeLevels: make(map[LogLevel]bool),
nrApp: nrApp,
zapLogger: zapLogger,
enableNRLog: enableNRLog,
ctx: context.Background(),
}
}

func (l \*PrettyLogger) SetLevel(level LogLevel) {
l.mu.Lock()
defer l.mu.Unlock()
l.level = level
}

func (l \*PrettyLogger) checkLogLevel(level LogLevel) bool {
env := strings.ToLower(os.Getenv("ENV"))
// Trong production, bỏ qua INFO, DEBUG, SUCCESS bất kể LOG*LEVELS
if env == "production" {
if level == LevelInfo || level == LevelDebug || level == LevelSuccess {
return false
}
}
logLevelsStr := os.Getenv("LOG_LEVELS")
if logLevelsStr != "" {
currentLevel := strings.ToUpper(levelToString(level))
levels := strings.Split(strings.ToUpper(logLevelsStr), ",")
for *, l := range levels {
if strings.TrimSpace(l) == currentLevel {
return true
}
}
return false
}
return true
}

func (l *PrettyLogger) Debug(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelDebug) {
return
}
l.log(LevelDebug, eventCode, fields, message)
}
func (l *PrettyLogger) Info(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelInfo) {
return
}
l.log(LevelInfo, eventCode, fields, message)
}
func (l *PrettyLogger) Success(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelSuccess) {
return
}
l.log(LevelSuccess, eventCode, fields, message)
}
func (l *PrettyLogger) Warning(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelWarning) {
return
}
l.log(LevelWarning, eventCode, fields, message)
}
func (l *PrettyLogger) Error(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelError) {
return
}
l.log(LevelError, eventCode, fields, message)
}
func (l *PrettyLogger) Critical(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelCritical) {
return
}
l.log(LevelCritical, eventCode, fields, message, false)
}

func (l \*PrettyLogger) log(level LogLevel, eventCode string, fields map[string]interface{}, message string, systemWideAlert ...bool) {
entry := LogEntry{
Timestamp: time.Now().UTC(),
Level: level,
EventCode: eventCode,
Fields: fields,
Message: message,
}
_, file, line, _ := runtime.Caller(2)
caller := fmt.Sprintf("%s:%d", shortenFilePath(file), line)
l.mu.Lock()
defer l.mu.Unlock()
if l.colorful {
l.printColorful(entry, caller)
} else {
l.printPlain(entry, caller)
}
if l.enableNRLog {
logger := l.zapLogger
if logger == nil {
logger = GetZapLogger()
}
if logger != nil {
zapFields := []zap.Field{
zap.String("event_code", entry.EventCode),
zap.String("service", l.service),
zap.String("environment", os.Getenv("ENV")),
zap.String("caller", caller),
}
for k, v := range entry.Fields {
if k != "timestamp" && k != "level" && k != "logger" {
zapFields = append(zapFields, zap.Any(k, v))
}
}
if txn := newrelic.FromContext(l.ctx); txn != nil {
if txnLogger, err := GetTransactionLogger(txn); err == nil {
logger = txnLogger
traceMetadata := txn.GetTraceMetadata()
if traceMetadata.TraceID != "" {
zapFields = append(zapFields, zap.String("trace.id", traceMetadata.TraceID))
}
if traceMetadata.SpanID != "" {
zapFields = append(zapFields, zap.String("span.id", traceMetadata.SpanID))
}
linkingMetadata := txn.GetLinkingMetadata()
if linkingMetadata.EntityGUID != "" {
zapFields = append(zapFields, zap.String("entity.guid", linkingMetadata.EntityGUID))
}
if linkingMetadata.Hostname != "" {
zapFields = append(zapFields, zap.String("hostname", linkingMetadata.Hostname))
}
if linkingMetadata.EntityName != "" {
zapFields = append(zapFields, zap.String("entity.name", linkingMetadata.EntityName))
}
}
}
switch entry.Level {
case LevelDebug:
logger.Debug(entry.Message, zapFields...)
case LevelInfo, LevelSuccess:
logger.Info(entry.Message, zapFields...)
case LevelWarning:
logger.Warn(entry.Message, zapFields...)
case LevelError:
logger.Error(entry.Message, zapFields...)
case LevelCritical:
logger.Fatal(entry.Message, zapFields...)
}
}
}
}

// Định nghĩa màu từng trường cho mỗi level
type LevelColors struct {
LevelColor *color.Color
OutputColor *color.Color
ServiceColor *color.Color
MessageColor *color.Color
FileColor \*color.Color
}

var levelColorMap = map[LogLevel]LevelColors{
LevelDebug: {
LevelColor: color.New(color.FgHiCyan),
OutputColor: color.New(color.FgCyan),
ServiceColor: color.New(color.FgBlue),
MessageColor: color.New(color.FgHiWhite),
FileColor: color.New(color.FgHiCyan),
},
LevelInfo: {
LevelColor: color.New(color.FgHiGreen),
OutputColor: color.New(color.FgGreen),
ServiceColor: color.New(color.FgHiMagenta),
MessageColor: color.New(color.FgWhite),
FileColor: color.New(color.FgHiGreen),
},
LevelSuccess: {
LevelColor: color.New(color.FgHiGreen),
OutputColor: color.New(color.FgHiWhite),
ServiceColor: color.New(color.FgGreen),
MessageColor: color.New(color.FgHiGreen),
FileColor: color.New(color.FgHiWhite),
},
LevelWarning: {
LevelColor: color.New(color.FgHiYellow),
OutputColor: color.New(color.FgYellow),
ServiceColor: color.New(color.FgHiYellow),
MessageColor: color.New(color.FgYellow),
FileColor: color.New(color.FgHiYellow),
},
LevelError: {
LevelColor: color.New(color.FgHiRed),
OutputColor: color.New(color.FgRed),
ServiceColor: color.New(color.FgHiRed),
MessageColor: color.New(color.FgRed),
FileColor: color.New(color.FgHiRed),
},
LevelCritical: {
LevelColor: color.New(color.BgRed, color.FgHiWhite),
OutputColor: color.New(color.BgRed, color.FgHiWhite),
ServiceColor: color.New(color.BgHiRed, color.FgHiWhite),
MessageColor: color.New(color.BgHiRed, color.FgHiWhite),
FileColor: color.New(color.BgRed, color.FgHiWhite),
},
}

func (l \*PrettyLogger) printColorful(entry LogEntry, caller string) {
timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
levelStr := strings.ToUpper(levelToString(entry.Level))
clr, ok := levelColorMap[entry.Level]
if !ok {
clr = LevelColors{
LevelColor: color.New(color.FgWhite),
OutputColor: color.New(color.FgWhite),
ServiceColor: color.New(color.FgWhite),
MessageColor: color.New(color.FgWhite),
FileColor: color.New(color.FgWhite),
}
}

    coloredTimestamp := clr.OutputColor.Sprint(timestamp)
    coloredLevel := clr.LevelColor.Sprint(levelStr)
    coloredService := clr.ServiceColor.Sprintf("[%s]", l.service)
    coloredCaller := clr.FileColor.Sprint(caller)
    coloredMessage := clr.MessageColor.Sprint(entry.Message)

    logLine := fmt.Sprintf("%s [%s] %s %s - %s",
    	coloredTimestamp,
    	coloredLevel,
    	coloredService,
    	coloredCaller,
    	coloredMessage,
    )
    if len(entry.Fields) > 0 {
    	fields := make([]string, 0, len(entry.Fields))
    	for k, v := range entry.Fields {
    		fields = append(fields, fmt.Sprintf("%s=%v", k, v))
    	}
    	logLine += color.New(color.FgHiBlack).Sprintf(" | %s", strings.Join(fields, " "))
    }
    if entry.EventCode != "" {
    	logLine = color.New(color.FgCyan).Sprintf("[%s] ", entry.EventCode) + logLine
    }
    fmt.Fprintln(l.output, logLine)

}

func (l \*PrettyLogger) printPlain(entry LogEntry, caller string) {
timestamp := entry.Timestamp.Format("2006-01-02 15:04:05.000")
levelStr := strings.ToUpper(levelToString(entry.Level))
logLine := fmt.Sprintf("%s [%s] [%s] %s - %s",
timestamp,
levelStr,
l.service,
caller,
entry.Message,
)
if len(entry.Fields) > 0 {
fields := make([]string, 0, len(entry.Fields))
for k, v := range entry.Fields {
fields = append(fields, fmt.Sprintf("%s=%v", k, v))
}
logLine += fmt.Sprintf(" | %s", strings.Join(fields, " "))
}
if entry.EventCode != "" {
logLine = fmt.Sprintf("[%s] %s", entry.EventCode, logLine)
}
fmt.Fprintln(l.output, logLine)
}

func levelToString(level LogLevel) string {
switch level {
case LevelDebug:
return "DEBUG"
case LevelInfo:
return "INFO"
case LevelSuccess:
return "SUCCESS"
case LevelWarning:
return "WARN"
case LevelError:
return "ERROR"
case LevelCritical:
return "CRITICAL"
default:
return "UNKNOWN"
}
}

// Helper functions to create service-specific loggers
func NewAccountLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceAccount,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewCourseLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceCourse,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewFormLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceForm,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewQuestionLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceQuestion,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewLeaderboardLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceLeaderboard,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewMatchLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceMatch,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewNotebookLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceNotebook,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewProgressLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceProgress,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewWikiLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceWiki,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}
func NewOtherLogger(parent *PrettyLogger) *PrettyLogger {
return &PrettyLogger{
level: parent.level,
output: parent.output,
service: ServiceOther,
colorful: parent.colorful,
activeLevels: parent.activeLevels,
nrApp: parent.nrApp,
zapLogger: parent.zapLogger,
enableNRLog: parent.enableNRLog,
ctx: parent.ctx,
}
}

func shortenFilePath(path string) string {
parts := strings.Split(path, "/")
if len(parts) > 3 {
return strings.Join(parts[len(parts)-3:], "/")
}
return path
}

```
<-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=-=->

ví dụ tôi chọn api là /user/register thì GoAPIAnalyzer với
node1:
```

func (h *UserHandler) Register(ctx context.Context, w http.ResponseWriter, r *http.Request) {
h.logger.Info("register_start", nil, "Starting user registration process")

    var req accountDTO.RegisterRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
    	h.logger.Error("register_invalid_body", map[string]interface{}{"error": err.Error()}, "Invalid request body")
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(http.StatusBadRequest)
    	json.NewEncoder(w).Encode(map[string]string{
    		"error": "Invalid request body",
    	})
    	return
    }

    user := &accountModel.User{
    	ID:       uuid.New(),
    	Email:    req.Email,
    	Username: req.Username,
    	Password: req.Password,
    }

    h.logger.Debug("register_attempt", map[string]interface{}{
    	"email":    req.Email,
    	"username": req.Username,
    }, "Attempting to register new user")

    if err := h.service.Register(ctx, user); err != nil {
    	h.logger.Error("register_failed", map[string]interface{}{
    		"error":    err.Error(),
    		"email":    req.Email,
    		"username": req.Username,
    	}, "User registration failed")
    	w.Header().Set("Content-Type", "application/json")
    	w.WriteHeader(http.StatusBadRequest)
    	json.NewEncoder(w).Encode(map[string]string{
    		"error": err.Error(),
    	})
    	return
    }

    h.logger.Info("register_success", map[string]interface{}{
    	"user_id":  user.ID.String(),
    	"email":    user.Email,
    	"username": user.Username,
    }, "User registered successfully")

    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(accountDTO.UserResponse{
    	ID:        user.ID,
    	Email:     user.Email,
    	Username:  user.Username,
    	CreatedAt: user.CreatedAt,
    	UpdatedAt: user.UpdatedAt,
    })

}

```

node2:
```

func (l \*PrettyLogger) Info(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelInfo) {
return
}
l.log(LevelInfo, eventCode, fields, message)
}

```

giải thích node2: vì ở dòng code này "h.logger.Info("register_start", nil, "Starting user registration process")" có h.logger.Info là hàm do người dùng tạo ra

node3:
```

type RegisterRequest struct {
Email string `json:"email" validate:"required,email,max=255"`
Username string `json:"username" validate:"required,min=3,max=50,alphanum"`
Password string `json:"password" validate:"required,min=8,max=100"`
}

```

node4:
```

func (l \*PrettyLogger) Error(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelError) {
return
}
l.log(LevelError, eventCode, fields, message)
}

```

node5:
```

type User struct {
ID uuid.UUID `gorm:"type:uuid;primaryKey;default:uuid_generate_v4()" json:"id"`
Email string `gorm:"uniqueIndex;not null;size:255" json:"email"`
Username string `gorm:"uniqueIndex;not null;size:255" json:"username"`
Password string `gorm:"type:text" json:"-"`
GoogleID string `gorm:"column:google_id;default:null" json:"google_id,omitempty"`
FacebookID string `gorm:"column:facebook_id;default:null" json:"facebook_id,omitempty"`
CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

```

node6:
```

func (l \*PrettyLogger) Debug(eventCode string, fields map[string]interface{}, message string) {
if !l.checkLogLevel(LevelDebug) {
return
}
l.log(LevelDebug, eventCode, fields, message)
}

```

node7:
```

func (s *UserService) Register(ctx context.Context, user *accountModel.User) error {
s.logger.Info("register_start", map[string]interface{}{
"email": user.Email,
}, "Starting user registration")

    if user == nil || user.Email == "" || user.Username == "" {
    	return apierrors.ErrInvalidInput
    }

    if user.Password != "" {
    	hashedPassword, err := s.hashPassword(user.Password)
    	if err != nil {
    		return err
    	}
    	user.Password = hashedPassword
    }

    if err := s.repo.Create(ctx, user); err != nil {
    	if err == accountRepository.ErrUserDuplicateEmail {
    		return apierrors.ErrEmailExists
    	}
    	if err == accountRepository.ErrUserDuplicateUsername {
    		return apierrors.ErrUsernameExists
    	}
    	if err == accountRepository.ErrUserDuplicateSocialID {
    		return apierrors.ErrSocialAccountExists
    	}
    	return apierrors.ErrServerError
    }

    return nil

}

```

và còn nhiều node khác. còn có chức năng lọc node loại từ ví dụ ta có thể blacklist file "logger.go" để ko tạo các node có hàm của logger.go
```

- yêu cầu:

* cho tôi fullcode các file trong dự án
