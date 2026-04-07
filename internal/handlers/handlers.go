package handlers

import (
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"summer-camp-scheduler/internal/database"
	"summer-camp-scheduler/internal/models"
	"summer-camp-scheduler/internal/scheduler"
)

type Handler struct {
	store *database.Store
	tmpl  *template.Template
}

func New(store *database.Store) *Handler {
	funcMap := template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}
	tmpl := template.Must(template.New("").Funcs(funcMap).ParseGlob("web/templates/*.html"))
	return &Handler{store: store, tmpl: tmpl}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /{$}", h.dashboard)

	mux.HandleFunc("GET /units", h.unitsPage)
	mux.HandleFunc("POST /units", h.createUnit)
	mux.HandleFunc("POST /units/{id}/delete", h.deleteUnit)
	mux.HandleFunc("GET /units/{id}/schedule", h.unitScheduleView)
	mux.HandleFunc("GET /units/{id}/roster", h.unitRosterView)

	mux.HandleFunc("GET /scouts", h.scoutsPage)
	mux.HandleFunc("POST /scouts", h.addScout)
	mux.HandleFunc("POST /scouts/{id}/delete", h.deleteScout)
	mux.HandleFunc("GET /scouts/{id}/preferences", h.preferencesPage)
	mux.HandleFunc("POST /scouts/{id}/preferences", h.savePreferences)
	mux.HandleFunc("GET /scouts/{id}/schedule", h.scoutScheduleView)

	mux.HandleFunc("GET /activities", h.activitiesPage)
	mux.HandleFunc("GET /schedule", h.schedulePage)
	mux.HandleFunc("POST /schedule/run", h.runScheduler)
	mux.HandleFunc("POST /schedule/clear", h.clearSchedule)
	mux.HandleFunc("POST /assignments/{id}/delete", h.deleteAssignment)
	mux.HandleFunc("POST /assignments/{id}/lock", h.toggleLock)
	mux.HandleFunc("POST /assignments", h.addAssignment)

	mux.HandleFunc("GET /api/scouts", h.apiScouts)
}

// ============ Dashboard ============

func (h *Handler) dashboard(w http.ResponseWriter, r *http.Request) {
	stats, _ := h.store.GetStats()
	h.render(w, "dashboard.html", map[string]any{"Stats": stats, "Page": "dashboard"})
}

// ============ Units ============

func (h *Handler) unitsPage(w http.ResponseWriter, r *http.Request) {
	units, _ := h.store.ListUnits()
	h.render(w, "units.html", map[string]any{"Units": units, "Page": "units"})
}

func (h *Handler) createUnit(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	if name != "" {
		if _, err := h.store.CreateUnit(name); err != nil {
			log.Printf("Error creating unit: %v", err)
		}
	}
	http.Redirect(w, r, "/units", http.StatusFound)
}

func (h *Handler) deleteUnit(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	h.store.DeleteUnit(id)
	http.Redirect(w, r, "/units", http.StatusFound)
}

// ============ Scouts ============

func (h *Handler) scoutsPage(w http.ResponseWriter, r *http.Request) {
	units, _ := h.store.ListUnits()
	scouts, _ := h.store.ListAllScouts()

	type UnitGroup struct {
		Unit   models.Unit
		Scouts []models.Scout
	}
	scoutMap := make(map[int][]models.Scout)
	for _, sc := range scouts {
		scoutMap[sc.UnitID] = append(scoutMap[sc.UnitID], sc)
	}
	var groups []UnitGroup
	for _, u := range units {
		groups = append(groups, UnitGroup{Unit: u, Scouts: scoutMap[u.ID]})
	}

	h.render(w, "scouts.html", map[string]any{
		"Units":  units,
		"Groups": groups,
		"Page":   "scouts",
	})
}

func (h *Handler) addScout(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	name := strings.TrimSpace(r.FormValue("name"))
	unitID, _ := strconv.Atoi(r.FormValue("unit_id"))
	if name != "" && unitID > 0 {
		if _, err := h.store.AddScout(name, unitID); err != nil {
			log.Printf("Error adding scout: %v", err)
		}
	}
	http.Redirect(w, r, "/scouts", http.StatusFound)
}

func (h *Handler) deleteScout(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	h.store.DeleteScout(id)
	http.Redirect(w, r, "/scouts", http.StatusFound)
}

// ---- Schedule grid helpers ----

type scheduleCell struct {
	Activity string
	Area     string
	Color    string
	RowSpan  int
	IsBreak  bool
	IsEmpty  bool
	Skip     bool
}

type scheduleRow struct {
	Label    string
	TimeInfo string
	IsLunch  bool
	Cells    []scheduleCell // [Mon/Tue, Wed/Thu]
}

