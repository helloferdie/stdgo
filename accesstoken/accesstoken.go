package accesstoken

import (
	"database/sql"
	"time"

	"github.com/helloferdie/stdgo/audittrail/event"
	"github.com/helloferdie/stdgo/db"
	"github.com/helloferdie/stdgo/libquery"
	"github.com/helloferdie/stdgo/libstring"
	"github.com/jmoiron/sqlx"
)

// AccessToken -
type AccessToken struct {
	ID           string       `db:"id" json:"id"`
	UserID       int64        `db:"user_id" json:"user_id"`
	AccountID    int64        `db:"account_id" json:"account_id"`
	ClientID     int64        `db:"client_id" json:"client_id"`
	RefreshToken string       `db:"refresh_token" json:"refresh_token"`
	DeviceToken  string       `db:"device_token" json:"device_token"`
	LastLoginIP  string       `db:"last_login_ip" json:"last_login_ip"`
	IsRevoke     bool         `db:"is_revoke" json:"is_revoke"`
	CreatedAt    sql.NullTime `db:"created_at" json:"created_at"`
	UpdatedAt    sql.NullTime `db:"updated_at" json:"updated_at"`
	DeletedAt    sql.NullTime `db:"deleted_at" json:"deleted_at"`
}

var moduleName = "access_token"
var table = "access_tokens"

// Create -
func (at *AccessToken) Create(d *sqlx.DB, creatorID int64) (string, error) {
	at.CreatedAt.Valid = true
	at.CreatedAt.Time = time.Now().UTC()
	at.UpdatedAt.Valid = true
	at.UpdatedAt.Time = time.Now().UTC()

	query, val := db.PrepareInsert(table, at, []string{"-"})
	_, _, err := db.Exec(d, query, val)
	if err == nil {
		at.GetByID(d, at.ID)
		go event.CreateAuditTrail(map[string]interface{}{
			"operation":   "add",
			"module_name": moduleName,
			"table_name":  table,
			"table_pk":    at.ID,
			"change":      libstring.JSONEncode(at),
			"remark":      "",
			"created_by":  creatorID,
		})
	}
	return at.ID, err
}

// Save -
func (at *AccessToken) Save(d *sqlx.DB, creatorID int64) error {
	old := new(AccessToken)
	old.GetByID(d, at.ID)

	query, val, diff := db.PrepareUpdate(table, old, at, []string{}, "", map[string]interface{}{"id": at.ID})
	if len(diff) > 0 {
		_, _, err := db.Exec(d, query, val)
		if err == nil {
			go event.CreateAuditTrail(map[string]interface{}{
				"operation":   "edit",
				"module_name": moduleName,
				"table_name":  table,
				"table_pk":    at.ID,
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
func (at *AccessToken) Delete(d *sqlx.DB, creatorID int64, softDelete bool) error {
	query, val := db.PrepareDelete(table, at.ID, softDelete)
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
				"table_pk":    at.ID,
				"change":      libstring.JSONEncode(at),
				"remark":      remark,
				"created_by":  creatorID,
			})
		}()
	}
	return err
}

// GetByID -
func (at *AccessToken) GetByID(d *sqlx.DB, id string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE id = :id AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"id": id,
	}
	exist, err := db.Get(d, at, query, values)
	return exist, err
}

// GetByRefreshToken -
func (at *AccessToken) GetByRefreshToken(d *sqlx.DB, token string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE refresh_token = :refresh_token AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"refresh_token": token,
	}
	exist, err := db.Get(d, at, query, values)
	return exist, err
}

// GetByDeviceToken -
func (at *AccessToken) GetByDeviceToken(d *sqlx.DB, token string) (bool, error) {
	query := "SELECT * FROM " + table + " WHERE device_token = :token AND deleted_at IS NULL LIMIT 1"
	values := map[string]interface{}{
		"device_token": token,
	}
	exist, err := db.Get(d, at, query, values)
	return exist, err
}

// List -
func List(d *sqlx.DB, params map[string]interface{}, orderParams map[string]interface{}) ([]AccessToken, int64, error) {
	list := []AccessToken{}
	values := map[string]interface{}{}
	condition := "AND deleted_at IS NULL "

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "account_id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "client_id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "user_id",
		Condition: "equal",
	}, params, condition, values)

	condition, values, _ = libquery.QueryCondition(libquery.Config{
		Param:     "is_revoke",
		Condition: "equal",
	}, params, condition, values)

	defaultOrder := map[string]interface{}{
		"field":     "created_at",
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

// ListActiveDeviceToken -
func ListActiveDeviceToken(d *sqlx.DB, accountID int64, limit int64) ([]string, error) {
	type result struct {
		DeviceToken string `db:"device_token"`
	}
	list := []result{}
	values := map[string]interface{}{
		"account_id": accountID,
		"limit":      limit,
	}
	query := "SELECT DISTINCT(device_token) AS device_token FROM " + table + " WHERE 1=1 AND deleted_at IS NULL AND account_id = :account_id AND is_revoke = 0 AND device_token != '' AND client_id IN (1, 2) ORDER BY updated_at DESC LIMIT :limit"
	err := db.Select(d, &list, query, values)
	tmp := []string{}
	if err == nil {
		for _, v := range list {
			tmp = append(tmp, v.DeviceToken)
		}
	}
	return tmp, err
}

// CheckDeviceTokenUniqueUser -
func CheckDeviceTokenUniqueUser(d *sqlx.DB, deviceToken string, accountID int64) (int64, error) {
	values := map[string]interface{}{
		"account_id":   accountID,
		"device_token": deviceToken,
	}

	type pagination struct {
		TotalItems int64 `db:"total"`
	}
	p := new(pagination)

	query := "SELECT COUNT(" + table + ".id) AS total FROM " + table + " WHERE deleted_at IS NULL AND is_revoke = 0 AND device_token = :device_token AND account_id != :account_id "
	_, err := db.Get(d, p, query, values)
	if err != nil {
		return 0, err
	}
	return p.TotalItems, err
}
