package db

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/helloferdie/stdgo/libslice"
	"github.com/helloferdie/stdgo/logger"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
)

// ConnectionString -
func ConnectionString() string {
	conn := "default:default@(127.0.0.1:3306)/golang?timeout=10s&charset=utf8mb4&parseTime=true"
	dbHost := os.Getenv("db_host")
	if dbHost != "" {
		host := os.Getenv("db_host")
		user := os.Getenv("db_user")
		pass := os.Getenv("db_pass")
		port := os.Getenv("db_port")
		dbName := os.Getenv("db_name")
		conn = user + ":" + pass + "@(" + host + ":" + port + ")/" + dbName + "?timeout=10s&charset=utf8mb4&parseTime=true"
	}
	return conn
}

// CustomConnectionString -
func CustomConnectionString(config map[string]string) string {
	host := config["db_host"]
	user := config["db_user"]
	pass := config["db_pass"]
	port := config["db_port"]
	dbName := config["db_name"]
	conn := user + ":" + pass + "@(" + host + ":" + port + ")/" + dbName + "?timeout=10s&charset=utf8&parseTime=true"
	return conn
}

// Open -
func Open(conn string) (*sqlx.DB, error) {
	return OpenRetry(conn, 3)
}

// OpenRetry -
func OpenRetry(conn string, maxRetry int) (*sqlx.DB, error) {
	if conn == "" {
		conn = ConnectionString()
	}
	if maxRetry <= 0 {
		maxRetry = 3
	}
	for retry := 0; retry < maxRetry; retry++ {
		db, err := sqlx.Connect("mysql", conn)
		if err != nil {
			if retry == 2 {
				// Stop retry after 3 times
				logger.PrintLogEntry("error", fmt.Sprintf("Error establish database connection after retry 3 times %v", err), true)
				return db, err
			}

			errType := reflect.TypeOf(err).String()
			/*	if err == mysql.ErrPktSyncMul {
					// Sleep and wait for connection availability
					time.Sleep(time.Second * 3)
					continue
				} else {
					_, exist := libslice.Contains(errType, []string{"*net.OpError"})
					if exist {
						fmt.Println("im here")
						time.Sleep(time.Second * 3)
						continue
					}
				}*/
			logger.PrintLogEntry("error", fmt.Sprintf("Error establish database connection %v type: %s", err, errType), true)

			time.Sleep(time.Second * 3)
			continue
		}
		return db, nil
	}
	return nil, nil
}

// Exec -
func Exec(db *sqlx.DB, query string, values map[string]interface{}) (int64, int64, error) {
	result, err := db.NamedExec(query, values)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error execute query %v", err)
		return 0, 0, err
	}
	id, err := result.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return 0, 0, err
	}
	return id, rows, nil
}

// ExecList -
func ExecList(db *sqlx.DB, list []interface{}) error {
	tx, err := db.Beginx()
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error begin query transactions %v", err)
		return err
	}

	for k, data := range list {
		d, ok := data.(map[string]interface{})
		if !ok {
			logger.MakeLogEntry(nil, true).Errorf("Failed to convert interface{} for statement %v", strconv.Itoa(k))
			return fmt.Errorf("%s", "general.error_query_transaction")
		}

		q, qExist := d["query"].(string)
		if !qExist {
			logger.MakeLogEntry(nil, true).Errorf("Failed to find query for statement %v", strconv.Itoa(k))
			return fmt.Errorf("%s", "general.error_query_transaction")
		}

		v, vExist := d["values"].(map[string]interface{})
		if !vExist {
			logger.MakeLogEntry(nil, true).Errorf("Failed to find argument values for statement %v", strconv.Itoa(k))
			return fmt.Errorf("%s", "general.error_query_transaction")
		}

		_, err = tx.NamedExec(q, v)
		if err != nil {
			logger.MakeLogEntry(nil, true).Errorf("Failed to find execute statement %v", strconv.Itoa(k))
			break
		}
	}

	if err == nil {
		tx.Commit()
	} else {
		logger.MakeLogEntry(nil, true).Errorf("Error execute list query transactions, operation has been rollback %v", err)
		tx.Rollback()
	}
	return nil
}

// Get -
func Get(db *sqlx.DB, list interface{}, query string, values map[string]interface{}) (bool, error) {
	exist := false
	rows, err := db.NamedQuery(query, values)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error get query %v", err)
		return exist, err
	}
	defer rows.Close()

	if rows.Next() {
		err = rows.StructScan(list)
		if err != nil {
			logger.MakeLogEntry(nil, true).Errorf("Error scan row %v", err)
			return exist, err
		}
		exist = true
	}
	rows.Close()
	return exist, nil
}

