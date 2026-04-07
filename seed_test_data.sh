#!/bin/bash
# Seed script: loads test scouts with random preferences and fill_schedule flags.
# Usage: ./seed_test_data.sh
# Requires: docker compose running with the camp DB accessible.

set -e

MYSQL="docker compose exec -T db mysql -uroot -pcampscheduler camp"

# Activities without prerequisites
ACTIVITIES=(
  "Archery"
  "Artificial Intelligence"
  "Astronomy"
  "Basketry"
  "C.O.P.E."
  "Canoeing"
  "Chemistry"
  "Chess"
  "Climbing"
  "Cybersecurity"
  "Engineering"
  "Fingerprinting"
  "Fire Safety"
  "Fishing"
  "Forestry"
  "Game Design"
  "Geocaching"
  "Geology"
  "Graphic Arts"
  "Insect Study"
  "Kayaking"
  "Law"
  "Leatherwork"
  "Mammal Study"
  "Metalwork"
  "Motorboating + Rowing"
  "Music"
  "Nature"
  "Orienteering"
  "Pioneering"
  "Pulp and Paper"
  "Reptile and Amphibian"
  "Rifle Shooting"
  "Sculpture"
  "Search and Rescue"
  "Shotgun Shooting"
  "Small Boat Sailing"
  "Soil & Water Conservation"
  "Space Exploration"
  "Trail to First Class"
  "Watersports"
  "Wilderness Survival"
  "Woodcarving"
)

GIRL_FIRST_NAMES=(
  "Emma" "Olivia" "Ava" "Sophia" "Isabella"
  "Mia" "Charlotte" "Amelia" "Harper" "Evelyn"
  "Abigail" "Emily" "Ella" "Elizabeth" "Camila"
  "Luna" "Sofia" "Avery" "Mila" "Aria"
  "Scarlett" "Penelope" "Layla" "Chloe" "Victoria"
  "Madison" "Eleanor" "Grace" "Nora" "Riley"
  "Zoey" "Hannah"
)

BOY_FIRST_NAMES=(
  "Liam" "Noah" "Oliver" "James" "Elijah"
  "William" "Henry" "Lucas" "Benjamin" "Theodore"
  "Jack" "Levi" "Alexander" "Mason" "Ethan"
  "Daniel" "Jacob" "Michael" "Logan" "Jackson"
  "Sebastian" "Aiden" "Owen" "Samuel" "Ryan"
  "Nathan" "Carter" "Luke" "Dylan" "Caleb"
  "Isaac" "Connor" "Hunter" "Wyatt" "Jayden"
  "Asher" "Leo" "Thomas" "Adrian" "Joshua"
  "Christopher" "Andrew" "Lincoln" "Mateo" "Ezra"
)

LAST_NAMES=(
  "Smith" "Johnson" "Williams" "Brown" "Jones"
  "Garcia" "Miller" "Davis" "Rodriguez" "Martinez"
  "Hernandez" "Lopez" "Gonzalez" "Wilson" "Anderson"
  "Thomas" "Taylor" "Moore" "Jackson" "Martin"
  "Lee" "Perez" "Thompson" "White" "Harris"
  "Sanchez" "Clark" "Ramirez" "Lewis" "Robinson"
  "Walker" "Young" "Allen" "King" "Wright"
  "Scott" "Torres" "Nguyen" "Hill" "Flores"
  "Green" "Adams" "Nelson" "Baker" "Hall"
  "Rivera" "Campbell" "Mitchell" "Carter" "Roberts"
)

NUM_ACTIVITIES=${#ACTIVITIES[@]}

# Pick N random unique activities for a scout (3-6 preferences)
pick_activities() {
  local count=$(( (RANDOM % 4) + 3 ))  # 3 to 6
  local indices=()
  while [[ ${#indices[@]} -lt $count ]]; do
    local idx=$(( RANDOM % NUM_ACTIVITIES ))
    local dup=0
    for i in "${indices[@]}"; do
      if [[ $i -eq $idx ]]; then dup=1; break; fi
    done
    if [[ $dup -eq 0 ]]; then
      indices+=("$idx")
    fi
  done
  for i in "${indices[@]}"; do
    echo "${ACTIVITIES[$i]}"
  done
}

# Insert a unit and return its ID
insert_unit() {
  local unit_name="$1"
  $MYSQL -N -e "INSERT INTO units (name) VALUES ('$unit_name') ON DUPLICATE KEY UPDATE id=LAST_INSERT_ID(id); SELECT LAST_INSERT_ID();" 2>/dev/null | tail -1
}

# Seed scouts into a unit
seed_troop() {
  local unit_name="$1"
  local count="$2"
  local gender="$3"  # "girl" or "boy"

  echo "Creating unit: $unit_name ($count $gender scouts)..."
  local unit_id
  unit_id=$(insert_unit "$unit_name")
  unit_id=$(echo "$unit_id" | tr -d '[:space:]')

  local -a names
  if [[ "$gender" == "girl" ]]; then
    names=("${GIRL_FIRST_NAMES[@]}")
  else
    names=("${BOY_FIRST_NAMES[@]}")
  fi

  local sql=""
  local pref_sql=""
  local fill_sql=""

  for ((i=0; i<count; i++)); do
    local first="${names[$((i % ${#names[@]}))]}"
    local last="${LAST_NAMES[$((RANDOM % ${#LAST_NAMES[@]}))]}"
    local scout_name="${first} ${last}"

    # ~40% chance of fill_schedule
    local fill=0
    if (( RANDOM % 10 < 4 )); then
      fill=1
    fi

    # Insert scout
    sql="INSERT INTO scouts (name, unit_id, fill_schedule) VALUES ('${scout_name}', ${unit_id}, ${fill});"
    local scout_id
    scout_id=$($MYSQL -N -e "${sql} SELECT LAST_INSERT_ID();" 2>/dev/null | tail -1)
    scout_id=$(echo "$scout_id" | tr -d '[:space:]')

    # Pick random preferences
    local priority=1
    local prefs
    prefs=$(pick_activities)
    while IFS= read -r act_name; do
      pref_sql+="INSERT INTO scout_preferences (scout_id, activity_name, priority) VALUES (${scout_id}, '${act_name}', ${priority});"
      priority=$((priority + 1))
    done <<< "$prefs"

    echo "  Added: ${scout_name} (fill=${fill}, $(echo "$prefs" | wc -l) prefs)"
  done

  # Batch insert all preferences
  if [[ -n "$pref_sql" ]]; then
    echo "$pref_sql" | $MYSQL 2>/dev/null
  fi
}

echo "=== Seeding Test Data ==="
echo ""

echo "Clearing existing data..."
$MYSQL -e "DELETE FROM assignments; DELETE FROM scout_preferences; DELETE FROM scouts; DELETE FROM units;" 2>/dev/null
echo "Done."
echo ""

seed_troop "Troop 9123" 20 "girl"
echo ""
seed_troop "Troop 123"  50 "boy"
echo ""
seed_troop "Troop 9999" 22 "girl"
echo ""
seed_troop "Troop 999"  38 "boy"
echo ""
seed_troop "Troop 42"   20 "boy"

echo ""
echo "=== Done! ==="
echo "Total: 150 scouts across 5 troops"
echo "Visit http://localhost:8080 to view and run the scheduler."
