package models

// Unit represents a Scout unit (troop, pack, etc.)
type Unit struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// Scout represents an individual scout
type Scout struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	UnitID       int    `json:"unit_id"`
	UnitName     string `json:"unit_name,omitempty"`
	FillSchedule bool   `json:"fill_schedule"`
}

// ProgramArea is a camp program area (Aquatics, Eco-STEM, etc.)
type ProgramArea struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TimeBlock is a scheduling period (A, B, AB, C, D, CD, ABCD)
type TimeBlock struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

// Activity is a merit badge or activity session offered at a specific block/half-week
type Activity struct {
	ID               int    `json:"id"`
	Name             string `json:"name"`
	ProgramAreaID    int    `json:"program_area_id"`
	ProgramAreaName  string `json:"program_area_name,omitempty"`
	ProgramAreaColor string `json:"program_area_color,omitempty"`
	TimeBlockID      int    `json:"time_block_id"`
	TimeBlockName    string `json:"time_block_name,omitempty"`
	Capacity         int    `json:"capacity"`
	HalfWeek         string `json:"half_week"`
	HasPrerequisites bool   `json:"has_prerequisites"`
	Enrolled         int    `json:"enrolled,omitempty"`
}

// ScoutPreference stores a scout's desired activity by name and priority
type ScoutPreference struct {
	ID           int    `json:"id"`
	ScoutID      int    `json:"scout_id"`
	ActivityName string `json:"activity_name"`
	Priority     int    `json:"priority"`
}

// Assignment links a scout to a specific activity session
type Assignment struct {
	ID           int    `json:"id"`
	ScoutID      int    `json:"scout_id"`
	ScoutName    string `json:"scout_name,omitempty"`
	ActivityID   int    `json:"activity_id"`
	ActivityName string `json:"activity_name,omitempty"`
	ProgramArea      string `json:"program_area,omitempty"`
	ProgramAreaColor string `json:"program_area_color,omitempty"`
	TimeBlock        string `json:"time_block,omitempty"`
	HalfWeek         string `json:"half_week,omitempty"`
	Locked           bool   `json:"locked"`
}
