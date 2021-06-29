package language

import (
	"database/sql"

	"strconv"

	"github.com/helloferdie/stdgo/audittrail/event"
	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/libquery"
	"github.com/helloferdie/stdgo/libstring"
	"github.com/jmoiron/sqlx"
)

// Language -
type Language struct {
	ID         int64        `db:"id" json:"id"`
	Label      string       `db:"label" json:"label"`
	LabelShort string       `db:"label_short" json:"label_short"`
	CreatedAt  sql.NullTime `db:"created_at" json:"created_at"`
	UpdatedAt  sql.NullTime `db:"updated_at" json:"updated_at"`
	DeletedAt  sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

var moduleName = "language"
var table = "languages"

// Create -
func (la *Language) Create(d *sqlx.DB, creatorID int64) (int64, error) {
	query, val := db.PrepareInsert(table, la, []string{})
	id, _, err := db.Exec(d, query, val)
	if err == nil {
		la.GetByID(d, id)
		go event.CreateAuditTrail(map[string]interface{}{
			"operation":   "add",
			"module_name": moduleName,
			"table_name":  table,
			"table_pk":    strconv.FormatInt(la.ID, 10),
			"change":      libstring.JSONEncode(la),
			"remark":      "",
			"created_by":  creatorID,
		})
	}
	return id, err
}

// Save -
func (la *Language) Save(d *sqlx.DB, creatorID int64) error {
	old := new(Language)
	old.GetByID(d, la.ID)

	query, val, diff := db.PrepareUpdate(table, old, la, []string{}, "", map[string]interface{}{"id": la.ID})
	if len(diff) > 0 {
		_, _, err := db.Exec(d, query, val)
		if err == nil {
			go event.CreateAuditTrail(map[string]interface{}{
				"operation":   "edit",
				"module_name": moduleName,
				"table_name":  table,
				"table_pk":    strconv.FormatInt(la.ID, 10),
				"change":      libstring.JSONEncode(diff),
				"remark":      "",
				"created_by":  creatorID,
			})
		}
		return err
	}
	return nil
}

// Delete -
func (la *Language) Delete(d *sqlx.DB, creatorID int64, softDelete bool) error {
	query, val := db.PrepareDelete(table, la.ID, softDelete)
	_, _, err := db.Exec(d, query, val)
	if err == nil {
		go func() {
			remark := ""
			if !softDelete {
				remark = "permanent delete"
			}
			event.CreateAuditTrail(map[string]interface{}{
				"operation":   "edit",
				"module_name": moduleName,
				"table_name":  table,
				"table_pk":    strconv.FormatInt(la.ID, 10),
				"change":      libstring.JSONEncode(la),
				"remark":      remark,
				"created_by":  creatorID,
			})
		}()
	}
	return err
}

// List -
func List(d *sqlx.DB, params map[string]interface{}, orderParams map[string]interface{}) ([]Language, int64, error) {
	list := []Language{}
	values := map[string]interface{}{}
	condition := "AND deleted_at IS NULL "

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "label",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "label_short",
		Condition: "like",
	}, params, condition, values)

	defaultOrder := map[string]interface{}{
		"field":     "label",
		"direction": "asc",
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
func (la *Language) GetByID(d *sqlx.DB, id int64) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE id = :id AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"id": id,
	}
	exist, err := db.Get(d, la, query, values)
	return exist, err
}

// GetByLabel -
func (la *Language) GetByLabel(d *sqlx.DB, label string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE label LIKE :label AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"label": label,
	}
	exist, err := db.Get(d, la, query, values)
	return exist, err
}

// GetByLabelShort -
func (la *Language) GetByLabelShort(d *sqlx.DB, ls string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE label_short LIKE :label_short AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"label_short": ls,
	}
	exist, err := db.Get(d, la, query, values)
	return exist, err
}

// MassCheckID -
func MassCheckID(d *sqlx.DB, list []int64) (bool, error) {
	type pagination struct {
		TotalItems int `db:"total"`
	}
	p := new(pagination)

	values := map[string]interface{}{}
	condition := "AND deleted_at IS NULL "

	params := map[string]interface{}{"id": list}
	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "id",
		Condition: "in",
	}, params, condition, values)

	query := "SELECT COUNT(" + table + ".id) AS total FROM " + table + " WHERE 1=1 " + condition
	_, err := db.Get(d, p, query, values)
	if err != nil {
		return false, err
	}

	if p.TotalItems == len(list) {
		return true, nil
	}
	return false, nil
}
