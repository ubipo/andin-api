package api

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

func getByUID(db *sqlx.DB, tableConfig TableConfig, uid string, dest interface{}) error {
	err := db.Get(dest, fmt.Sprintf("SELECT %s FROM %s WHERE uid=$1;", tableConfig.Columns, tableConfig.TableName), uid)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Found no %s with <uid> (%s)", tableConfig.elementName(), uid)
	}
	return err
}

func getByID(db *sqlx.DB, tableConfig TableConfig, id int, dest interface{}) error {
	err := db.Get(dest, fmt.Sprintf("SELECT %s FROM %s WHERE id=$1;", tableConfig.Columns, tableConfig.TableName), id)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Found no %s with the a specific internal id, this is a data consistency error that should never occur", tableConfig.elementName())
	}
	return err
}

func getFilteredRoomsByBuildingID(db *sqlx.DB, filterConfig buildingRoomFilterConfig, id int) ([]Room, error) {
	args := []interface{}{id}

	qOptionalLevelFilter := ""
	var levelFilterValue []interface{}
	if filterConfig.level.use {
		qOptionalLevelFilter = fmt.Sprintf("AND level = $%d", len(args)+1)
		levelFilterValue = []interface{}{filterConfig.level.filter}
	}
	args = append(args, levelFilterValue...)

	qOptionalLevelPostfixFilter := ""
	var levelPostfixFilterValue []interface{}
	if filterConfig.levelPostfix.use {
		qOptionalLevelPostfixFilter = fmt.Sprintf("AND level_postfix = $%d", len(args)+1)
		levelPostfixFilterValue = []interface{}{filterConfig.levelPostfix.filter}
	}
	args = append(args, levelPostfixFilterValue...)

	qOptionalNameFilter := ""
	var nameFilterValue []interface{}
	if filterConfig.name.use {
		qOptionalNameFilter = fmt.Sprintf("AND (name ILIKE $%d OR ref ILIKE $%d)", len(args)+1, len(args)+1)
		nameFilterValue = []interface{}{fmt.Sprintf("%%%s%%", filterConfig.name.filter)}
	}
	args = append(args, nameFilterValue...)

	var rooms []Room
	q := fmt.Sprintf(`
		SELECT %s FROM %s WHERE building=$1 %s %s %s;
	`, roomConfig.Columns, roomConfig.TableName, qOptionalLevelFilter, qOptionalLevelPostfixFilter, qOptionalNameFilter)
	err := db.Select(&rooms, q, args...)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Found no rooms for this building")
	}
	return rooms, err
}

func getIntersectingRooms(db *sqlx.DB, filterConfig roomIntersectFilterConfig, id int) ([]Room, error) {
	args := []interface{}{id}

	qOptionalLevelFilter := ""
	var levelFilterValue []interface{}
	if filterConfig.level.use || filterConfig.sameLevel.use {
		var qLevelFilterValue string
		if filterConfig.level.use {
			qLevelFilterValue = "= $2"
			levelFilterValue = []interface{}{filterConfig.level.filter}
		} else {
			if filterConfig.sameLevel.filter {
				qLevelFilterValue = "= (SELECT level FROM a)"
			} else {
				qLevelFilterValue = "<> (SELECT level FROM a)"
			}
		}
		qOptionalLevelFilter = fmt.Sprintf("AND level %s", qLevelFilterValue)
	}
	args = append(args, levelFilterValue...)

	qOptionalLevelPostfixFilter := ""
	var levelPostfixFilterValue []interface{}
	if filterConfig.levelPostfix.use || filterConfig.sameLevelPostfix.use {
		var qLevelPostfixFilterValue string
		if filterConfig.levelPostfix.use {
			qLevelPostfixFilterValue = fmt.Sprintf("= %d", len(args)+1)
			levelPostfixFilterValue = []interface{}{filterConfig.levelPostfix.filter}
		} else {
			if filterConfig.sameLevelPostfix.filter {
				qLevelPostfixFilterValue = "= (SELECT level_postfix FROM a)"
			} else {
				qLevelPostfixFilterValue = "<> (SELECT level_postfix FROM a)"
			}
		}
		qOptionalLevelPostfixFilter = fmt.Sprintf("AND level_postfix %s", qLevelPostfixFilterValue)
	}
	args = append(args, levelPostfixFilterValue...)

	var rooms []Room
	q := fmt.Sprintf(`
		WITH a AS (
			SELECT geometry, level, level_postfix FROM %s WHERE id = $1
		)
		SELECT %s FROM %s AS b WHERE id<>$1 AND ST_Intersects((select geometry from a), b.geometry) %s %s;
	`, roomConfig.TableName, roomConfig.Columns, roomConfig.TableName, qOptionalLevelFilter, qOptionalLevelPostfixFilter)
	err := db.Select(&rooms, q, args...)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("Found no rooms that intersect the given room")
	}
	return rooms, err
}

const qAreaColumn = ", ST_Area(geometry) AS area"
const qAreaFilter = "AND area BETWEEN $5 AND $6"
const qSort = "ORDER BY"

func getFiltered(db *sqlx.DB, tableConfig TableConfig, filterConfig rootGeographyFilterConfig, dest interface{}) error {
	qOptionalAreaColumn := ""
	qOptionalAreaFilter := ""
	var areaArgs []interface{}
	if filterConfig.area.use {
		qOptionalAreaColumn = qAreaColumn
		qOptionalAreaFilter = qAreaFilter
		areaArgs = []interface{}{filterConfig.area.filter.Min, filterConfig.area.filter.Max}
	}

	qOptionalSort := ""
	if filterConfig.sortChoice.use {
		switch filterConfig.sortChoice.filter {
		case SortDistance:
			qOptionalSort = fmt.Sprintf("%s distance", qSort)
		case SortArea:
			qOptionalSort = fmt.Sprintf("%s area", qSort)
		}
	}
	df := filterConfig.distanceFrom
	args := append([]interface{}{df.Coordinates.Lon, df.Coordinates.Lat, df.Min, df.Max}, areaArgs...)
	q := fmt.Sprintf(`
		SELECT * FROM (
			SELECT %s, ST_Distance(ST_MakePoint($1, $2), geometry) as distance %s FROM %s
		) AS ti WHERE distance BETWEEN $3 AND $4 %s %s;
	`, tableConfig.Columns, qOptionalAreaColumn, tableConfig.TableName, qOptionalAreaFilter, qOptionalSort)

	err := db.Select(dest, q, args...)
	if err == sql.ErrNoRows {
		return fmt.Errorf("Found no %s for the specified filters, maybe broaden your search?", tableConfig.elementNamePlural())
	}
	return err
}