func buildScheduleRows(assignments []models.Assignment) []scheduleRow {
	type asgKey struct{ hw, block string }
	asgMap := make(map[asgKey]models.Assignment)
	for _, asg := range assignments {
		hws := []string{asg.HalfWeek}
		if asg.HalfWeek == "full" {
			hws = []string{"first", "second"}
		}
		var blocks []string
		switch asg.TimeBlock {
		case "AB":
			blocks = []string{"A", "B"}
		case "CD":
			blocks = []string{"C", "D"}
		case "ABCD":
			blocks = []string{"A", "B", "C", "D"}
		default:
			blocks = []string{asg.TimeBlock}
		}
		for _, hw := range hws {
			for _, b := range blocks {
				asgMap[asgKey{hw, b}] = asg
			}
		}
	}

	type blockDef struct {
		name, time string
		isLunch    bool
	}
	blockDefs := []blockDef{
		{"A", "9:00 \u2013 10:30", false},
		{"B", "10:30 \u2013 12:00", false},
		{"lunch", "", true},
		{"C", "2:00 \u2013 3:30", false},
		{"D", "3:30 \u2013 5:00", false},
	}
	halfWeeks := []string{"first", "second"}

	skipCell := make(map[string]bool)

	var rows []scheduleRow
	for _, bd := range blockDefs {
		row := scheduleRow{Label: bd.name, TimeInfo: bd.time, IsLunch: bd.isLunch}
		if bd.isLunch {
			for _, hw := range halfWeeks {
				if skipCell["lunch|"+hw] {
					row.Cells = append(row.Cells, scheduleCell{Skip: true})
				} else {
					row.Cells = append(row.Cells, scheduleCell{IsBreak: true, RowSpan: 1})
				}
			}
			rows = append(rows, row)
			continue
		}
		for _, hw := range halfWeeks {
			key := bd.name + "|" + hw
			if skipCell[key] {
				row.Cells = append(row.Cells, scheduleCell{Skip: true})
				continue
			}
			asg, ok := asgMap[asgKey{hw, bd.name}]
			if !ok {
				row.Cells = append(row.Cells, scheduleCell{IsEmpty: true, RowSpan: 1})
				continue
			}
			span := 1
			switch asg.TimeBlock {
			case "ABCD":
				if bd.name == "A" {
					span = 5
					skipCell["B|"+hw] = true
					skipCell["lunch|"+hw] = true
					skipCell["C|"+hw] = true
					skipCell["D|"+hw] = true
				}
			case "AB":
				if bd.name == "A" {
					span = 2
					skipCell["B|"+hw] = true
				}
			case "CD":
				if bd.name == "C" {
					span = 2
					skipCell["D|"+hw] = true
				}
			}
			row.Cells = append(row.Cells, scheduleCell{
				Activity: asg.ActivityName,
				Area:     asg.ProgramArea,
				Color:    asg.ProgramAreaColor,
				RowSpan:  span,
			})
		}
		rows = append(rows, row)
	}
	return rows
}

func (h *Handler) scoutScheduleView(w http.ResponseWriter, r *http.Request) {
	scoutID, _ := strconv.Atoi(r.PathValue("id"))
	scouts, _ := h.store.ListAllScouts()

	var scout models.Scout
	for _, s := range scouts {
		if s.ID == scoutID {
			scout = s
			break
		}
	}

	assignments, _ := h.store.GetScoutAssignments(scoutID)

	h.render(w, "scout_schedule.html", map[string]any{
		"Scout": scout,
		"Rows":  buildScheduleRows(assignments),
		"Page":  "scouts",
	})
}

func (h *Handler) unitScheduleView(w http.ResponseWriter, r *http.Request) {
	unitID, _ := strconv.Atoi(r.PathValue("id"))
	units, _ := h.store.ListUnits()

	var unit models.Unit
	for _, u := range units {
		if u.ID == unitID {
			unit = u
			break
		}
	}

	scouts, _ := h.store.ListScoutsByUnit(unitID)

	type ScoutGrid struct {
		Scout models.Scout
		Rows  []scheduleRow
	}
	var grids []ScoutGrid
	for _, sc := range scouts {
		asgns, _ := h.store.GetScoutAssignments(sc.ID)
		grids = append(grids, ScoutGrid{Scout: sc, Rows: buildScheduleRows(asgns)})
	}

	h.render(w, "unit_schedule.html", map[string]any{
		"Unit":  unit,
		"Grids": grids,
		"Page":  "units",
	})
}

