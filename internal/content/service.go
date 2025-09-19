package content

import (
	"encoding/json"
	"cmsapp/internal/db"
)

func CreateEntry(entry *ContentEntry, schema Schema) error {
	if err := db.DB.Create(entry).Error; err != nil {
		return err
	}

	var entryData map[string]interface{}
	if err := json.Unmarshal(entry.Data, &entryData); err != nil {
		return err
	}

	for _, field := range schema.Fields {
		if field.Type == "relation" {
			if val, ok := entryData[field.Name]; ok {
				switch v := val.(type) {
				case float64: 
					rel := EntryRelation{
						SourceEntryID: entry.ID,
						FieldName:     field.Name,
						TargetEntryID: uint(v),
					}
					db.DB.Create(&rel)

				case []interface{}: 
					for _, id := range v {
						rel := EntryRelation{
							SourceEntryID: entry.ID,
							FieldName:     field.Name,
							TargetEntryID: uint(id.(float64)),
						}
						db.DB.Create(&rel)
					}
				}
				delete(entryData, field.Name)
			}
		}
	}

	cleanData, _ := json.Marshal(entryData)
	return db.DB.Model(entry).Update("data", cleanData).Error
}

func UpdateEntry(entry *ContentEntry, schema Schema, newData map[string]interface{}) error {

	db.DB.Where("source_entry_id = ?", entry.ID).Delete(&EntryRelation{})

	for _, field := range schema.Fields {
		if field.Type == "relation" {
			if val, ok := newData[field.Name]; ok {
				switch v := val.(type) {
				case float64:
					rel := EntryRelation{
						SourceEntryID: entry.ID,
						FieldName:     field.Name,
						TargetEntryID: uint(v),
					}
					db.DB.Create(&rel)

				case []interface{}:
					for _, id := range v {
						rel := EntryRelation{
							SourceEntryID: entry.ID,
							FieldName:     field.Name,
							TargetEntryID: uint(id.(float64)),
						}
						db.DB.Create(&rel)
					}
				}
				delete(newData, field.Name)
			}
		}
	}

	cleanData, _ := json.Marshal(newData)
	entry.Data = cleanData
	return db.DB.Save(entry).Error
}


