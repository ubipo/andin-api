package api

type FilteredRoom struct {
	Distance float32
	Area     *float32
	Room
}

type FilteredBuilding struct {
	Distance float32
	Area     *float32
	Building
}
