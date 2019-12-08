package api

type DistanceFrom struct {
	Coordinates Coordinates
	Place       string
	Min         float32
	Max         float32
}

type Area struct {
	Min float32
	Max float32
}

type Coordinates struct {
	Lon float64
	Lat float64
}

// SortChoice defines how to sort rooms or buildings
type SortChoice int

// SortChoice enum
const (
	SortDistance SortChoice = 0
	SortArea     SortChoice = 1
)
