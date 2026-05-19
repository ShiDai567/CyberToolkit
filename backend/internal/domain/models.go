package domain

import "time"

type Difficulty string

const (
	DifficultyBeginner     Difficulty = "beginner"
	DifficultyIntermediate Difficulty = "intermediate"
	DifficultyAdvanced     Difficulty = "advanced"
	DifficultyExpert       Difficulty = "expert"
)

type ToolStatus string

const (
	ToolStatusDraft     ToolStatus = "draft"
	ToolStatusPublished ToolStatus = "published"
	ToolStatusArchived  ToolStatus = "archived"
)

type SubmissionStatus string

const (
	SubmissionStatusPending  SubmissionStatus = "pending"
	SubmissionStatusApproved SubmissionStatus = "approved"
	SubmissionStatusRejected SubmissionStatus = "rejected"
)

type Category struct {
	ID          string    `json:"id"`
	Slug        string    `json:"slug"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        string    `json:"icon"`
	SortOrder   int       `json:"sortOrder"`
	IsVisible   bool      `json:"isVisible"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

type Tag struct {
	ID        string    `json:"id"`
	Slug      string    `json:"slug"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type Tool struct {
	ID               string     `json:"id"`
	Slug             string     `json:"slug"`
	Name             string     `json:"name"`
	ShortDescription string     `json:"shortDescription"`
	LongDescription  string     `json:"longDescription"`
	CategoryID       string     `json:"categoryId"`
	Difficulty       Difficulty `json:"difficulty"`
	Icon             string     `json:"icon"`
	Featured         bool       `json:"featured"`
	Status           ToolStatus `json:"status"`
	WebsiteURL       string     `json:"websiteUrl"`
	GitHubURL        string     `json:"githubUrl,omitempty"`
	ViewCount        int        `json:"viewCount"`
	FavoriteCount    int        `json:"favoriteCount"`
	PublishedAt      *time.Time `json:"publishedAt,omitempty"`
	CreatedAt        time.Time  `json:"createdAt"`
	UpdatedAt        time.Time  `json:"updatedAt"`
}

type ToolTag struct {
	ToolID string `json:"toolId"`
	TagID  string `json:"tagId"`
}

type Submission struct {
	ID             string           `json:"id"`
	Type           string           `json:"type"`
	SubmittedBy    string           `json:"submittedBy,omitempty"`
	ToolID         string           `json:"toolId,omitempty"`
	SubmitterEmail string           `json:"submitterEmail,omitempty"`
	Payload        map[string]any   `json:"payload"`
	Status         SubmissionStatus `json:"status"`
	ReviewerID     string           `json:"reviewerId,omitempty"`
	ReviewNote     string           `json:"reviewNote,omitempty"`
	CreatedAt      time.Time        `json:"createdAt"`
	ReviewedAt     *time.Time       `json:"reviewedAt,omitempty"`
}

type AuditLog struct {
	ID           string         `json:"id"`
	UserID       string         `json:"userId,omitempty"`
	Action       string         `json:"action"`
	ResourceType string         `json:"resourceType"`
	ResourceID   string         `json:"resourceId,omitempty"`
	BeforeData   map[string]any `json:"beforeData,omitempty"`
	AfterData    map[string]any `json:"afterData,omitempty"`
	CreatedAt    time.Time      `json:"createdAt"`
}

type User struct {
	ID           string     `json:"id"`
	Email        string     `json:"email"`
	DisplayName  string     `json:"displayName"`
	PasswordHash string     `json:"-"`
	Role         string     `json:"role"`
	IsActive     bool       `json:"isActive"`
	LastLoginAt  *time.Time `json:"lastLoginAt,omitempty"`
	CreatedAt    time.Time  `json:"createdAt"`
}

type ToolFilters struct {
	Query      string
	Category   string
	Difficulty string
	Tag        string
	Featured   *bool
	Page       int
	PageSize   int
	Sort       string
	Status     string
}
