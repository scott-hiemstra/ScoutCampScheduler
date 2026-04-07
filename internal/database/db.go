package database

import (
	"database/sql"

	"summer-camp-scheduler/internal/models"
)

type Store struct {
	db *sql.DB
}

func NewStore(db *sql.DB) *Store {
	return &Store{db: db}
}

// --------------- Units ---------------

func (s *Store) CreateUnit(name string) (int64, error) {
	res, err := s.db.Exec("INSERT INTO units (name) VALUES (?)", name)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListUnits() ([]models.Unit, error) {
	rows, err := s.db.Query("SELECT id, name FROM units ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var units []models.Unit
	for rows.Next() {
		var u models.Unit
		if err := rows.Scan(&u.ID, &u.Name); err != nil {
			return nil, err
		}
		units = append(units, u)
	}
	return units, nil
}

func (s *Store) DeleteUnit(id int) error {
	_, err := s.db.Exec("DELETE FROM units WHERE id = ?", id)
	return err
}

// --------------- Scouts ---------------

func (s *Store) AddScout(name string, unitID int) (int64, error) {
	res, err := s.db.Exec("INSERT INTO scouts (name, unit_id) VALUES (?, ?)", name, unitID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) ListScoutsByUnit(unitID int) ([]models.Scout, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.name, s.unit_id, u.name, s.fill_schedule
		FROM scouts s JOIN units u ON s.unit_id = u.id
		WHERE s.unit_id = ? ORDER BY s.name`, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scouts []models.Scout
	for rows.Next() {
		var sc models.Scout
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.UnitID, &sc.UnitName, &sc.FillSchedule); err != nil {
			return nil, err
		}
		scouts = append(scouts, sc)
	}
	return scouts, nil
}

func (s *Store) ListAllScouts() ([]models.Scout, error) {
	rows, err := s.db.Query(`
		SELECT s.id, s.name, s.unit_id, u.name, s.fill_schedule
		FROM scouts s JOIN units u ON s.unit_id = u.id
		ORDER BY u.name, s.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var scouts []models.Scout
	for rows.Next() {
		var sc models.Scout
		if err := rows.Scan(&sc.ID, &sc.Name, &sc.UnitID, &sc.UnitName, &sc.FillSchedule); err != nil {
			return nil, err
		}
		scouts = append(scouts, sc)
	}
	return scouts, nil
}

func (s *Store) SetFillSchedule(scoutID int, fill bool) error {
	_, err := s.db.Exec("UPDATE scouts SET fill_schedule = ? WHERE id = ?", fill, scoutID)
	return err
}

func (s *Store) DeleteScout(id int) error {
	_, err := s.db.Exec("DELETE FROM scouts WHERE id = ?", id)
	return err
}

// --------------- Preferences ---------------

func (s *Store) GetPreferences(scoutID int) ([]models.ScoutPreference, error) {
	rows, err := s.db.Query(`
		SELECT id, scout_id, activity_name, priority
		FROM scout_preferences WHERE scout_id = ? ORDER BY priority`, scoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var prefs []models.ScoutPreference
	for rows.Next() {
		var p models.ScoutPreference
		if err := rows.Scan(&p.ID, &p.ScoutID, &p.ActivityName, &p.Priority); err != nil {
			return nil, err
		}
		prefs = append(prefs, p)
	}
	return prefs, nil
}

func (s *Store) SetPreferences(scoutID int, prefs []models.ScoutPreference) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if _, err := tx.Exec("DELETE FROM scout_preferences WHERE scout_id = ?", scoutID); err != nil {
		return err
	}
	for _, p := range prefs {
		if _, err := tx.Exec(
			"INSERT INTO scout_preferences (scout_id, activity_name, priority) VALUES (?, ?, ?)",
			scoutID, p.ActivityName, p.Priority,
		); err != nil {
			return err
		}
	}
	return tx.Commit()
}

// --------------- Activities ---------------

func (s *Store) ListActivities() ([]models.Activity, error) {
	rows, err := s.db.Query(`
		SELECT a.id, a.name, a.program_area_id, pa.name, pa.color,
		       a.time_block_id, tb.name, a.capacity, a.half_week, a.has_prerequisites,
		       (SELECT COUNT(*) FROM assignments asg WHERE asg.activity_id = a.id) as enrolled
		FROM activities a
		JOIN program_areas pa ON a.program_area_id = pa.id
		JOIN time_blocks tb ON a.time_block_id = tb.id
		ORDER BY pa.id, tb.id, a.half_week, a.name`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var activities []models.Activity
	for rows.Next() {
		var act models.Activity
		if err := rows.Scan(&act.ID, &act.Name, &act.ProgramAreaID, &act.ProgramAreaName,
			&act.ProgramAreaColor, &act.TimeBlockID, &act.TimeBlockName,
			&act.Capacity, &act.HalfWeek, &act.HasPrerequisites, &act.Enrolled); err != nil {
			return nil, err
		}
		activities = append(activities, act)
	}
	return activities, nil
}

func (s *Store) ListActivityNames() ([]string, error) {
	rows, err := s.db.Query("SELECT DISTINCT name FROM activities ORDER BY name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var names []string
	for rows.Next() {
		var n string
		if err := rows.Scan(&n); err != nil {
			return nil, err
		}
		names = append(names, n)
	}
	return names, nil
}

// --------------- Assignments ---------------

func (s *Store) GetAssignments(unitID int) ([]models.Assignment, error) {
	query := `
		SELECT asg.id, asg.scout_id, s.name, asg.activity_id, a.name,
		       pa.name, tb.name, a.half_week, asg.locked
		FROM assignments asg
		JOIN scouts s ON asg.scout_id = s.id
		JOIN activities a ON asg.activity_id = a.id
		JOIN program_areas pa ON a.program_area_id = pa.id
		JOIN time_blocks tb ON a.time_block_id = tb.id`
	var rows *sql.Rows
	var err error
	if unitID > 0 {
		query += " WHERE s.unit_id = ? ORDER BY s.name, tb.id, a.half_week"
		rows, err = s.db.Query(query, unitID)
	} else {
		query += " ORDER BY s.name, tb.id, a.half_week"
		rows, err = s.db.Query(query)
	}
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assignments []models.Assignment
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.ScoutID, &a.ScoutName, &a.ActivityID,
			&a.ActivityName, &a.ProgramArea, &a.TimeBlock, &a.HalfWeek, &a.Locked); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (s *Store) GetAllAssignments() ([]models.Assignment, error) {
	return s.GetAssignments(0)
}

func (s *Store) GetUnitAssignmentsWithColor(unitID int) ([]models.Assignment, error) {
	query := `
		SELECT asg.id, asg.scout_id, sc.name, asg.activity_id, a.name,
		       pa.name, tb.name, a.half_week, asg.locked, pa.color
		FROM assignments asg
		JOIN scouts sc ON asg.scout_id = sc.id
		JOIN activities a ON asg.activity_id = a.id
		JOIN program_areas pa ON a.program_area_id = pa.id
		JOIN time_blocks tb ON a.time_block_id = tb.id
		WHERE sc.unit_id = ?
		ORDER BY pa.id, a.name, tb.id, a.half_week, sc.name`
	rows, err := s.db.Query(query, unitID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assignments []models.Assignment
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.ScoutID, &a.ScoutName, &a.ActivityID,
			&a.ActivityName, &a.ProgramArea, &a.TimeBlock, &a.HalfWeek, &a.Locked, &a.ProgramAreaColor); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (s *Store) CreateAssignment(scoutID, activityID int) (int64, error) {
	res, err := s.db.Exec("INSERT INTO assignments (scout_id, activity_id) VALUES (?, ?)", scoutID, activityID)
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) DeleteAssignment(id int) error {
	_, err := s.db.Exec("DELETE FROM assignments WHERE id = ? AND locked = FALSE", id)
	return err
}

func (s *Store) ToggleLock(id int) error {
	_, err := s.db.Exec("UPDATE assignments SET locked = NOT locked WHERE id = ?", id)
	return err
}

func (s *Store) ClearUnlockedAssignments() error {
	_, err := s.db.Exec("DELETE FROM assignments WHERE locked = FALSE")
	return err
}

// --------------- Lookups ---------------

func (s *Store) GetScoutAssignments(scoutID int) ([]models.Assignment, error) {
	query := `
		SELECT asg.id, asg.scout_id, sc.name, asg.activity_id, a.name,
		       pa.name, tb.name, a.half_week, asg.locked, pa.color
		FROM assignments asg
		JOIN scouts sc ON asg.scout_id = sc.id
		JOIN activities a ON asg.activity_id = a.id
		JOIN program_areas pa ON a.program_area_id = pa.id
		JOIN time_blocks tb ON a.time_block_id = tb.id
		WHERE asg.scout_id = ?
		ORDER BY tb.id, a.half_week`
	rows, err := s.db.Query(query, scoutID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var assignments []models.Assignment
	for rows.Next() {
		var a models.Assignment
		if err := rows.Scan(&a.ID, &a.ScoutID, &a.ScoutName, &a.ActivityID,
			&a.ActivityName, &a.ProgramArea, &a.TimeBlock, &a.HalfWeek, &a.Locked, &a.ProgramAreaColor); err != nil {
			return nil, err
		}
		assignments = append(assignments, a)
	}
	return assignments, nil
}

func (s *Store) ListProgramAreas() ([]models.ProgramArea, error) {
	rows, err := s.db.Query("SELECT id, name, color FROM program_areas ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var areas []models.ProgramArea
	for rows.Next() {
		var pa models.ProgramArea
		if err := rows.Scan(&pa.ID, &pa.Name, &pa.Color); err != nil {
			return nil, err
		}
		areas = append(areas, pa)
	}
	return areas, nil
}

func (s *Store) ListTimeBlocks() ([]models.TimeBlock, error) {
	rows, err := s.db.Query("SELECT id, name, start_time, end_time FROM time_blocks ORDER BY id")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var blocks []models.TimeBlock
	for rows.Next() {
		var tb models.TimeBlock
		if err := rows.Scan(&tb.ID, &tb.Name, &tb.StartTime, &tb.EndTime); err != nil {
			return nil, err
		}
		blocks = append(blocks, tb)
	}
	return blocks, nil
}

func (s *Store) GetStats() (map[string]int, error) {
	stats := make(map[string]int)
	var units, scouts, activities, assignments int
	s.db.QueryRow("SELECT COUNT(*) FROM units").Scan(&units)
	s.db.QueryRow("SELECT COUNT(*) FROM scouts").Scan(&scouts)
	s.db.QueryRow("SELECT COUNT(DISTINCT name) FROM activities").Scan(&activities)
	s.db.QueryRow("SELECT COUNT(*) FROM assignments").Scan(&assignments)
	stats["units"] = units
	stats["scouts"] = scouts
	stats["activities"] = activities
	stats["assignments"] = assignments
	return stats, nil
}
