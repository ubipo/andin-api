package api

import (
	"fmt"

	"github.com/mitchellh/mapstructure"
)

type rootGeographyFilterConfig struct {
	distanceFrom DistanceFrom
	area         optionalAreaFilter
	sortChoice   optionalSortChoiceFilter
}

func parseRootGeographyFilterArgs(args map[string]interface{}) (rootGeographyFilterConfig, error) {
	var config rootGeographyFilterConfig

	df := args["distanceFrom"].(map[string]interface{})
	coordsSpecified := df["coordinates"] != nil
	placeSpecified := df["place"] != nil
	var distanceFrom DistanceFrom
	mapstructure.Decode(args["distanceFrom"], &distanceFrom)
	if distanceFrom.Max > maxFilterDistance {
		return config, fmt.Errorf("<distanceFrom.Max> (%f) cannot be greater than %d", distanceFrom.Max, maxFilterDistance)
	}

	if placeSpecified {
		geocoded, err := geocode(df["place"].(string))
		if err != nil {
			if !coordsSpecified {
				return config, fmt.Errorf("error geocoding <place> for distanceFrom filter without fallback <coordinates> (\"%s\")", err)
			}
		}
		distanceFrom.Coordinates = geocoded
	} else if !coordsSpecified {
		return config, fmt.Errorf("must specify either <coordinates> or <place> on distanceFrom filter")
	}

	areaArg := args["area"]
	var area Area
	mapstructure.Decode(areaArg, &area)

	sortArg := args["sort"]
	var sortChoice SortChoice
	if sortArg != nil {
		sortChoice = SortChoice(sortArg.(int))
	}

	return rootGeographyFilterConfig{
		distanceFrom: distanceFrom,
		area:         optionalAreaFilter{areaArg != nil, area},
		sortChoice:   optionalSortChoiceFilter{sortArg != nil, sortChoice},
	}, nil
}

type roomIntersectFilterConfig struct {
	level            optionalIntFilter
	levelPostfix     optionalStringFilter
	sameLevel        optionalBoolFilter
	sameLevelPostfix optionalBoolFilter
}

func parseRoomIntersectFilterArgs(args map[string]interface{}) (roomIntersectFilterConfig, error) {
	levelArg := args["level"]
	var level int
	if levelArg != nil {
		level = levelArg.(int)
	}

	sameLevelArg := args["sameLevel"]
	var sameLevel bool
	if sameLevelArg != nil {
		sameLevel = sameLevelArg.(bool)
		if levelArg != nil {
			return roomIntersectFilterConfig{}, fmt.Errorf("cannot filter on both <level> and <sameLevel> at the same time")
		}
	}

	levelPostfixArg := args["levelPostfix"]
	var levelPostfix string
	if levelPostfixArg != nil {
		levelPostfix = levelPostfixArg.(string)
	}

	sameLevelPostfixArg := args["sameLevelPostfix"]
	var sameLevelPostfix bool
	if sameLevelPostfixArg != nil {
		sameLevelPostfix = sameLevelPostfixArg.(bool)
		if levelPostfixArg != nil {
			return roomIntersectFilterConfig{}, fmt.Errorf("cannot filter on both <levelPostfix> and <sameLevelPostfix> at the same time")
		}
	}

	return roomIntersectFilterConfig{
		level:            optionalIntFilter{levelArg != nil, level},
		levelPostfix:     optionalStringFilter{levelPostfixArg != nil, levelPostfix},
		sameLevel:        optionalBoolFilter{sameLevelArg != nil, sameLevel},
		sameLevelPostfix: optionalBoolFilter{sameLevelPostfixArg != nil, sameLevelPostfix},
	}, nil
}

type buildingRoomFilterConfig struct {
	level        optionalIntFilter
	levelPostfix optionalStringFilter
	name         optionalStringFilter
}

func parseBuildingRoomFilterArgs(args map[string]interface{}) (buildingRoomFilterConfig, error) {
	levelArg := args["level"]
	var level int
	if levelArg != nil {
		level = levelArg.(int)
	}

	levelPostfixArg := args["levelPostfix"]
	var levelPostfix string
	if levelPostfixArg != nil {
		levelPostfix = levelPostfixArg.(string)
	}

	nameArg := args["name"]
	var name string
	if nameArg != nil {
		name = nameArg.(string)
	}

	return buildingRoomFilterConfig{
		level:        optionalIntFilter{levelArg != nil, level},
		levelPostfix: optionalStringFilter{levelPostfixArg != nil, levelPostfix},
		name:         optionalStringFilter{nameArg != nil, name},
	}, nil
}
