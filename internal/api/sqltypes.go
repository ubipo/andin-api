package api

import "time"

// TableConfig represents a table in the sql data structure
type TableConfig struct {
	TableName         string
	ElementName       string
	ElementNamePlural string
	Columns           string
}

func (tableConfig *TableConfig) elementName() string {
	if tableConfig.ElementName != "" {
		return tableConfig.ElementName
	}
	return tableConfig.TableName
}

func (tableConfig *TableConfig) elementNamePlural() string {
	if tableConfig.ElementNamePlural != "" {
		return tableConfig.ElementNamePlural
	}
	return tableConfig.elementName() + "s"
}

// Survey represents an sql survey
type Survey struct {
	ID       int
	UID      string `json:"uid"`
	Surveyor string `json:"surveyor"`
	External bool   `json:"external"`
}

var surveyConfig = TableConfig{
	TableName: "survey",
	Columns:   "id, uid, surveyor, external",
}

// OsmElement represents an sql osm_element
type OsmElement struct {
	ID         int
	UID        string `json:"uid"`
	OsmID      int    `json:"id" db:"osm_id"`
	OsmType    string `json:"type" db:"osm_type"`
	OsmVersion int    `json:"version" db:"osm_version"`
}

var osmElementConfig = TableConfig{
	TableName:   "osm_element",
	ElementName: "osm element",
	Columns:     "id, uid, osm_id, osm_type, osm_version",
}

// Simport represents an sql import (import is a reserved keyword)
type Simport struct {
	ID     int
	UID    string    `json:"uid"`
	Date   time.Time `json:"date"`
	Script string    `json:"script"`
}

var simportConfig = TableConfig{
	TableName: "import",
	Columns:   "id, uid, date, script",
}

// DataSource represents an sql data_source
type DataSource struct {
	ID     int
	Osm    *int `json:"osm"`
	Survey *int `json:"survey"`
	Import int  `json:"import"`
}

var dataSourceConfig = TableConfig{
	TableName:   "data_source",
	ElementName: "data source",
	Columns:     "id, osm, survey, import",
}

// Address represents an sql address
type Address struct {
	ID       int
	Free     string `json:"osm"`
	Locality string `json:"survey"`
	Region   string `json:"region"`
	Postcode string `json:"postcode"`
	Country  string `json:"country"`
}

var addressConfig = TableConfig{
	TableName: "address",
	Columns:   "id, free, locality, region, postcode, country",
}

// Building represents an sql building
type Building struct {
	ID         int
	UID        string  `json:"uid"`
	Name       *string `json:"name"`
	Geometry   string  `json:"geometry"`
	Address    int     `json:"address"`
	DataSource int     `json:"dataSource" db:"data_source"`
}

var buildingConfig = TableConfig{
	TableName: "building",
	Columns:   "id, uid, name, ST_AsText(geometry) as geometry, address, data_source",
}

// Room represents an sql room
type Room struct {
	ID           int
	UID          string  `json:"uid"`
	Name         *string `json:"name"`
	Geometry     string  `json:"geometry"`
	Level        int     `json:"level"`
	LevelPostfix *string `json:"levelPostfix" db:"level_postfix"`
	Ref          *string `json:"ref"`
	Building     int     `json:"building"`
	DataSource   int     `json:"dataSource" db:"data_source"`
}

var roomConfig = TableConfig{
	TableName: "room",
	Columns:   "id, uid, name, ST_AsText(geometry) as geometry, level, level_postfix, ref, building, data_source",
}
