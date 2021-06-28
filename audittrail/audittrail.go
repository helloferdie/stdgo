package audittrail

import (
	"database/sql"
	"os"
	"strconv"
	"time"

	"github.com/helloferdie/stdgo/db"

	"github.com/helloferdie/stdgo/libquery"

	"github.com/jmoiron/sqlx"
	"github.com/sony/sonyflake"
)

// AuditTrail -
type AuditTrail struct {
	ID         string       `db:"id" json:"id"`
	Operation  string       `db:"operation" json:"operation"`
	ModuleName string       `db:"module_name" json:"module_name"`
	TableName  string       `db:"table_name" json:"table_name"`
	TablePK    string       `db:"table_pk" json:"table_pk"`
	Change     string       `db:"change" json:"change"`
	Remark     string       `db:"remark" json:"remark"`
	ServiceIP  string       `db:"service_ip" json:"service_ip"`
	CreatedBy  int64        `db:"created_by" json:"created_by"`
	CreatedAt  sql.NullTime `db:"created_at" json:"created_at"`
}

var moduleName = "audit_trail"
var table = "audit_trails"
var sf *sonyflake.Sonyflake

func init() {
	loc, _ := time.LoadLocation("UTC")
	sfTime, _ := time.ParseInLocation("2006-01-02 15:04:05", "2020-01-01 00:00:00", loc)

	var st sonyflake.Settings
	st.StartTime = sfTime
	sf = sonyflake.NewSonyflake(st)
}

// GenerateID -
func (at *AuditTrail) GenerateID() {
	id, err := sf.NextID()
	for err != nil {
		id, err = sf.NextID()
	}
	at.ID = strconv.FormatUint(id, 10)
}

// Create -
func (at *AuditTrail) Create(d *sqlx.DB) (string, error) {
	name, err := os.Hostname()
	if err == nil {
		at.ServiceIP = name + ":" + os.Getenv("port")
	}
	at.CreatedAt.Valid = true
	at.CreatedAt.Time = time.Now().UTC()
	at.GenerateID()
	query, val := db.PrepareInsert(table, at, []string{"updated_at", "deleted_at"})
	_, _, err = db.Exec(d, query, val)
	return at.ID, err
}

// List -
func List(d *sqlx.DB, params map[string]interface{}, orderParams map[string]interface{}) ([]AuditTrail, int64, error) {
	list := []AuditTrail{}
	values := map[string]interface{}{}
	condition := " "

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "module_name",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "table_name",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "table_pk",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "operation",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "change",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "remark",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "created_by",
		Condition: "equal",
	}, params, condition, values)

	defaultOrder := map[string]interface{}{
		"field":     "created_at",
		"direction": "desc",
		"start":     int64(0),
		"limit":     int64(10),
	}

	type pagination struct {
		TotalItems int64 `db:"total"`
	}
	p := new(pagination)

	query := "SELECT COUNT(" + table + ".id) AS total FROM " + table + " WHERE 1=1 " + condition
	_, err := db.Get(d, p, query, values)
	if err != nil {
		return list, 0, err
	}

	orderCondition := db.PrepareOrder(orderParams, defaultOrder)
	query = "SELECT * FROM " + table + " WHERE 1=1 " + condition + orderCondition
	err = db.Select(d, &list, query, values)
	return list, p.TotalItems, err
}

// GetByID -
func (at *AuditTrail) GetByID(d *sqlx.DB, id string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE id = :id LIMIT 1"
	values := map[string]interface{}{
		"id": id,
	}
	exist, err := db.Get(d, at, query, values)
	return exist, err
}
