package repository

import (
	"infra/api/internal/infra/postgres"
	"testing"
)

func TestCreateEvent(t *testing.T) {
	r := InitEventsRepo()

	db := postgres.InitTest(postgres.TEST_CONFIG)

	err := r.Create(db, "webhook", 1, "{}")
	t.Log(err)

	err = r.Create(db, "webhook", 1, "{}")
	t.Log(err)

	err = r.Create(db, "webhook", 2, "{}")
	t.Log(err)

	err = r.Create(db, "invoice_processing", 1, "{}")
	t.Log(err)

	err = r.Create(db, "invoice_processing", 1, "{}")
	t.Log(err)

	err = r.Create(db, "invoice_processing", 2, "{}")
	t.Log(err)

}
