package main

import (
	"database/sql"
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/cheggaaa/pb"
)

var from = ConnectFrom()

var to = ConnectTo()

func main() {
	diffs := GetDiffInCols()

	tables := listTables(from)

	count := len(tables)

	for i := 0; i < count; i++ {
		if tables[i] == "migrations" || tables[i] == "permissions" || tables[i] == "model_has_permissions" || tables[i] == "model_has_roles" || tables[i] == "role_has_permissions" || tables[i] == "roles" {
			continue
		}

		columns := getTableColumns(ConnectFrom(), tables[i])

		if columns[0] == "id" {
			doseTableExistsInDiffrentTables, _ := InArray(tables[i], keys(diffs))

			if doseTableExistsInDiffrentTables {
				rows, _ := from.Query(fmt.Sprintf("SELECT * FROM %s", tables[i]))

				PrepareRowsToBeInsertedIntoOtherConnectionsTables(from, to, rows, columns, tables[i])
			}
		}
	}
}

// PrepareRowsToBeInsertedIntoOtherConnectionsTables This Function Is To Generate And Insert Queries
// It Into Other Connections Table.
func PrepareRowsToBeInsertedIntoOtherConnectionsTables(fromConnection *sql.DB, connection *sql.DB, rows *sql.Rows, columns []string, tblName string) {
	var count int

	from.QueryRow("SELECT COUNT(*) FROM " + tblName).Scan(&count)

	// create and start new bar
	bar := pb.StartNew(count)

	for rows.Next() {

		args := make([]interface{}, len(columns))

		sql := "INSERT INTO `" + (tblName) + "`(`" + (strings.Join(columns, "`, `")) + "`) "

		for i := range columns {
			var v interface{}
			args[i] = &v
		}

		if err := rows.Scan(args...); err != nil {
			log.Fatal(err)
		}

		names := []string{}

		for ii, val := range args {

			t := reflect.TypeOf((*(val.(*interface{}))))

			names = append(names, "?")

			if nil == t {
				args[ii] = nil
				continue
			}

			switch t.Kind() {
			case reflect.Int64, reflect.Int32, reflect.Int:
				args[ii] = (*(val.(*interface{}))).(int64)
			case reflect.Float32, reflect.Float64:
				args[ii] = (*(val.(*interface{}))).(float64)
			case reflect.Slice:
				if t.Elem().Kind() == reflect.Uint || t.Elem().Kind() == reflect.Uint8 {
					args[ii] = string((*(val.(*interface{}))).([]uint8))
				}
			default:
				args[ii] = (*(val.(*interface{}))).(string)
			}
		}

		sql += " VALUES(" + (strings.Join(names, ",")) + ")"

		if _, err := connection.Exec(sql, args...); err != nil {
			log.Println(err.Error(), strings.Join(columns, "`, `"))
		}

		bar.Increment()
	}

	bar.Finish()
}

// listTables List The Database Tables
// @param connection *sql.DB
// @return []string
func listTables(connection *sql.DB) []string {
	listedTables, err := connection.Query("SHOW TABLES")

	if err != nil {
		panic(err)
	}

	var table string

	var tables []string

	for listedTables.Next() {
		listedTables.Scan(&table)
		tables = append(tables, table)
	}

	return tables
}

// GetDiffInCols Get The Diffrence Between Databases Tables Columns New & The OLD One.
// @return map[string]interface{}
func GetDiffInCols() map[string]interface{} {
	fromTables := listTables(from)

	toTables := listTables(to)

	var diffs = make(map[string]interface{})

	for index := 0; index < len(toTables); index++ {
		splittedTable := strings.Split(toTables[index], "_")

		if len(splittedTable) > 0 && splittedTable[len(splittedTable)-1] == "view" {
			continue
		}

		var result []interface{}

		doseTableExistsInFromTables, _ := InArray(toTables[index], fromTables)

		toTableColumns := getTableColumns(to, toTables[index])

		var fromTableColumns []string

		if doseTableExistsInFromTables {
			fromTableColumns = getTableColumns(from, toTables[index])
		}

		result = append(result, difference(toTableColumns, fromTableColumns)) // diffColumnsAdd
		result = append(result, difference(fromTableColumns, toTableColumns)) // diffColumnsUnset
		result = append(result, toTableColumns)                               // to_columns
		result = append(result, fromTableColumns)                             // from_columns

		if doseTableExistsInFromTables {
			diffs[toTables[index]] = result
		}
	}

	return diffs
}

// InArray Checks If Item In Array Or Not
// Similar to "in_array" in php.
// @param val interface{}
// @param array interface{}
// @return (exists bool, index int)
func InArray(val interface{}, array interface{}) (exists bool, index int) {
	exists = false
	index = -1

	switch reflect.TypeOf(array).Kind() {
	case reflect.Slice:
		s := reflect.ValueOf(array)

		for i := 0; i < s.Len(); i++ {
			if reflect.DeepEqual(val, s.Index(i).Interface()) == true {
				index = i
				exists = true
				return
			}
		}
	}

	return
}

// Get The Table Names Of Database.
// @param connection *sql.DB
// @param table string
// @return []string
func getTableColumns(connection *sql.DB, table string) []string {
	rows, err := connection.Query(fmt.Sprintf("SELECT * FROM %s", table))

	if err != nil {
		panic(err)
	}

	tableColumns, err := rows.Columns()

	if err != nil {
		panic(err)
	}

	return tableColumns
}

// Get The Diffrence Between Two Sclices.
// Like "array_diff" in php.
// @param slice1 []string
// @param slice2 []string
// @return []string
func difference(slice1 []string, slice2 []string) []string {
	var diff []string

	// Loop two times, first to find slice1 strings not in slice2,
	// second loop to find slice2 strings not in slice1
	for i := 0; i < 2; i++ {
		for _, s1 := range slice1 {
			found := false
			for _, s2 := range slice2 {
				if s1 == s2 {
					found = true
					break
				}
			}
			// String not found. We add it to return slice
			if !found {
				diff = append(diff, s1)
			}
		}
		// Swap the slices, only if it was the first loop
		if i == 0 {
			slice1, slice2 = slice2, slice1
		}
	}

	return diff
}

// Get The Keys Of a map[string]interface{}
// @param elements map[string]interface{}
// @return []string
func keys(elements map[string]interface{}) []string {
	keys := make([]string, 0, len(elements))
	for k := range elements {
		keys = append(keys, k)
	}
	return keys
}
