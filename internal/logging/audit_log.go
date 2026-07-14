package logging

import (
	"context"
	"fmt"
	"log/slog"
	"make-backend/internal/database"
	"make-backend/internal/database/models"
	"regexp"
)

type AuditLogger struct {
	store *database.Store
}

func (al *AuditLogger) Create(makerspaceId int, messageType string, fmtString string, entities ...models.LogEntity) {
	plain := CreatePlainString(fmtString, entities...)
	format := CreateFormatString(fmtString, entities...)
	data := DataForEntities(fmtString, entities)

	_, err := al.store.AuditLogs.CreateAuditLog(context.TODO(), &makerspaceId, plain, format, messageType, data)
	if err != nil {
		slog.Warn("Failed to create audit log with auto data", "err", err)
	}
}

func (al *AuditLogger) CreateUnassociated(messageType string, fmtString string, entities ...models.LogEntity) {
	plain := CreatePlainString(fmtString, entities...)
	format := CreateFormatString(fmtString, entities...)
	data := DataForEntities(fmtString, entities)

	_, err := al.store.AuditLogs.CreateAuditLog(context.TODO(), nil, plain, format, messageType, data)
	if err != nil {
		slog.Warn("Failed to create unassociated audit log with auto data", "err", err)
	}
}

func (al *AuditLogger) CreateWithData(makerspaceId int, messageType string, fmtString string, entities ...models.LogEntity) {
	plain := CreatePlainString(fmtString, entities...)
	format := CreateFormatString(fmtString, entities...)
	data := DataForEntities(fmtString, entities)

	_, err := al.store.AuditLogs.CreateAuditLog(context.TODO(), &makerspaceId, plain, format, messageType, data)
	if err != nil {
		slog.Warn("Failed to create audit log with manual data", "err", err)
	}
}
func (al *AuditLogger) CreateUnassociatedWithData(messageType string, data map[string]any, fmtString string, entities ...models.LogEntity) {
	plain := CreatePlainString(fmtString, entities...)
	format := CreateFormatString(fmtString, entities...)

	_, err := al.store.AuditLogs.CreateAuditLog(context.TODO(), nil, plain, format, messageType, data)
	if err != nil {
		slog.Warn("Failed to create unassociated audit log with manual data", "err", err)
	}

}

var auditLogRegex = regexp.MustCompile(`{(\w+)}`)

func addMappingTypesToObject(eType string, entities []models.LogEntity, target map[string]any) {
	if len(entities) == 1 {
		target[eType+"_id"] = entities[0].Id
		target[eType+"_label"] = entities[0].Label
		return
	}
	for i, entity := range entities {
		target[fmt.Sprintf("%s_%d_id", eType, i)] = entity.Id
		target[fmt.Sprintf("%s_%d_label", eType, i)] = entity.Label
	}
}

func DataForEntities(fmtString string, entities []models.LogEntity) map[string]any {
	var mapping map[string][]models.LogEntity = map[string][]models.LogEntity{}
	locations := auditLogRegex.FindAllString(fmtString, -1)

	for i, entityType := range locations {
		if i > len(entities) {
			slog.Warn("Invalid audit log format string - too many specifiers without entities", "num_specifiers", len(locations), "num_entities", len(entities))
			break
		}
		if len(entityType) > 2 {
			entityType = entityType[1 : len(entityType)-1]
		}

		mapping[entityType] = append(mapping[entityType], entities[i])
	}

	data := map[string]any{}
	for entityType, sortedEntities := range mapping {
		addMappingTypesToObject(entityType, sortedEntities, data)
	}
	return data
}

func CreatePlainString(fmtString string, entities ...models.LogEntity) string {
	index := 0
	replacer := func(s string) string {
		if len(s) > 2 {
			s = s[1 : len(s)-1]
		}
		var entity models.LogEntity
		if index > len(entities) {
			entity = models.LogEntity{Id: -1, Label: "unknown"}
		} else {
			entity = entities[index]
		}
		index += 1
		return entity.Label
	}
	return auditLogRegex.ReplaceAllStringFunc(fmtString, replacer)
}
func CreateFormatString(fmtString string, entities ...models.LogEntity) string {
	index := 0
	replacer := func(s string) string {
		if len(s) > 2 {
			s = s[1 : len(s)-1]
		}
		var entity models.LogEntity
		if index > len(entities) {
			entity = models.LogEntity{Id: -1, Label: "unknown"}
		} else {
			entity = entities[index]
		}
		index += 1
		return fmt.Sprintf("<%s:%d:%s>", s, entity.Id, entity.Label)
	}
	return auditLogRegex.ReplaceAllStringFunc(fmtString, replacer)

}
