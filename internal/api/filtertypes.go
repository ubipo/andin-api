package api

type optionalIntFilter struct {
	use    bool
	filter int
}

type optionalStringFilter struct {
	use    bool
	filter string
}

type optionalBoolFilter struct {
	use    bool
	filter bool
}

type optionalAreaFilter struct {
	use    bool
	filter Area
}

type optionalSortChoiceFilter struct {
	use    bool
	filter SortChoice
}