// Select -
func Select(db *sqlx.DB, list interface{}, query string, values map[string]interface{}) error {
	nstmt, err := db.PrepareNamed(query)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error select prepare named query %v", err)
		return err
	}
	err = nstmt.Select(list, values)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error select query %v", err)
		return err
	}
	return nil
}

// PrepareInsert -
func PrepareInsert(table string, data interface{}, skip []string) (string, map[string]interface{}) {
	rVal := reflect.ValueOf(data)
	if rVal.Kind() == reflect.Ptr {
		rVal = rVal.Elem()
	}
	rType := rVal.Type()

	col := ""
	val := ""

	isCustom := true
	if len(skip) == 0 {
		skip = []string{"id", "created_at", "updated_at", "deleted_at"}
		isCustom = false
	}

	v := map[string]interface{}{}
	for i := 0; i < rVal.NumField(); i++ {
		tag := rType.Field(i).Tag.Get("db")
		if tag == "" {
			continue
		}
		_, exist := libslice.Contains(tag, skip)
		if exist {
			continue
		}

		if col != "" {
			col += ", "
			val += ", "
		}
		col += "`" + tag + "`"
		val += ":" + tag

		v[tag] = rVal.Field(i).Interface()
	}
	if !isCustom {
		col += ", `created_at`, `updated_at`"
		val += ", :created_at, :updated_at"

		v["created_at"] = time.Now().UTC()
		v["updated_at"] = time.Now().UTC()
	}

	query := "INSERT INTO " + table + " (" + col + ") VALUES (" + val + ")"
	return query, v
}

// PrepareInsertOnly -
func PrepareInsertOnly(table string, data interface{}, only []string) (string, map[string]interface{}) {
	rVal := reflect.ValueOf(data)
	if rVal.Kind() == reflect.Ptr {
		rVal = rVal.Elem()
	}
	rType := rVal.Type()

	col := ""
	val := ""

	v := map[string]interface{}{}
	for i := 0; i < rVal.NumField(); i++ {
		tag := rType.Field(i).Tag.Get("db")
		if tag == "" {
			continue
		}
		_, exist := libslice.Contains(tag, only)
		if !exist {
			continue
		}

		if col != "" {
			col += ", "
			val += ", "
		}
		col += "`" + tag + "`"
		val += ":" + tag

		v[tag] = rVal.Field(i).Interface()
	}

	query := "INSERT INTO " + table + " (" + col + ") VALUES (" + val + ")"
	return query, v
}

// PrepareUpdate -
func PrepareUpdate(table string, old interface{}, new interface{}, skip []string, condition string, conditionVal map[string]interface{}) (string, map[string]interface{}, map[string]interface{}) {
	rVal := reflect.ValueOf(old)
	if rVal.Kind() == reflect.Ptr {
		rVal = rVal.Elem()
	}
	rType := rVal.Type()

	nVal := reflect.ValueOf(new)
	if nVal.Kind() == reflect.Ptr {
		nVal = nVal.Elem()
	}

	col := ""
	isCustom := true
	if len(skip) == 0 {
		skip = []string{"id", "created_at", "updated_at"}
		isCustom = false
	}

	v := map[string]interface{}{}
	diff := map[string]interface{}{}
	for i := 0; i < rVal.NumField(); i++ {
		tag := rType.Field(i).Tag.Get("db")
		if tag == "" {
			continue
		}
		_, exist := libslice.Contains(tag, skip)
		if exist {
			continue
		}
		if rVal.Field(i).Interface() == nVal.Field(i).Interface() {
			continue
		}

		if col != "" {
			col += ", "
		}
		col += "`" + tag + "` = :" + tag
		v[tag] = nVal.Field(i).Interface()
		diff[tag] = map[string]interface{}{
			"o": rVal.Field(i).Interface(),
			"n": nVal.Field(i).Interface(),
		}
	}

	if !isCustom {
		col += ", updated_at = :updated_at"
		v["updated_at"] = time.Now().UTC()
	}
	if condition == "" {
		condition += "AND id = :id "
	}
	for k, c := range conditionVal {
		v[k] = c
	}
	query := "UPDATE " + table + " SET " + col + " WHERE 1=1 " + condition
	return query, v, diff
}

