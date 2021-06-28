package eventfaillog

import (
	"database/sql"
	"time"

	"github.com/helloferdie/stdgo/db"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
)

// EventFailLog -
type EventFailLog struct {
	ID               string       `db:"id" json:"id"`
	QueueName        string       `db:"queue_name" json:"queue_name"`
	AppID            string       `db:"app_id" json:"app_id"`
	FailoverEndpoint string       `db:"failover_endpoint" json:"failover_endpoint"`
	Payload          string       `db:"payload" json:"payload"`
	Remark           string       `db:"remark" json:"remark"`
	CreatedAt        sql.NullTime `db:"created_at" json:"created_at"`
	UpdatedAt        sql.NullTime `db:"updated_at" json:"updated_at"`
	DeletedAt        sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

var table = "event_fail_logs"

// Create -
func (ev *EventFailLog) Create(d *sqlx.DB) (string, error) {
	n := new(EventFailLog)
	ev.ID = n.GenerateID(d)
	ev.CreatedAt.Valid = true
	ev.CreatedAt.Time = time.Now().UTC()
	query, val := db.PrepareInsert(table, ev, []string{"updated_at", "deleted_at"})
	_, _, err := db.Exec(d, query, val)
	return ev.ID, err
}

// GenerateID -
func (ev *EventFailLog) GenerateID(d *sqlx.DB) string {
	nUUID := uuid.New()
	exist, _ := ev.GetByID(d, nUUID.String())
	for exist {
		nUUID = uuid.New()
		exist, _ = ev.GetByID(d, nUUID.String())
	}
	return nUUID.String()
}

// GetByID -
func (ev *EventFailLog) GetByID(d *sqlx.DB, id string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE id LIKE :id AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"id": id,
	}
	exist, err := db.Get(d, ev, query, values)
	return exist, err
}
