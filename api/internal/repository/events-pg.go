package repository

import (
	"encoding/json"
	"fmt"
	"infra/api/internal/domain"
	"infra/api/internal/infra/postgres"

	"gorm.io/gorm"
)

type EventsRepo struct {
}

func InitEventsRepo() *EventsRepo {
	return &EventsRepo{}
}
func (r *EventsRepo) Create(tx *gorm.DB, eventType string, eventRelationID uint, payload string) error {
	if !json.Valid([]byte(payload)) {
		return fmt.Errorf("invalid payload: %s", payload)
	}

	_, err := r.Find(tx, eventRelationID, eventType)
	if err != nil {
		fmt.Println("CREATE ERROR: ", err)
		if !postgres.IsNotFound(err) {
			fmt.Println("IS NOT FOUND: ", err)
			return err
		}
		return tx.Create(&domain.Events{Type: eventType, RelationID: eventRelationID, Payload: payload, Status: "new"}).Error
	}
	return nil
}

func (r *EventsRepo) Done(tx *gorm.DB, eventRelationID uint, eventType string) error {
	return tx.Model(&domain.Events{}).Where(domain.Events{RelationID: eventRelationID, Type: eventType}).Update("status", "done").Error
}

func (r *EventsRepo) Find(tx *gorm.DB, eventRelationID uint, eventType string) (*domain.Events, error) {
	var existsEvent domain.Events
	return &existsEvent, tx.Where(&domain.Events{RelationID: eventRelationID, Type: eventType}).First(&existsEvent).Error
}
