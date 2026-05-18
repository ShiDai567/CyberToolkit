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
	Authenticate(email, password string) (string, string, domain.User, error)
	Register(email, password, displayName string) (string, string, domain.User, error)
	ValidateSession(token string) (domain.User, bool)
	DeleteSession(token string)
	RefreshSession(refreshToken string) (string, string, domain.User, error)
	FindUserByEmail(email string) (domain.User, bool)
	UpdateUserProfile(userID string, displayName string) (domain.User, error)
	UpdateUserPassword(userID, currentPassword, newPassword string) error
	RevokeUserSessions(userID string, exceptToken string) (int, error)
}
