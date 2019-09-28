package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

type Field struct {
	Name 	string
	Type 	string
	Reqired bool
	IsKey	bool
}

type Table struct {
	Name 	string
	Key 	string
	Fields	[]Field
}

type DataBase struct {
	Tables	[]Table
	Names 	[]string
	DB 		*sql.DB
}

type GetParams struct {
	Offset 		int
	Limit 		int
	TableName	string
}

type GetRowParams struct {
	TableName 	string
	Key			interface{}
}

type RP map[string]interface{}


// GET section ------------------------------
func parseGetParams(r *http.Request) (*GetParams, error) {
	params := &GetParams{}

	if tmp := r.FormValue("limit"); tmp == "" {
		params.Limit = 5
	} else {
		l, err := strconv.Atoi(tmp)

		if err != nil {
			params.Limit = 5
		} else {
			params.Limit = l
		}
	}

	if tmp := r.FormValue("offset"); tmp == "" {
		params.Offset = 0
	} else {
		l, err := strconv.Atoi(tmp)

		if err != nil {
			params.Offset = 0
		} else {
			params.Offset = l
		}
	}

	table := strings.Split(r.URL.Path, "/")[1]

	params.TableName = table
	return params, nil
}
// END of GET section -----------------------------------


// ------------ DELETE section

func parseDeleteParams(r *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	parts := strings.Split(r.URL.Path, "/")

	if len(parts) != 3 {
		return nil, fmt.Errorf("wrong delete format")
	}

	params["table"] = parts[1]
	params["key"] = parts[2]

	return params, nil
}

func (dbInstance *DataBase) createDeleteQuery(params map[string]interface{}) (string, error) {
	table, err := dbInstance.getTable(params["table"].(string))

	if err != nil {
		return "", err
	}

	return fmt.Sprintf("DELETE FROM %s WHERE %s = ?", table.Name, table.Key), nil
}

func (dbInstance *DataBase) execDeleteQuery(params map[string]interface{}, query string) (int64, error) {
	res, err := dbInstance.DB.Exec(query, params["key"])

	if err != nil {
		return 0, err
	} else {
		result, _ := res.RowsAffected()
		return result, nil
	}
}


// END of DELETE section -------------


// ------------ POST section
func parsePostParams(r *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&params)
	if err != nil {
		return nil, err
	}
	parts := strings.Split(r.URL.Path, "/")

	params["table"] = parts[1]
	params["key"] = parts[2]

	return params, nil
}

func (dbInstance *DataBase) validatePostParams(params *map[string]interface{}) error {
	table, err := dbInstance.getTable((*params)["table"].(string))

	if err != nil {
		return err
	}

	for key, _ := range *params {
		if key == "table" || key == "key" {
			continue
		}

		if key == table.Key {  // check for primary key update
			return fmt.Errorf("field %s have invalid type", key)
		}

		// check for params that are unknown to current table
		delFlag := true
		for _, field := range table.Fields {
			if field.Name == key {
				delFlag = false
			}
		}
		if delFlag {
			delete(*params, key)
		}
	}

	for key, val := range *params {
		if key == "key" || key == "table" {
			continue
		}
		field := table.getFieldByName(key)

		val, err :=	field.convertValFromStr(val)

		if err != nil {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}

		if val == nil && field.Reqired {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}

		(*params)[field.Name] = val  // update value in params map to converted type
	}
	return nil
}

func (dbInstance *DataBase) createPostQuery(params map[string]interface{}) (string, error) {
	table, err := dbInstance.getTable(params["table"].(string))

	if err != nil {
		return "", err
	}

	values := make([]string, len(params) - 2) // table and primary key

	num := 0
	for i, _ := range params {
		if i == "table" || i == "key" {
			continue
		}
		values[num] = fmt.Sprintf("%s=?", i)
		num++
	}
	return fmt.Sprintf("UPDATE %s SET %s WHERE %s = ?", table.Name,
		strings.Join(values, ","), table.Key), nil
}


func (dbInstance *DataBase) execPostQuery(params map[string]interface{}, query string) (int64, error) {
	_, err := dbInstance.getTable(params["table"].(string))

	if err != nil {
		return 0, err
	}

	vals := make([]interface{}, len(params) - 1) // exclude table field

	i := 0
	for key, val := range params {
		if key == "table" || key == "key"{
			continue
		}

		vals[i] = val
		i++
	}
	vals[i] = params["key"]

	res, err := dbInstance.DB.Exec(query, vals...)
	if err != nil {
		return 0, err
	} else {
		result, _ := res.RowsAffected()
		return result, nil
	}
}


// END of POST section --------------------

func (f *Field) getDefValue() interface{} {
	switch f.Type {
	case "int":
		return 0
	case "text":
		return ""
	case "varchar":
		return ""
	}
	return nil
}

