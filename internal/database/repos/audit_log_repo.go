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
	Range         DateRange
	Categories    []string
	AllCategories bool
	SearchText    string
}

type MultiMakerspaceLogFilter struct {
	LogFilter
	Makerspaces         []int
	IncludeUnassociated bool
}

// mapping of log message category to the specific types of messages to look for
func CategoriesToTypes(cats []string) []string {
	return cats
}

type AuditLogRepository interface {
	GetLogs(ctx context.Context, page int, limit int, filter MultiMakerspaceLogFilter) ([]models.AuditLog, error)
	MakerspaceLogs(ctx context.Context, makerspace_id int, page int, limit int, filter LogFilter) ([]models.AuditLog, error)

	CreateAuditLog(ctx context.Context, makerspaceId *int, plainText string, formatText string, message_type string, data map[string]any) (int, error)
}

type AuditLogRepo struct {
	DB *sql.DB
}

func (r *AuditLogRepo) CreateAuditLog(ctx context.Context, makerspaceId *int, plainString string, formatString string, message_type string, data map[string]any) (int, error) {
	query := `INSERT INTO audit_logs (makerspace_id, plain_string, format_string, message_type, data) VALUES ($1, $2, $3, $4, $5) RETURNING id`

	jsonData, err := json.Marshal(data)
	if err != nil {
		return -1, fmt.Errorf("failed to marshal json for audit log data: %w", err)
	}

	var id_result int

	err = r.DB.QueryRowContext(ctx, query, makerspaceId, plainString, formatString, message_type, json.RawMessage(jsonData)).Scan(&id_result)
	if err != nil {
		return 0, fmt.Errorf("failed to insert audit log: %w", err)
	}

	return id_result, nil
}

func (r *AuditLogRepo) GetLogs(ctx context.Context, offset int, limit int, filter MultiMakerspaceLogFilter) ([]models.AuditLog, error) {
	conditions := []string{}
	args := []any{}
	addArgGetIndex := func(arg any) int {
		args = append(args, arg)
		return len(args)
	}
	query := "SELECT id, timestamp, makerspace_id, plain_string, format_string, message_type, data FROM audit_logs"

	// not empty & includeUnassociated -> ANY or NULL
	// not empty & !includeUnassociated -> ANY and NOT NULL -> ANY filters out nulls
	//  empty & includeUnassociated -> IS NULL
	//  empty & !includeUnassociated -> always returns nothing
	if len(filter.Makerspaces) == 0 && !filter.IncludeUnassociated {
		// don't even bother querying
		return []models.AuditLog{}, nil
	} else if len(filter.Makerspaces) == 0 && filter.IncludeUnassociated {
		makerspaceFilter := "makerspace_id IS NULL"
		conditions = append(conditions, makerspaceFilter)
	} else {
		// have a list of makerspaces to check
		index := addArgGetIndex(pq.Array(filter.Makerspaces))
		var assocFilter string = ""
		if filter.IncludeUnassociated {
			assocFilter = "OR makerspace_id IS NULL"
		}
		makerspaceFilter := fmt.Sprintf("(makerspace_id = ANY($%d) %s)", index, assocFilter)
		conditions = append(conditions, makerspaceFilter)
	}

	if filter.Range.StartDate != nil {
		index := addArgGetIndex(filter.Range.StartDate)
		startDateFilter := fmt.Sprintf("timestamp >= $%d", index)
		conditions = append(conditions, startDateFilter)
	}
	if filter.Range.EndDate != nil {
		index := addArgGetIndex(filter.Range.EndDate)
		endDateFilter := fmt.Sprintf("timestamp <= $%d", index)
		conditions = append(conditions, endDateFilter)
	}

	types := CategoriesToTypes(filter.Categories)

	// AllCategories -> dont add this filter
	// nonempty &  !AllCategories -> ANY[cats2types(cats)]
	// empty & !AllCategories -> nothing ever returned

	if !filter.AllCategories {
		if len(types) == 0 {
			return []models.AuditLog{}, nil
		} else if len(types) > 0 {
			index := addArgGetIndex(pq.Array(types))

			catFilter := fmt.Sprintf("message_type = ANY($%d)", index)
			conditions = append(conditions, catFilter)
		}
	}

	if filter.SearchText != "" {
		searchText := "%" + filter.SearchText + "%"
		index := addArgGetIndex(searchText)
		textFilter := fmt.Sprintf("plain_text ILIKE $%d", index)
		conditions = append(conditions, textFilter)
	}

	pageInfo := fmt.Sprintf("LIMIT %d OFFSET %d", limit, offset)

	fullQuery := query + " WHERE " + strings.Join(conditions, " AND ") + " " + pageInfo

	rows, err := r.DB.QueryContext(ctx, fullQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to construct QueryContext: %w", err)
	}
	logs := []models.AuditLog{}
	for rows.Next() {
		var log models.AuditLog
		err := rows.Scan(
			&log.Id,
			&log.Timestamp,
			&log.MakerspaceID,
			&log.PlainString,
			&log.FormatString,
			&log.MessageType,
			&log.Data,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan for audit log row: %w", err)
		}
		logs = append(logs, log)

	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("failed to search audit logs: %w", err)
	}

	return logs, nil

}

func (r *AuditLogRepo) MakerspaceLogs(ctx context.Context, makerspace_id int, offset int, limit int, filter LogFilter) ([]models.AuditLog, error) {
	return r.GetLogs(ctx, offset, limit, MultiMakerspaceLogFilter{filter, []int{makerspace_id}, false})
}
