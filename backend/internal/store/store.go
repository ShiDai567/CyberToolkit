package store

import "cybertoolkit/backend/internal/domain"

type Store interface {
	Stats() map[string]int
	ListCategories(visibleOnly bool) []domain.Category
	CategoryByID(id string) (domain.Category, bool)
	CreateCategory(input domain.Category) domain.Category
	UpdateCategory(id string, update domain.Category) (domain.Category, error)
	ListTags() []domain.Tag
	ListTools(filters domain.ToolFilters, admin bool) ([]domain.Tool, int)
	GetToolBySlug(slug string, admin bool) (domain.Tool, bool)
	GetToolByID(id string) (domain.Tool, bool)
	CreateTool(input domain.Tool) domain.Tool
	UpdateTool(id string, update domain.Tool) (domain.Tool, error)
	ArchiveTool(id string) error
	TagsForTool(toolID string) []domain.Tag
	RelatedTools(categoryID, exceptToolID string, limit int) []domain.Tool
	ReplaceToolTags(toolID string, tagNames []string)
	CreateSubmission(submission domain.Submission) domain.Submission
	Authenticate(account, password, ip, userAgent string) (string, string, domain.User, error)
	Register(username, email, password, displayName, ip, userAgent string) (string, string, domain.User, error)
	ValidateSession(token string) (domain.User, bool)
	DeleteSession(token string)
	RefreshSession(refreshToken, ip, userAgent string) (string, string, domain.User, error)
	FindUserByEmail(email string) (domain.User, bool)
	UpdateUserProfile(userID string, displayName string) (domain.User, error)
	UpdateUserPassword(userID, currentPassword, newPassword string) error
	RevokeUserSessions(userID string, exceptToken string) (int, error)
	ListUserSessions(userID string) ([]domain.Session, error)

	// Admin: user management
	ListUsers(page, pageSize int) ([]domain.User, int)
	UpdateUserRole(userID, role string) (domain.User, error)
	SetUserActive(userID string, active bool) error

	// Admin: submission management
	ListSubmissions(status string, page, pageSize int) ([]domain.Submission, int)
	ReviewSubmission(id, reviewerID, status, note string) (domain.Submission, error)

	// Admin: dashboard stats
	AdminStats() map[string]int

	// Admin: audit logs
	ListAuditLogs(page, pageSize int) ([]domain.AuditLog, int)
	CreateAuditLog(log domain.AuditLog)

	// Admin: session management
	ListSessions(page, pageSize int) ([]domain.Session, int)
	RevokeSession(accessToken string) error
}