func (f *Field) convertValFromStr(inp interface{}) (interface{}, error) {
	if inp == nil {
		return nil, nil
	}

	switch f.Type {
	case "text":
		val, ok := inp.(string)

		if !ok {
			return nil, fmt.Errorf("Converstion error")
		}
		return val, nil
	case "varchar":
		val, ok := inp.(string)

		if !ok {
			return nil, fmt.Errorf("Converstion error")
		}
		return val, nil
	case "int":
		val, ok := inp.(float64)

		if !ok {
			return nil, fmt.Errorf("Converstion error")
		}
		return int(val), nil
	default:
		return nil, nil
	}
}



// PUT section ------------------------------
func parsePutParams(r *http.Request) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&params)
	if err != nil {
		return nil, err
	}
	params["table"] = strings.Split(r.URL.Path, "/")[1]

	return params, nil
}

func (dbInstance *DataBase) validatePutParams(params *map[string]interface{}) error {
	table, err := dbInstance.getTable((*params)["table"].(string))

	if err != nil {
		return err
	}

	for key, _ := range *params {
		if key == "table" {
			continue
		}

		if key == table.Key {
			delete(*params, key)
			continue
		}

		delFlag := true
		for _, field := range table.Fields {
			if field.Name == key {
				delFlag = false
			}
		}
		if delFlag {
			delete(*params, key)
		}
	}

	for _, field := range table.Fields {
		if (*params)[field.Name] == nil && field.Reqired {
			(*params)[field.Name] = field.getDefValue()
		} else if !field.Reqired {
			(*params)[field.Name] = nil
		}
	}

	for _, field := range table.Fields {
		if field.Name == table.Key {
			continue
		}
		val := (*params)[field.Name]
		val, err :=	field.convertValFromStr(val)

		if err != nil {
			return fmt.Errorf("field %s have invalid type", field.Name)
		}
		(*params)[field.Name] = val
	}
	return nil
}

func (dbInstance *DataBase) createPutQuery(params map[string]interface{}) (string, error) {
	table, err := dbInstance.getTable(params["table"].(string))

	if err != nil {
		return "", err
	}

	values := make([]string, len(table.Fields))
	placeholders := make([]string, len(table.Fields))

	for i, field := range table.Fields {
		values[i] = field.Name
		placeholders[i] = "?"
	}
	return fmt.Sprintf("INSERT INTO %s (%s) values (%s)", params["table"].(string),
		strings.Join(values, ","), strings.Join(placeholders, ",")), nil
}


func (dbInstance *DataBase) execPutQuery(params map[string]interface{}, query string) (int64, error) {
	table, err := dbInstance.getTable(params["table"].(string))

	if err != nil {
		return 0, err
	}

	vals := make([]interface{}, len(params) - 1)

	for i, val := range table.Fields {
		vals[i] = params[val.Name]
	}

	res, err := dbInstance.DB.Exec(query, vals...)
	if err != nil {
		return 0, err
	} else {
		result, _ := res.LastInsertId()
		return result, nil
	}
}


// End of PUT section -------------------------------

func (t * Table) getFieldByName(name string) *Field {
	for _, field := range t.Fields {
		if field.Name == name {
			return &field
		}
	}
	return nil
}

func (t *Table) getRowStruct() []interface{} {
	rowStruct := make([]interface{}, len(t.Fields))

	for idx, val := range t.Fields {
		switch val.Type {
		case "int":
			rowStruct[idx] = new(sql.NullInt64)
		case "varchar":
			rowStruct[idx] = new(sql.NullString)
		case "text":
			rowStruct[idx] = new(sql.NullString)
		}
	}

	return rowStruct
}

func (t *Table) transformRowForResponse(row []interface{}) map[string]interface{} {
	transformedRow := make(map[string]interface{})

	for idx, val := range row {
		switch val.(type) {
		case *sql.NullInt64:
			if value, ok := val.(*sql.NullInt64); ok {
				if value.Valid {
					transformedRow[t.Fields[idx].Name] = value.Int64
				} else {
					transformedRow[t.Fields[idx].Name] = nil
				}
			}
		case *sql.NullString:
			if value, ok := val.(*sql.NullString); ok {
				if value.Valid {
					transformedRow[t.Fields[idx].Name] = value.String
				} else {
					transformedRow[t.Fields[idx].Name] = nil
				}
			}
		}
	}
	return transformedRow
}

func (dbInstance *DataBase) getTable(name string) (*Table, error) {
	for i := 0; i < len(dbInstance.Tables); i++ {
		if dbInstance.Tables[i].Name == name {
			return &dbInstance.Tables[i], nil
		}
	}
	return nil, fmt.Errorf("unknown table")
}

func (dbInstance *DataBase) getRow(params GetRowParams) (map[string]interface{}, error) {

	table, err := dbInstance.getTable(params.TableName)

	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s WHERE %s = ?", params.TableName, table.Key)

	res := dbInstance.DB.QueryRow(query, params.Key)

	row := table.getRowStruct()

	err = res.Scan(row...)

	if err != nil {
		return nil, fmt.Errorf("record not found")
	}
	return table.transformRowForResponse(row), nil
}

