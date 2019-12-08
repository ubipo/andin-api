package api

import (
	"log"

	"github.com/graphql-go/graphql"
	"github.com/jmoiron/sqlx"
)

func gqlSF(scalar *graphql.Scalar) *graphql.Field {
	return &graphql.Field{
		Type: scalar,
	}
}

func gqlRootGeographyFilteredObject(name string, fieldName string, wrappedType *graphql.Object) *graphql.Object {
	return graphql.NewObject(graphql.ObjectConfig{
		Name: name,
		Fields: graphql.Fields{
			fieldName: &graphql.Field{
				Type: wrappedType,
			},
			"distance": gqlSF(graphql.Float),
			"area":     gqlSF(graphql.Float),
		},
	})
}

const maxFilterDistance = 2000

func generateSchema(db *sqlx.DB) graphql.Schema {
	/*
		> Types
		Mirrors the sql datastructure but with extra fields for indirect relations.
	*/
	var surveyType graphql.Object
	var osmElementType graphql.Object
	var importType graphql.Object
	var dataSourceType graphql.Object
	var addressType graphql.Object
	var buildingType graphql.Object
	var roomType graphql.Object

	surveyType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "Survey",
		Fields: graphql.Fields{
			"uid":      gqlSF(graphql.String),
			"surveyor": gqlSF(graphql.String),
			"external": gqlSF(graphql.Boolean),
		},
	})

	osmElementType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "OsmElement",
		Fields: graphql.Fields{
			"uid": gqlSF(graphql.String),
			"id": &graphql.Field{
				Type: graphql.Int,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(OsmElement).OsmID
					return id, nil
				},
			},
			"type":    gqlSF(graphql.String),
			"version": gqlSF(graphql.String),
		},
	})

	importType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "Import",
		Fields: graphql.Fields{
			"uid":    gqlSF(graphql.String),
			"date":   gqlSF(graphql.DateTime),
			"script": gqlSF(graphql.String),
		},
	})

	dataSourceType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "DataSource",
		Fields: graphql.Fields{
			"osm": &graphql.Field{
				Type: &osmElementType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(DataSource).Osm
					if id == nil {
						return nil, nil
					}
					var osmElement OsmElement
					err := getByID(db, osmElementConfig, *id, &osmElement)
					return osmElement, err
				},
			},
			"survey": &graphql.Field{
				Type: &surveyType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(DataSource).Survey
					if id == nil {
						return nil, nil
					}
					var survey Survey
					err := getByID(db, surveyConfig, *id, &survey)
					return survey, err
				},
			},
			"import": &graphql.Field{
				Type: &importType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(DataSource).Import
					var simport Simport
					err := getByID(db, simportConfig, id, &simport)
					return simport, err
				},
			},
		},
	})

	addressType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "Address",
		Fields: graphql.Fields{
			"free":     gqlSF(graphql.String),
			"locality": gqlSF(graphql.String),
			"region":   gqlSF(graphql.String),
			"postcode": gqlSF(graphql.String),
			"country":  gqlSF(graphql.String),
		},
	})

	buildingType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "Building",
		Fields: graphql.Fields{
			"uid":      gqlSF(graphql.String),
			"name":     gqlSF(graphql.String),
			"geometry": gqlSF(graphql.String),
			"address": &graphql.Field{
				Type: &addressType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(Building).Address
					var address Address
					err := getByID(db, addressConfig, id, &address)
					return address, err
				},
			},
			"dataSource": &graphql.Field{
				Type: &dataSourceType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(Building).DataSource
					var dataSource DataSource
					err := getByID(db, dataSourceConfig, id, &dataSource)
					return dataSource, err
				},
			},
			"rooms": &graphql.Field{
				Type: &graphql.List{
					OfType: &roomType,
				},
				Args: graphql.FieldConfigArgument{
					"level": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"levelPostfix": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"name": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					filterConfig, err := parseBuildingRoomFilterArgs(params.Args)
					if err != nil {
						return nil, err
					}
					id := params.Source.(Building).ID
					rooms, err := getFilteredRoomsByBuildingID(db, filterConfig, id)
					return rooms, err
				},
			},
		},
	})

	roomType = *graphql.NewObject(graphql.ObjectConfig{
		Name: "Room",
		Fields: graphql.Fields{
			"uid":          gqlSF(graphql.String),
			"name":         gqlSF(graphql.String),
			"geometry":     gqlSF(graphql.String),
			"level":        gqlSF(graphql.Int),
			"levelPostfix": gqlSF(graphql.String),
			"ref":          gqlSF(graphql.String),
			"building": &graphql.Field{
				Type: &buildingType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(Room).Building
					var building Building
					err := getByID(db, buildingConfig, id, &building)
					return building, err
				},
			},
			"dataSource": &graphql.Field{
				Type: &dataSourceType,
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					id := params.Source.(Room).DataSource
					var dataSource DataSource
					err := getByID(db, dataSourceConfig, id, &dataSource)
					return dataSource, err
				},
			},
			"intersecting": &graphql.Field{
				Type: &graphql.List{
					OfType: &roomType,
				},
				Args: graphql.FieldConfigArgument{
					"level": &graphql.ArgumentConfig{
						Type: graphql.Int,
					},
					"levelPostfix": &graphql.ArgumentConfig{
						Type: graphql.String,
					},
					"sameLevel": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
					"sameLevelPostfix": &graphql.ArgumentConfig{
						Type: graphql.Boolean,
					},
				},
				Resolve: func(params graphql.ResolveParams) (interface{}, error) {
					filterConfig, err := parseRoomIntersectFilterArgs(params.Args)
					if err != nil {
						return nil, err
					}
					id := params.Source.(Room).ID
					rooms, err := getIntersectingRooms(db, filterConfig, id)
					return rooms, err
				},
			},
		},
	})

	var filteredRoomType = *gqlRootGeographyFilteredObject("FilteredRoom", "room", &roomType)
	var filteredBuildingType = *gqlRootGeographyFilteredObject("FilteredBuilding", "building", &buildingType)

	var coordinatesType = graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name: "Coordinates",
			Fields: graphql.InputObjectConfigFieldMap{
				"lon": &graphql.InputObjectFieldConfig{
					Type: graphql.NewNonNull(graphql.Float),
				},
				"lat": &graphql.InputObjectFieldConfig{
					Type: graphql.NewNonNull(graphql.Float),
				},
			},
		},
	)

	var distanceFromType = graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name: "DistanceFrom",
			Fields: graphql.InputObjectConfigFieldMap{
				"coordinates": &graphql.InputObjectFieldConfig{
					Type: coordinatesType,
				},
				"place": &graphql.InputObjectFieldConfig{
					Type: graphql.String,
				},
				"min": &graphql.InputObjectFieldConfig{
					Type:         graphql.Int,
					DefaultValue: 0,
				},
				"max": &graphql.InputObjectFieldConfig{
					Type:         graphql.Int,
					DefaultValue: 500,
				},
			},
		},
	)

	var areaType = graphql.NewInputObject(
		graphql.InputObjectConfig{
			Name: "Area",
			Fields: graphql.InputObjectConfigFieldMap{
				"min": &graphql.InputObjectFieldConfig{
					Type:         graphql.Int,
					DefaultValue: 0,
				},
				"max": &graphql.InputObjectFieldConfig{
					Type:         graphql.Int,
					DefaultValue: 500,
				},
			},
		},
	)

	sortEnum := graphql.NewEnum(graphql.EnumConfig{
		Name: "SortEnum",
		Values: graphql.EnumValueConfigMap{
			"DISTANCE": &graphql.EnumValueConfig{
				Value: 0,
			},
			"AREA": &graphql.EnumValueConfig{
				Value: 1,
			},
		},
	})

	rootGeographyFilterArgs := graphql.FieldConfigArgument{
		"distanceFrom": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(distanceFromType),
		},
		"area": &graphql.ArgumentConfig{
			Type: areaType,
		},
		"sort": &graphql.ArgumentConfig{
			Type: sortEnum,
		},
	}

	uidArgs := graphql.FieldConfigArgument{
		"uid": &graphql.ArgumentConfig{
			Type: graphql.NewNonNull(graphql.String),
		},
	}

	/*
		> Root Fields
		Base fields to start a query from.
		Includes i.a. all elements with a uid.
	*/
	rootFields := graphql.Fields{
		"building": &graphql.Field{
			Type: &buildingType,
			Args: uidArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				uid := params.Args["uid"].(string)
				var building Building
				err := getByUID(db, buildingConfig, uid, &building)
				return building, err
			},
		},
		"room": &graphql.Field{
			Type: &roomType,
			Args: uidArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				uid := params.Args["uid"].(string)
				var room Room
				err := getByUID(db, roomConfig, uid, &room)
				return room, err
			},
		},
		"rooms": &graphql.Field{
			Type: &graphql.List{
				OfType: &filteredRoomType,
			},
			Args: rootGeographyFilterArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				filterConfig, err := parseRootGeographyFilterArgs(params.Args)
				if err != nil {
					return nil, err
				}
				var rooms []FilteredRoom
				err = getFiltered(db, roomConfig, filterConfig, &rooms)
				return rooms, err
			},
		},
		"buildings": &graphql.Field{
			Type: &graphql.List{
				OfType: &filteredBuildingType,
			},
			Args: rootGeographyFilterArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				filterConfig, err := parseRootGeographyFilterArgs(params.Args)
				if err != nil {
					return nil, err
				}
				var buildings []FilteredBuilding
				err = getFiltered(db, buildingConfig, filterConfig, &buildings)
				return buildings, err
			},
		},
		"import": &graphql.Field{
			Type: &importType,
			Args: uidArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				uid := params.Args["uid"].(string)
				var simport Simport
				err := getByUID(db, simportConfig, uid, &simport)
				return simport, err
			},
		},
		"osmElement": &graphql.Field{
			Type: &osmElementType,
			Args: uidArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				uid := params.Args["uid"].(string)
				var osmElement OsmElement
				err := getByUID(db, osmElementConfig, uid, &osmElement)
				return osmElement, err
			},
		},
		"survey": &graphql.Field{
			Type: &surveyType,
			Args: uidArgs,
			Resolve: func(params graphql.ResolveParams) (interface{}, error) {
				uid := params.Args["uid"].(string)
				var survey Survey
				err := getByUID(db, surveyConfig, uid, &survey)
				return survey, err
			},
		},
	}

	/*
		> Schema
	*/
	rootQuery := graphql.NewObject(
		graphql.ObjectConfig{
			Name:   "RootQuery",
			Fields: rootFields,
		},
	)
	schemaConfig := graphql.SchemaConfig{Query: rootQuery}
	schema, err := graphql.NewSchema(schemaConfig)
	if err != nil {
		log.Fatalf("failed to create new schema, error: %v", err)
	}

	return schema
}
