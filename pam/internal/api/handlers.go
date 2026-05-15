package api

import (
	"net/http"
	"strconv"

	"github.com/berndmarcel860-byte/ongame/pam/internal/models"
	"github.com/berndmarcel860-byte/ongame/pam/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// --- Auth handlers ---

func (r *Router) handleRegister(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	user, err := r.userSvc.Register(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrEmailTaken, service.ErrUsernameTaken:
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "registration failed"})
		}
		return
	}
	c.JSON(http.StatusCreated, user)
}

func (r *Router) handleLogin(c *gin.Context) {
	var req models.LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := r.userSvc.Login(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		case service.ErrUserBanned:
			c.JSON(http.StatusForbidden, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "login failed"})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

// --- User handlers ---

func (r *Router) handleGetMe(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	user, err := r.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (r *Router) handleGetMyTransactions(c *gin.Context) {
	userID := c.MustGet("userID").(uuid.UUID)
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))

	txns, err := r.userSvc.GetTransactionHistory(c.Request.Context(), userID, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch transactions"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"transactions": txns})
}

// --- PAM / Valkyrie handlers ---

func (r *Router) handleGetPlayerBalance(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	bal, err := r.balanceSvc.GetBalance(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "balance not found"})
		return
	}
	c.JSON(http.StatusOK, bal)
}

func (r *Router) handlePlayerTransaction(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	var req models.TransactionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := r.balanceSvc.ProcessTransaction(c.Request.Context(), userID, &req)
	if err != nil {
		switch err {
		case service.ErrInsufficientFunds:
			c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
		case service.ErrBalanceNotFound:
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "transaction failed"})
		}
		return
	}
	c.JSON(http.StatusOK, resp)
}

func (r *Router) handleCreateSession(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	var req models.CreateSessionRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	resp, err := r.gameSvc.CreateSession(c.Request.Context(), userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "session creation failed"})
		return
	}
	c.JSON(http.StatusCreated, resp)
}

func (r *Router) handleEndSession(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	sessionID, err := uuid.Parse(c.Param("sessionId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid sessionId"})
		return
	}
	if err := r.gameSvc.EndSession(c.Request.Context(), userID, sessionID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusNoContent, nil)
}

// --- Admin handlers ---

func (r *Router) handleAdminListUsers(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "50"))
	offset, _ := strconv.Atoi(c.DefaultQuery("offset", "0"))
	users, err := r.userSvc.ListUsers(c.Request.Context(), limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list users"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"users": users})
}

func (r *Router) handleAdminGetUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	user, err := r.userSvc.GetByID(c.Request.Context(), userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}
	c.JSON(http.StatusOK, user)
}

func (r *Router) handleAdminAdjustBalance(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	var req models.AdjustBalanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := r.balanceSvc.AdjustBalance(c.Request.Context(), userID, req.Amount, req.Reason); err != nil {
		switch err {
		case service.ErrInsufficientFunds:
			c.JSON(http.StatusPaymentRequired, gin.H{"error": err.Error()})
		default:
			c.JSON(http.StatusInternalServerError, gin.H{"error": "adjustment failed"})
		}
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "adjusted"})
}

func (r *Router) handleAdminBanUser(c *gin.Context) {
	userID, err := uuid.Parse(c.Param("userId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid userId"})
		return
	}
	var req models.BanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if err := r.userSvc.BanUser(c.Request.Context(), userID); err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "banned"})
}

func (r *Router) handleAdminDemoWinRate(c *gin.Context) {
	stats, err := r.balanceSvc.GetAdminDemoWinRate(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to load admin stats"})
		return
	}
	c.JSON(http.StatusOK, stats)
}