func (h *Handler) unitRosterView(w http.ResponseWriter, r *http.Request) {
	unitID, _ := strconv.Atoi(r.PathValue("id"))
	units, _ := h.store.ListUnits()

	var unit models.Unit
	for _, u := range units {
		if u.ID == unitID {
			unit = u
			break
		}
	}

	assignments, _ := h.store.GetUnitAssignmentsWithColor(unitID)

	// Group: HalfWeek+Block -> ProgramArea -> Activity -> []scouts
	type ActivityEntry struct {
		Name   string
		Scouts []string
	}
	type AreaEntry struct {
		Name       string
		Color      string
		Activities []ActivityEntry
	}
	type SlotGroup struct {
		Label    string
		HalfWeek string
		Block    string
		Areas    []AreaEntry
	}

	// Collect scouts per slot|area|activity
	type slotKey struct{ hw, block string }
	type entryKey struct {
		hw, block, area, activity string
	}
	scoutLists := make(map[entryKey][]string)
	areaColors := make(map[string]string)

	for _, asg := range assignments {
		areaColors[asg.ProgramArea] = asg.ProgramAreaColor
		// Expand multi-block and full-week
		var blocks []string
		switch asg.TimeBlock {
		case "AB":
			blocks = []string{"AB"}
		case "CD":
			blocks = []string{"CD"}
		case "ABCD":
			blocks = []string{"ABCD"}
		default:
			blocks = []string{asg.TimeBlock}
		}
		hws := []string{asg.HalfWeek}
		if asg.HalfWeek == "full" {
			hws = []string{"first", "second"}
		}
		for _, hw := range hws {
			for _, b := range blocks {
				key := entryKey{hw, b, asg.ProgramArea, asg.ActivityName}
				scoutLists[key] = append(scoutLists[key], asg.ScoutName)
			}
		}
	}

	blockTimes := map[string]string{
		"A": "9:00 \u2013 10:30", "B": "10:30 \u2013 12:00",
		"AB": "9:00 \u2013 12:00",
		"C": "2:00 \u2013 3:30", "D": "3:30 \u2013 5:00",
		"CD": "2:00 \u2013 5:00",
		"ABCD": "9:00 \u2013 5:00",
	}
	hwLabels := map[string]string{"first": "Mon / Tue", "second": "Wed / Thu"}
	blockOrder := []string{"A", "B", "AB", "C", "D", "CD", "ABCD"}
	hwOrder := []string{"first", "second"}

	var slots []SlotGroup
	for _, hw := range hwOrder {
		for _, blk := range blockOrder {
			// Collect areas for this slot
			areasSeen := make(map[string]bool)
			var areaList []string
			for k := range scoutLists {
				if k.hw == hw && k.block == blk {
					if !areasSeen[k.area] {
						areasSeen[k.area] = true
						areaList = append(areaList, k.area)
					}
				}
			}
			if len(areaList) == 0 {
				continue
			}
			sort.Strings(areaList)
			var areas []AreaEntry
			for _, area := range areaList {
				actsSeen := make(map[string]bool)
				var actNames []string
				for k := range scoutLists {
					if k.hw == hw && k.block == blk && k.area == area && !actsSeen[k.activity] {
						actsSeen[k.activity] = true
						actNames = append(actNames, k.activity)
					}
				}
				sort.Strings(actNames)
				var acts []ActivityEntry
				for _, actName := range actNames {
					key := entryKey{hw, blk, area, actName}
					sl := scoutLists[key]
					sort.Strings(sl)
					acts = append(acts, ActivityEntry{Name: actName, Scouts: sl})
				}
				areas = append(areas, AreaEntry{Name: area, Color: areaColors[area], Activities: acts})
			}
			label := hwLabels[hw] + " / Block " + blk + " (" + blockTimes[blk] + ")"
			slots = append(slots, SlotGroup{Label: label, HalfWeek: hw, Block: blk, Areas: areas})
		}
	}

	h.render(w, "unit_roster.html", map[string]any{
		"Unit":  unit,
		"Slots": slots,
		"Page":  "units",
	})
}

// ============ Preferences ============

func (h *Handler) preferencesPage(w http.ResponseWriter, r *http.Request) {
	scoutID, _ := strconv.Atoi(r.PathValue("id"))
	scouts, _ := h.store.ListAllScouts()

	var scout models.Scout
	for _, s := range scouts {
		if s.ID == scoutID {
			scout = s
			break
		}
	}

	prefs, _ := h.store.GetPreferences(scoutID)
	activityNames, _ := h.store.ListActivityNames()

	h.render(w, "preferences.html", map[string]any{
		"Scout":         scout,
		"Preferences":   prefs,
		"ActivityNames": activityNames,
		"FillSchedule":  scout.FillSchedule,
		"Page":          "scouts",
	})
}