// PrepareDelete -
func PrepareDelete(table string, pk interface{}, softDelete bool) (string, map[string]interface{}) {
	v := map[string]interface{}{
		"id": pk,
	}
	query := "DELETE FROM " + table + " WHERE id = :id"
	if softDelete {
		query = "UPDATE " + table + " SET updated_at = :updated_at, deleted_at = :deleted_at WHERE id = :id"
		v["updated_at"] = time.Now().UTC()
		v["deleted_at"] = time.Now().UTC()
	}
	return query, v
}

// PrepareOrder -
func PrepareOrder(params map[string]interface{}, def map[string]interface{}) string {
	query := ""
	orderVal, orderExist := params["field"].(string)
	defVal, _ := def["field"].(string)
	customOrder := false
	if orderExist && orderVal != "" {
		query += "ORDER BY `" + orderVal + "` "
	} else {
		defCustomVal, _ := def["custom"].(string)
		if defCustomVal == "" {
			query += "ORDER BY `" + defVal + "` "
		} else {
			query += defCustomVal + " "
			customOrder = true
		}
	}

	if !customOrder {
		orderVal, orderExist = params["direction"].(string)
		defVal, _ = def["direction"].(string)
		if orderExist && orderVal != "" {
			query += orderVal + " "
		} else {
			query += defVal + " "
		}
	}

	showVal, showExist := params["show"].(bool)
	if !showExist || !showVal {
		query += "LIMIT "
		startVal, startExist := params["start"].(int64)
		defStartVal, _ := def["start"].(int64)
		if startExist {
			query += strconv.FormatInt(startVal, 10) + ", "
		} else {
			query += strconv.FormatInt(defStartVal, 10) + ", "
		}

		limitVal, limitExist := params["limit"].(int64)
		defLimitVal, _ := def["limit"].(int64)
		if limitExist {
			query += strconv.FormatInt(limitVal, 10) + " "
		} else {
			query += strconv.FormatInt(defLimitVal, 10) + " "
		}
	}
	return query
}

// CheckTableExists -
func CheckTableExists(db *sqlx.DB, dbName string, table string) (bool, error) {
	exist := false

	if dbName == "" {
		dbName = os.Getenv("db_name")
	}

	values := map[string]interface{}{
		"table": table,
		"db":    dbName,
	}

	type pagination struct {
		TotalItems int64 `db:"total"`
	}
	p := new(pagination)

	query := "SELECT COUNT(*) AS total FROM information_schema.tables WHERE table_schema = :db AND table_name = :table LIMIT 1;"
	_, err := Get(db, p, query, values)
	if err == nil && p.TotalItems == 1 {
		exist = true
	}
	return exist, err
}

// Query -
func Query(db *sqlx.DB, query string, args []interface{}) (*sqlx.Rows, error) {
	rows, err := db.Queryx(query, args...)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error select query %v", err)
		return nil, err
	}
	return rows, nil
}

// InsertMultiple -
func InsertMultiple(db *sqlx.DB, table string, data interface{}, value []interface{}, skip []string) (int64, int64, error) {
	rVal := reflect.ValueOf(data)
	if rVal.Kind() == reflect.Ptr {
		rVal = rVal.Elem()
	}
	rType := rVal.Type()

	col := ""

	if len(skip) == 0 {
		skip = []string{"id", "created_at", "updated_at", "deleted_at"}
	}

	for i := 0; i < rVal.NumField(); i++ {
		tag := rType.Field(i).Tag.Get("db")
		if tag == "" {
			continue
		}
		_, exist := libslice.Contains(tag, skip)
		if exist {
			continue
		}

		if col != "" {
			col += ", "
		}
		col += "`" + tag + "`"
	}

	var vals []interface{}

	query := "INSERT INTO " + table + " (" + col + ") VALUES "
	cols := strings.Split(col, ",")
	for _, row := range value {
		query += `(?` + strings.Repeat(",?", len(cols)-1) + `),`

		v := row.([]interface{})
		for i := 0; i < len(v); i++ {
			vals = append(vals, v[i])
		}
	}

	//trim the last ,
	query = query[0 : len(query)-1]

	//prepare the statement
	stmt, err := db.Prepare(query)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error prepare query %v", err)
		return 0, 0, err
	}

	//format all vals at once
	res, err := stmt.Exec(vals...)
	if err != nil {
		logger.MakeLogEntry(nil, true).Errorf("Error exec query %v", err)
		return 0, 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, 0, err
	}
	rows, err := res.RowsAffected()
	if err != nil {
		return 0, 0, err
	}

	return id, rows, err
}
