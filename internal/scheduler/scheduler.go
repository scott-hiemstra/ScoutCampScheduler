package scheduler

import (
	"math/rand"
	"sort"

	"summer-camp-scheduler/internal/models"
)

// BlocksOverlap returns true if two time blocks occupy overlapping times.
func BlocksOverlap(block1, block2 string) bool {
	if block1 == block2 {
		return true
	}
	overlaps := map[string][]string{
		"A":    {"AB", "ABCD"},
		"B":    {"AB", "ABCD"},
		"AB":   {"A", "B", "ABCD"},
		"C":    {"CD", "ABCD"},
		"D":    {"CD", "ABCD"},
		"CD":   {"C", "D", "ABCD"},
		"ABCD": {"A", "B", "AB", "C", "D", "CD"},
	}
	for _, o := range overlaps[block1] {
		if o == block2 {
			return true
		}
	}
	return false
}

// HalfWeeksConflict returns true if two half-week values cannot coexist in the same block.
func HalfWeeksConflict(hw1, hw2 string) bool {
	if hw1 == "full" || hw2 == "full" {
		return true
	}
	return hw1 == hw2
}

// ScoutScheduleRequest bundles a scout and their preferences for scheduling.
type ScoutScheduleRequest struct {
	Scout        models.Scout
	Preferences  []models.ScoutPreference
	FillSchedule bool
}

