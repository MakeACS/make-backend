package repos

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"make-backend/internal/database/models"
	"strings"
	"time"

	"github.com/lib/pq"
)

type DateRange struct {
	StartDate *time.Time
	EndDate   *time.Time
}

type LogFilter struct {
	Range                DateRange
	Categories           []string
	IncludeUncategorized bool
	SearchText           string
}

type MultiMakerspaceLogFilter struct {
	LogFilter
	Makerspaces []int
}

type AuditLogRepository interface {
	AllLogs(ctx context.Context, page int, limit int, filter MultiMakerspaceLogFilter) ([]models.AuditLog, error)
	MakerspaceLogs(ctx context.Context, makerspace_id int, page int, limit int, filter LogFilter) ([]models.AuditLog, error)

	// AllLogsRegex(ctx context.Context, page int, limit int, filter MultiMakerspaceLogFilter)
	// MakerspaceLogsRegex(ctx context.Context, makerspace_id int, page int, limit int, filter LogFilter)

	CreateAuditLog(ctx context.Context, makerspaceId *int, plainText string, formatText string, message_type string, data map[string]any) (int, error)
}

type AuditLogRepo struct {
	DB *sql.DB
	// TypeToCategory map[string]models.Category
}

func (r *AuditLogRepo) CreateAuditLog(ctx context.Context, makerspaceId *int, plainString string, formatString string, message_type string, data map[string]any) (int, error) {
	query := `INSERT INTO audit_logs (makerspace_id, plain_string, format_string, message_type, data) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	jsonData, err := json.Marshal(data)
	if err != nil {
		return -1, fmt.Errorf("failed to marshal json for audit log data: %w", err)
	}

	var id_result int

	err = r.DB.QueryRowContext(ctx, query, makerspaceId, plainString, formatString, message_type, jsonData).Scan(&id_result)
	if err != nil {
		return 0, fmt.Errorf("failed to insert audit log: %w", err)
	}

	return id_result, nil
}

func CategoriesToTypes(cats []string) []string {
	return cats
}

func (r *AuditLogRepo) AllLogs(ctx context.Context, page int, limit int, filter MultiMakerspaceLogFilter) ([]models.AuditLog, error) {
	conditions := []string{}
	args := []any{}
	addArgGetId := func(arg any) int {
		args = append(args, arg)
		return len(args)
	}
	query := "SELECT id, timestamp, makerspace_id, plain_string, format_string, message_type, data"
	if len(filter.Makerspaces) > 0 {
		id := addArgGetId(pq.Array(filter.Makerspaces))
		makerspaceFilter := fmt.Sprintf("makerspace_id = ANY([$%d])", id)
		conditions = append(conditions, makerspaceFilter)
	}
	if filter.Range.StartDate != nil {
		id := addArgGetId(filter.Range.StartDate)
		startDateFilter := fmt.Sprintf("timestamp >= $%d", id)
		conditions = append(conditions, startDateFilter)
	}
	if filter.Range.EndDate != nil {
		id := addArgGetId(filter.Range.EndDate)
		endDateFilter := fmt.Sprintf("timestamp <= $%d", id)
		conditions = append(conditions, endDateFilter)
	}

	types := CategoriesToTypes(filter.Categories)

	// nonempty &  includeUncategorized -> ANY[types] or NULL
	// nonempty & !includeUncategorized -> ANY[types]
	// empty &  includeUncategorized -> filter nothing, all through
	// empty & !includeUncategorized -> NOT NULL
	//
	if len(types) > 0 {
		id := addArgGetId(pq.Array(types))
		var uncategorizedFilter string = ""
		if !filter.IncludeUncategorized {
			uncategorizedFilter = "OR message_type IS NULL"
		}
		catFilter := fmt.Sprintf("message_type = ANY([$%d]) %s", id, uncategorizedFilter)
		conditions = append(conditions, catFilter)

	} else if !filter.IncludeUncategorized {
		catFilter := "message_type IS NOT NULL"
		conditions = append(conditions, catFilter)
	}

	if filter.SearchText != "" {
		searchText := "%" + filter.SearchText + "%"
		id := addArgGetId(searchText)
		textFilter := fmt.Sprintf("plain_text ILIKE $%d", id)
		conditions = append(conditions, textFilter)
	}

	fullQuery := query + " WHERE " + strings.Join(conditions, " AND")

	err := r.DB.QueryRowContext(ctx, fullQuery, args...)
	if err != nil {
		return []models.AuditLog{}, fmt.Errorf("failed to search audit logs: %w", err)

	}
	return []models.AuditLog{}, nil

}

func (r *AuditLogRepo) MakerspaceLogs(ctx context.Context, makerspace_id int, page int, limit int, filter LogFilter) ([]models.AuditLog, error) {
	return r.AllLogs(ctx, page, limit, MultiMakerspaceLogFilter{filter, []int{makerspace_id}})
}