func (dbInstance *DataBase) getRows(params GetParams) ([]map[string]interface{}, error) {
	var out []map[string]interface{}

	table, err := dbInstance.getTable(params.TableName)

	if err != nil {
		return nil, err
	}

	query := fmt.Sprintf("SELECT * FROM %s LIMIT ? OFFSET ?", params.TableName)
	rows, err := dbInstance.DB.Query(query, params.Limit, params.Offset)

	if err != nil {
		return nil, err
	}


	defer rows.Close()

	for rows.Next() {
		row := table.getRowStruct()

		rows.Scan(row...)

		//fmt.Println(row)
		transformed := table.transformRowForResponse(row)

		out = append(out, transformed)
	}
	return out, nil
}

func InitDB(db *sql.DB) (*DataBase, error) {
	dbInstance := &DataBase{}
	dbInstance.DB = db

	dbInstance.DB.SetMaxIdleConns(1)
	res, err := dbInstance.DB.Query("SHOW TABLES")

	if err != nil {
		return nil, err
	}
	defer res.Close()

	var table string

	for res.Next() {
		res.Scan(&table)
		rows, err := dbInstance.DB.Query("SELECT column_name, if (column_key='PRI', true, false) as 'key', DATA_TYPE, if(is_nullable='NO', true, false) as nullable from information_schema.columns where  table_name = ? and table_schema=database()", table)
		if err != nil {
			return nil, err
		}

		var fields []Field
		var key string
		for rows.Next() {
			var f Field
			//fmt.Println(resColumns)
			rows.Scan(&f.Name, &f.IsKey, &f.Type, &f.Reqired)
			if f.IsKey {
				key = f.Name
			}
			fields = append(fields, f)

		}
		err = rows.Err()

		if err != nil {
			fmt.Println(err)
		}
		rows.Close()

		dbInstance.Tables = append(dbInstance.Tables, Table{
			Name:   table,
			Key:    key,
			Fields: fields,
		})
		dbInstance.Names = append(dbInstance.Names, table)

		if err != nil {
			fmt.Println(err)
		}
	}
	return dbInstance, nil
}


func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	dataBase, err := InitDB(db)

	if err != nil {
		return nil, err
	}
	serverMux := http.NewServeMux()

	serverMux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:

			if r.URL.Path == "/" {
				res, err := json.Marshal(RP{"response": RP{"tables": dataBase.Names}})
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(res)
				return
			}
			parts := strings.Split(r.URL.Path, "/")

			switch len(parts) {
			case 2:
				params, err := parseGetParams(r)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				rows, err := dataBase.getRows(*params)

				if err != nil {
					js, _ := json.Marshal(RP{"error" : err.Error()})
					w.WriteHeader(http.StatusNotFound)
					w.Write(js)
					return
				}

				js, err := json.Marshal(RP{"response": RP{"records" :rows}})

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}

				w.Write(js)
			case 3:
				params := GetRowParams{
					TableName: parts[1],
					Key:       parts[2],
				}
				row, err := dataBase.getRow(params)

				if err != nil {
					js, _ := json.Marshal(RP{"error" : err.Error()})
					w.WriteHeader(http.StatusNotFound)
					w.Write(js)
					return
				}

				js, err := json.Marshal(RP{"response": RP{"record" : row}})

				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					return
				}
				w.Write(js)
			}


		case http.MethodPut:
			params, err := parsePutParams(r)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}

			err = dataBase.validatePutParams(&params)

			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			query, err := dataBase.createPutQuery(params)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			i, err := dataBase.execPutQuery(params, query)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			table, _ := dataBase.getTable(params["table"].(string))

			js, _ := json.Marshal(RP{"response": RP{table.Key : i}})
			w.WriteHeader(http.StatusOK)
			w.Write(js)
		case http.MethodPost:

			params, err := parsePostParams(r)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			err = dataBase.validatePostParams(&params)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusBadRequest)
				w.Write(js)
				return
			}

			query, err := dataBase.createPostQuery(params)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			i, err := dataBase.execPostQuery(params, query)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			js, _ := json.Marshal(RP{"response": RP{"updated" : i}})
			w.WriteHeader(http.StatusOK)
			w.Write(js)

		case http.MethodDelete:

			params, err := parseDeleteParams(r)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}

			query, err := dataBase.createDeleteQuery(params)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}

			i, err := dataBase.execDeleteQuery(params, query)

			if err != nil {
				js, _ := json.Marshal(RP{"error" : err.Error()})
				w.WriteHeader(http.StatusNotFound)
				w.Write(js)
				return
			}
			js, _ := json.Marshal(RP{"response": RP{"deleted" : i}})
			w.WriteHeader(http.StatusOK)
			w.Write(js)
		}
	})
	return serverMux, nil
}