func (h *Handler) savePreferences(w http.ResponseWriter, r *http.Request) {
	scoutID, _ := strconv.Atoi(r.PathValue("id"))
	r.ParseForm()
	activities := r.Form["activity"]
	var prefs []models.ScoutPreference
	for i, name := range activities {
		name = strings.TrimSpace(name)
		if name != "" {
			prefs = append(prefs, models.ScoutPreference{
				ScoutID:      scoutID,
				ActivityName: name,
				Priority:     i + 1,
			})
		}
	}
	h.store.SetPreferences(scoutID, prefs)
	fillSchedule := r.FormValue("fill_schedule") == "on"
	h.store.SetFillSchedule(scoutID, fillSchedule)
	http.Redirect(w, r, "/scouts/"+strconv.Itoa(scoutID)+"/preferences", http.StatusFound)
}

// ============ Activities ============

func (h *Handler) activitiesPage(w http.ResponseWriter, r *http.Request) {
	activities, _ := h.store.ListActivities()
	programAreas, _ := h.store.ListProgramAreas()
	h.render(w, "activities.html", map[string]any{
		"Activities":   activities,
		"ProgramAreas": programAreas,
		"Page":         "activities",
	})
}

// ============ Schedule ============

func (h *Handler) schedulePage(w http.ResponseWriter, r *http.Request) {
	units, _ := h.store.ListUnits()
	unitID, _ := strconv.Atoi(r.URL.Query().Get("unit_id"))

	var assignments []models.Assignment
	if unitID > 0 {
		assignments, _ = h.store.GetAssignments(unitID)
	} else {
		assignments, _ = h.store.GetAllAssignments()
	}

	activities, _ := h.store.ListActivities()
	scouts, _ := h.store.ListAllScouts()

	h.render(w, "schedule.html", map[string]any{
		"Units":        units,
		"Assignments":  assignments,
		"Activities":   activities,
		"Scouts":       scouts,
		"SelectedUnit": unitID,
		"Page":         "schedule",
	})
}

func (h *Handler) runScheduler(w http.ResponseWriter, r *http.Request) {
	scouts, _ := h.store.ListAllScouts()
	activities, _ := h.store.ListActivities()
	existing, _ := h.store.GetAllAssignments()

	var requests []scheduler.ScoutScheduleRequest
	for _, s := range scouts {
		prefs, _ := h.store.GetPreferences(s.ID)
		if len(prefs) > 0 || s.FillSchedule {
			requests = append(requests, scheduler.ScoutScheduleRequest{
				Scout:        s,
				Preferences:  prefs,
				FillSchedule: s.FillSchedule,
			})
		}
	}

	newAssignments := scheduler.Schedule(requests, activities, existing)
	for _, asg := range newAssignments {
		h.store.CreateAssignment(asg.ScoutID, asg.ActivityID)
	}

	http.Redirect(w, r, "/schedule", http.StatusFound)
}

func (h *Handler) clearSchedule(w http.ResponseWriter, r *http.Request) {
	h.store.ClearUnlockedAssignments()
	http.Redirect(w, r, "/schedule", http.StatusFound)
}

func (h *Handler) deleteAssignment(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	h.store.DeleteAssignment(id)
	unitID := r.URL.Query().Get("unit_id")
	redirect := "/schedule"
	if unitID != "" {
		redirect += "?unit_id=" + unitID
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h *Handler) toggleLock(w http.ResponseWriter, r *http.Request) {
	id, _ := strconv.Atoi(r.PathValue("id"))
	h.store.ToggleLock(id)
	unitID := r.URL.Query().Get("unit_id")
	redirect := "/schedule"
	if unitID != "" {
		redirect += "?unit_id=" + unitID
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

func (h *Handler) addAssignment(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	scoutID, _ := strconv.Atoi(r.FormValue("scout_id"))
	activityID, _ := strconv.Atoi(r.FormValue("activity_id"))
	if scoutID > 0 && activityID > 0 {
		h.store.CreateAssignment(scoutID, activityID)
	}
	unitID := r.FormValue("unit_id")
	redirect := "/schedule"
	if unitID != "" && unitID != "0" {
		redirect += "?unit_id=" + unitID
	}
	http.Redirect(w, r, redirect, http.StatusFound)
}

// ============ API ============

func (h *Handler) apiScouts(w http.ResponseWriter, r *http.Request) {
	scouts, err := h.store.ListAllScouts()
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(scouts)
}

func (h *Handler) render(w http.ResponseWriter, name string, data map[string]any) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := h.tmpl.ExecuteTemplate(w, name, data); err != nil {
		log.Printf("Template error: %v", err)
		http.Error(w, "Template error", http.StatusInternalServerError)
	}
}
