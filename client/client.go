package client

import (
	"database/sql"
	"strconv"

	"github.com/helloferdie/stdgo/audittrail/event"
	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/libquery"
	"github.com/helloferdie/stdgo/libstring"
	"github.com/jmoiron/sqlx"
)

// Client -
type Client struct {
	ID           int64        `db:"id" json:"id"`
	UUID         string       `db:"uuid" json:"uuid"`
	ClientName   string       `db:"client_name" json:"client_name"`
	ClientSecret string       `db:"client_secret" json:"client_secret"`
	IsActive     bool         `db:"is_active" json:"is_active"`
	CreatedAt    sql.NullTime `db:"created_at" json:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at" json:"updated_at"`
	DeletedAt    sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

var moduleName = "client"
var table = "clients"

// Create -
func (cl *Client) Create(d *sqlx.DB, creatorID int64) (int64, error) {
	query, val := db.PrepareInsert(table, cl, []string{})
	id, _, err := db.Exec(d, query, val)
	if err == nil {
		cl.GetByID(d, id)
		go event.CreateAuditTrail(map[string]interface{}{
			"operation":   "add",
			"module_name": moduleName,
			"table_name":  table,
			"table_pk":    strconv.FormatInt(cl.ID, 10),
			"change":      libstring.JSONEncode(cl),
			"remark":      "",
			"created_by":  creatorID,
		})
	}
	return id, err
}

// Save -
func (cl *Client) Save(d *sqlx.DB, creatorID int64) error {
	old := new(Client)
	old.GetByID(d, cl.ID)

	query, val, diff := db.PrepareUpdate(table, old, cl, []string{}, "", map[string]interface{}{"id": cl.ID})
	if len(diff) > 0 {
		_, _, err := db.Exec(d, query, val)
		if err == nil {
			go event.CreateAuditTrail(map[string]interface{}{
				"operation":   "edit",
				"module_name": moduleName,
				"table_name":  table,
				"table_pk":    strconv.FormatInt(cl.ID, 10),
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
func (cl *Client) Delete(d *sqlx.DB, creatorID int64, softDelete bool) error {
	query, val := db.PrepareDelete(table, cl.ID, softDelete)
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
				"table_pk":    strconv.FormatInt(cl.ID, 10),
				"change":      libstring.JSONEncode(cl),
				"remark":      remark,
				"created_by":  creatorID,
			})
		}()
	}
	return err
}

// List -
func List(d *sqlx.DB, params map[string]interface{}, orderParams map[string]interface{}) ([]Client, int64, error) {
	list := []Client{}
	values := map[string]interface{}{}
	condition := "AND deleted_at IS NULL "

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "client_name",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "client_secret",
		Condition: "like",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "uuid",
		Condition: "like",
	}, params, condition, values)

	defaultOrder := map[string]interface{}{
		"field":     "uuid",
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
func (cl *Client) GetByID(d *sqlx.DB, id int64) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE id = :id AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"id": id,
	}
	exist, err := db.Get(d, cl, query, values)
	return exist, err
}

// GetByUUID -
func (cl *Client) GetByUUID(d *sqlx.DB, uuid string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE uuid = :uuid AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"uuid": uuid,
	}
	exist, err := db.Get(d, cl, query, values)
	return exist, err
}

// GetByClientName -
func (cl *Client) GetByClientName(d *sqlx.DB, name string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE client_name LIKE :client_name AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"client_name": name,
	}
	exist, err := db.Get(d, cl, query, values)
	return exist, err
}
