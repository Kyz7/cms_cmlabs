package content

import (
	"cmsapp/internal/auth"
	"cmsapp/internal/db"
	"github.com/gofiber/fiber/v2"
	"gorm.io/datatypes"
	"strings"
	"strconv"
	"encoding/json"
)

func CreateModelHandler(c *fiber.Ctx) error {
	var body struct {
		Name   string          `json:"name"`
		Schema Schema  `json:"schema"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	schemaJSON, err := json.Marshal(body.Schema)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	model := ContentModel{
		Name:   body.Name,
		Schema: schemaJSON,
	}

	if err := db.DB.Create(&model).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(fiber.Map{"message": "model created"})
}

func ListModelsHandler(c *fiber.Ctx) error {
	var models []ContentModel
	if err := db.DB.Find(&models).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error" : err.Error()})
	}
	return c.JSON(models)
}

func CreateEntryHandler(c *fiber.Ctx) error {
	var body struct {
		ModelID uint           `json:"model_id"`
		Data    datatypes.JSON `json:"data"`
		Status  string         `json:"status"`
		Slug    string         `json:"slug"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	var model ContentModel
	if err := db.DB.First(&model, body.ModelID).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "model not found"})
	}

	if err := ValidateEntry(model.Schema, body.Data); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err.Error()})
	}

	claims := c.Locals("user").(*auth.Claims)
	authorID := claims.UserID

	slug := body.Slug
	if slug == "" {
		var entryMap map[string]interface{}
		_ = json.Unmarshal(body.Data, &entryMap)
		if title, ok := entryMap["title"].(string); ok {
			slug = generateSlug(title)
		}
	}

	entry := ContentEntry{
		ModelID:  body.ModelID,
		Data:     body.Data,
		Status:   body.Status,
		Slug:     slug,
		AuthorID: authorID,
	}

	var schema Schema
	_ = json.Unmarshal(model.Schema, &schema)

	if err := CreateEntry(&entry, schema); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	_ = CreateAuditLog(entry.ID, claims.UserID, "created", nil)


	return c.JSON(entry)
}

func ListEntriesHandler(c *fiber.Ctx) error {
	modelID := c.Params("model_id")

	var entries []ContentEntry
	if err := db.DB.Where("model_id = ?", modelID).Find(&entries).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var result []map[string]interface{}
	for _, e := range entries {
		var entryData map[string]interface{}
		_ = json.Unmarshal(e.Data, &entryData)

		var rels []EntryRelation
		db.DB.Where("source_entry_id = ?", e.ID).Find(&rels)
		for _, rel := range rels {
			entryData[rel.FieldName] = append(entryData[rel.FieldName].([]uint), rel.TargetEntryID)
		}

		result = append(result, map[string]interface{}{
			"id":     e.ID,
			"slug":   e.Slug,
			"status": e.Status,
			"data":   entryData,
		})
	}

	return c.JSON(result)
}

func ListPublicEntriesHandler(c *fiber.Ctx) error {
	modelName := c.Params("model")

	var model ContentModel
	if err := db.DB.Where("name = ?", modelName).First(&model).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "model not found"})
	}

	var entries []ContentEntry
	if err := db.DB.Where("model_id = ? AND status = ?", model.ID, "published").Find(&entries).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	var result []map[string]interface{}
	for _, e := range entries {
		var entryData map[string]interface{}
		_ = json.Unmarshal(e.Data, &entryData)

		var rels []EntryRelation
		db.DB.Where("source_entry_id = ?", e.ID).Find(&rels)
		for _, rel := range rels {
			entryData[rel.FieldName] = append(entryData[rel.FieldName].([]uint), rel.TargetEntryID)
		}

		result = append(result, map[string]interface{}{
			"id":   e.ID,
			"slug": e.Slug,
			"data": entryData,
		})
	}

	return c.JSON(result)
}


func UpdateEntryHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Data   datatypes.JSON `json:"data"`
		Status string         `json:"status"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	var entry ContentEntry
	if err := db.DB.First(&entry, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "entry not found"})
	}

	var model ContentModel
	if err := db.DB.First(&model, entry.ModelID).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "failed to fetch model"})
	}

	var schema Schema
	_ = json.Unmarshal(model.Schema, &schema)

	if len(body.Data) > 0 {
		if err := ValidateEntry(model.Schema, body.Data); err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		 if entry.Status == "published" {
        entry.Status = "draft"
    	}

		var newData map[string]interface{}
		_ = json.Unmarshal(body.Data, &newData)

		if err := UpdateEntry(&entry, schema, newData); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
	}

	claims := c.Locals("user").(*auth.Claims)
	_ = CreateAuditLog(entry.ID, claims.UserID, "updated", nil)

	if body.Status != "" {
		userRole := c.Locals("user").(*auth.Claims).Role
		if userRole == "admin" {
			entry.Status = body.Status
			db.DB.Save(&entry)
		}
	}

	return c.JSON(entry)
}



func UpdateModelHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	var body struct {
		Name   string `json:"name"`
		Schema Schema `json:"schema"`
	}

	if err := c.BodyParser(&body); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid body"})
	}

	var model ContentModel
	if err := db.DB.First(&model, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "model not found"})
	}

	if body.Name != "" {
		model.Name = body.Name
	}

	if len(body.Schema.Fields) > 0 {
		schemaJSON, err := json.Marshal(body.Schema)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		model.Schema = schemaJSON
	}

	if err := db.DB.Save(&model).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(model)
}


func DeleteModelHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	db.DB.Where("model_id = ?", id).Delete(&ContentEntry{})

	if err := db.DB.Delete(&ContentModel{}, id).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error" : err.Error()})
	}
	return c.JSON(fiber.Map{"message" : "model deleted"})
}

func DeleteEntryHandler(c *fiber.Ctx) error {
	id := c.Params("id")

	var entry ContentEntry
	if err := db.DB.First(&entry, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "entry not found"})
	}
	if err := db.DB.Delete(&entry).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	claims := c.Locals("user").(*auth.Claims)
	_ = CreateAuditLog(entry.ID, claims.UserID, "deleted", map[string]interface{}{
		"data": entry.Data,
	})
	db.DB.Where("source_entry_id = ?", entry.ID).Delete(&EntryRelation{})
	return c.JSON(fiber.Map{"message": "entry deleted"})
}



func SetEntryStatusHandler(c *fiber.Ctx) error {
	userRole := c.Locals("user").(*auth.Claims).Role
	if userRole != "admin"{
    return c.Status(403).JSON(fiber.Map{"error": "forbidden"})}

	id := c.Params("id")
	action := c.Params("action")

	var entry ContentEntry
	if err := db.DB.First(&entry, id).Error; err != nil {
		return c.Status(404).JSON(fiber.Map{"error" : "entry not found"})
	}
	switch action {
	case "publish" :
		entry.Status = "published"
	case "unpublish" :
		entry.Status = "draft"
	case "suspend" :
		entry.Status = "suspend"
	default:
		return c.Status(400).JSON(fiber.Map{"error" : "invalid action"})
		
	}
	if err := db.DB.Save(&entry).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error" : err.Error()})
	}
	claims := c.Locals("user").(*auth.Claims)
	_ = CreateAuditLog(entry.ID, claims.UserID, "status_changed", map[string]interface{}{
	"new_status": entry.Status,
	})

	return c.JSON(entry)
}

func generateSlug(title string) string {
	slug := strings.ToLower(title)
	slug = strings.ReplaceAll(slug, " ", "-")
	return slug
}

func GetEntryHandler(c *fiber.Ctx) error {
	id, _ := strconv.Atoi(c.Params("id"))

	entry, relations, err := GetEntryWithRelations(uint(id))
	if err != nil {
		return c.Status(404).JSON(fiber.Map{"error": "entry not found"})
	}

	var entryData map[string]interface{}
	_ = json.Unmarshal(entry.Data, &entryData)

	for field, ids := range relations {
		if len(ids) == 1 {
			entryData[field] = ids[0]
		} else {
			entryData[field] = ids
		}
	}

	return c.JSON(fiber.Map{
		"id":     entry.ID,
		"slug":   entry.Slug,
		"status": entry.Status,
		"data":   entryData,
	})
}

func GetPublicEntryHandler(c *fiber.Ctx) error {
    slug := c.Params("slug")

    var entry ContentEntry
    if err := db.DB.Where("slug = ? AND status = ?", slug, "published").First(&entry).Error; err != nil {
        return c.Status(404).JSON(fiber.Map{"error": "entry not found"})
    }
    _, relations, _ := GetEntryWithRelations(entry.ID)

    var entryData map[string]interface{}
    _ = json.Unmarshal(entry.Data, &entryData)

    for field, ids := range relations {
        if len(ids) == 1 {
            entryData[field] = ids[0]
        } else {
            entryData[field] = ids
        }
    }

    return c.JSON(fiber.Map{
        "slug":   entry.Slug,
        "data":   entryData,
        "status": entry.Status,
    })
}

func CreateAuditLog(entryID, userID uint, action string, changes map[string]interface{}) error {
	log := AuditLog{
		EntryID: entryID,
		UserID:  userID,
		Action:  action,
	}
	if changes != nil {
		data, _ := json.Marshal(changes)
		log.Changes = data
	}
	return db.DB.Create(&log).Error
}

func GetAuditLogsHandler(c *fiber.Ctx) error {
	entryID := c.Params("id")

	var logs []AuditLog
	if err := db.DB.Where("entry_id = ?", entryID).Order("created_at desc").Find(&logs).Error; err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}

	return c.JSON(logs)
}

func GetEntryWithRelations(id uint) (*ContentEntry, map[string][]uint, error) {
	var entry ContentEntry
	if err := db.DB.First(&entry, id).Error; err != nil {
		return nil, nil, err
	}

	var rels []EntryRelation
	if err := db.DB.Where("source_entry_id = ?", id).Find(&rels).Error; err != nil {
		return &entry, nil, err
	}

	relations := make(map[string][]uint)
	for _, r := range rels {
		relations[r.FieldName] = append(relations[r.FieldName], r.TargetEntryID)
	}

	return &entry, relations, nil
}