// Schedule generates non-conflicting assignments for scouts based on preferences.
// It respects capacity, avoids time conflicts, and tries to buddy-up scouts from
// the same unit into the same program area at the same time block.
func Schedule(
	requests []ScoutScheduleRequest,
	activities []models.Activity,
	existingAssignments []models.Assignment,
) []models.Assignment {

	actByID := make(map[int]models.Activity)
	for _, a := range activities {
		actByID[a.ID] = a
	}

	actByName := make(map[string][]models.Activity)
	for _, a := range activities {
		actByName[a.Name] = append(actByName[a.Name], a)
	}

	enrollment := make(map[int]int)
	for _, asg := range existingAssignments {
		enrollment[asg.ActivityID]++
	}

	scoutAssignments := make(map[int][]models.Assignment)
	for _, asg := range existingAssignments {
		scoutAssignments[asg.ScoutID] = append(scoutAssignments[asg.ScoutID], asg)
	}

	unitScouts := make(map[int][]int)
	for _, req := range requests {
		unitScouts[req.Scout.UnitID] = append(unitScouts[req.Scout.UnitID], req.Scout.ID)
	}

	// Program-area + block tracking for buddy matching
	areaBlockScouts := make(map[string]map[int]bool)
	for _, asg := range existingAssignments {
		act := actByID[asg.ActivityID]
		key := act.ProgramAreaName + "|" + act.TimeBlockName + "|" + act.HalfWeek
		if areaBlockScouts[key] == nil {
			areaBlockScouts[key] = make(map[int]bool)
		}
		areaBlockScouts[key][asg.ScoutID] = true
	}

	// Most constrained scouts first
	sort.Slice(requests, func(i, j int) bool {
		return len(requests[i].Preferences) < len(requests[j].Preferences)
	})

	var newAssignments []models.Assignment

	for _, req := range requests {
		scout := req.Scout

		assignedNames := make(map[string]bool)
		for _, asg := range scoutAssignments[scout.ID] {
			assignedNames[actByID[asg.ActivityID].Name] = true
		}

		for _, pref := range req.Preferences {
			if assignedNames[pref.ActivityName] {
				continue
			}

			sessions := actByName[pref.ActivityName]
			if len(sessions) == 0 {
				continue
			}

			var candidates []models.Activity
			for _, sess := range sessions {
				if enrollment[sess.ID] >= sess.Capacity {
					continue
				}
				if conflicts(sess, scoutAssignments[scout.ID], actByID) {
					continue
				}
				candidates = append(candidates, sess)
			}
			if len(candidates) == 0 {
				continue
			}

			// Score candidates — prefer sessions where a unit-mate is in the same area/block
			best := candidates[0]
			bestScore := -1
			for _, cand := range candidates {
				score := 0
				key := cand.ProgramAreaName + "|" + cand.TimeBlockName + "|" + cand.HalfWeek
				if scouts, ok := areaBlockScouts[key]; ok {
					for _, mateID := range unitScouts[scout.UnitID] {
						if mateID != scout.ID && scouts[mateID] {
							score += 10
						}
					}
				}
				if cand.HalfWeek != "full" {
					keyFull := cand.ProgramAreaName + "|" + cand.TimeBlockName + "|full"
					if scouts, ok := areaBlockScouts[keyFull]; ok {
						for _, mateID := range unitScouts[scout.UnitID] {
							if mateID != scout.ID && scouts[mateID] {
								score += 10
							}
						}
					}
				}
				if score > bestScore {
					bestScore = score
					best = cand
				}
			}

			asg := models.Assignment{
				ScoutID:      scout.ID,
				ActivityID:   best.ID,
				ScoutName:    scout.Name,
				ActivityName: best.Name,
				ProgramArea:  best.ProgramAreaName,
				TimeBlock:    best.TimeBlockName,
				HalfWeek:     best.HalfWeek,
			}
			newAssignments = append(newAssignments, asg)
			scoutAssignments[scout.ID] = append(scoutAssignments[scout.ID], asg)
			enrollment[best.ID]++
			assignedNames[pref.ActivityName] = true

			key := best.ProgramAreaName + "|" + best.TimeBlockName + "|" + best.HalfWeek
			if areaBlockScouts[key] == nil {
				areaBlockScouts[key] = make(map[int]bool)
			}
			areaBlockScouts[key][scout.ID] = true
		}
	}

	// Second pass: fill remaining time blocks for scouts with FillSchedule=true
	// Use only single-block activities (A, B, C, D) to avoid over-committing
	fillBlocks := []string{"A", "B", "C", "D"}

	for _, req := range requests {
		if !req.FillSchedule {
			continue
		}
		scout := req.Scout

		// Track activity names already assigned to avoid duplicates
		fillAssignedNames := make(map[string]bool)
		for _, asg := range scoutAssignments[scout.ID] {
			fillAssignedNames[actByID[asg.ActivityID].Name] = true
		}

		for _, block := range fillBlocks {
			// Check both half-weeks for this block
			for _, hw := range []string{"first", "second"} {
				// See if this scout already has something in this block+half
				taken := false
				for _, asg := range scoutAssignments[scout.ID] {
					act := actByID[asg.ActivityID]
					if BlocksOverlap(block, act.TimeBlockName) && HalfWeeksConflict(hw, act.HalfWeek) {
						taken = true
						break
					}
				}
				if taken {
					continue
				}

				// Find all activities in this block+half with capacity, no prereqs
				var candidates []models.Activity
				for _, act := range activities {
					if act.TimeBlockName != block {
						continue
					}
					if act.HalfWeek != hw && act.HalfWeek != "full" {
						continue
					}
					if fillAssignedNames[act.Name] {
						continue
					}
					if act.HasPrerequisites {
						continue
					}
					if enrollment[act.ID] >= act.Capacity {
						continue
					}
					if conflicts(act, scoutAssignments[scout.ID], actByID) {
						continue
					}
					candidates = append(candidates, act)
				}
				if len(candidates) == 0 {
					continue
				}

				// Score by buddy, pick best (random tiebreak)
				rand.Shuffle(len(candidates), func(i, j int) {
					candidates[i], candidates[j] = candidates[j], candidates[i]
				})

				best := candidates[0]
				bestScore := -1
				for _, cand := range candidates {
					score := 0
					key := cand.ProgramAreaName + "|" + cand.TimeBlockName + "|" + cand.HalfWeek
					if scouts, ok := areaBlockScouts[key]; ok {
						for _, mateID := range unitScouts[scout.UnitID] {
							if mateID != scout.ID && scouts[mateID] {
								score += 10
							}
						}
					}
					if score > bestScore {
						bestScore = score
						best = cand
					}
				}

				asg := models.Assignment{
					ScoutID:      scout.ID,
					ActivityID:   best.ID,
					ScoutName:    scout.Name,
					ActivityName: best.Name,
					ProgramArea:  best.ProgramAreaName,
					TimeBlock:    best.TimeBlockName,
					HalfWeek:     best.HalfWeek,
				}
				newAssignments = append(newAssignments, asg)
				scoutAssignments[scout.ID] = append(scoutAssignments[scout.ID], asg)
				enrollment[best.ID]++
				fillAssignedNames[best.Name] = true

				key := best.ProgramAreaName + "|" + best.TimeBlockName + "|" + best.HalfWeek
				if areaBlockScouts[key] == nil {
					areaBlockScouts[key] = make(map[int]bool)
				}
				areaBlockScouts[key][scout.ID] = true
			}
		}
	}

	return newAssignments
}

func conflicts(proposed models.Activity, existing []models.Assignment, actByID map[int]models.Activity) bool {
	for _, asg := range existing {
		act := actByID[asg.ActivityID]
		if BlocksOverlap(proposed.TimeBlockName, act.TimeBlockName) &&
			HalfWeeksConflict(proposed.HalfWeek, act.HalfWeek) {
			return true
		}
	}
	return false
}
