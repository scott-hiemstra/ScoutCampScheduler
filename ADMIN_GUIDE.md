# Admin Guide — Summer Camp Scheduler

This guide covers setup, configuration, database management, and troubleshooting for administrators deploying the Summer Camp Scheduler.

---

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Configuration](#configuration)
4. [Database Management](#database-management)
5. [Seed Script](#seed-script)
6. [Updating Activities](#updating-activities)
7. [Backup and Restore](#backup-and-restore)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

- **Docker** and **Docker Compose** — [Install Docker](https://docs.docker.com/get-docker/)
- A modern web browser
- Terminal / command line access

No Go installation is needed — the application is built inside a Docker container.

---

## Installation

```bash
git clone <repository-url>
cd SummerCampScheduler
docker compose up --build
```

The first launch will:
1. Build the Go binary in a multi-stage Docker build
2. Start a MySQL 8.0 container
3. Run `init.sql` to create the schema and load the 2026 activity catalog
4. Start the web server on port **8080**

The app waits for the database to be ready (retries up to 30 times with 2-second intervals).

---

## Configuration

### Docker Compose Settings

All configuration is in `docker-compose.yml`:

| Setting | Default | Description |
|---------|---------|-------------|
| `MYSQL_ROOT_PASSWORD` | `campscheduler` | Database root password |
| `MYSQL_DATABASE` | `camp_scheduler` | Database name |
| App port | `8080:8080` | Host:container port mapping |
| DB port | `3306:3306` | Exposed for external tools (optional) |

### Changing the Port

To run on a different port, edit the `ports` mapping in `docker-compose.yml`:

```yaml
app:
  ports:
    - "3000:8080"  # Access at http://localhost:3000
```

### Database Password

To change the database password, update it in **both** the `db` and `app` service:

```yaml
db:
  environment:
    MYSQL_ROOT_PASSWORD: your-new-password

app:
  environment:
    DB_DSN: root:your-new-password@tcp(db:3306)/camp_scheduler?parseTime=true
```

---

## Database Management

### Accessing MySQL Directly

While the containers are running:

```bash
docker compose exec db mysql -uroot -pcampscheduler camp_scheduler
```

### Schema Overview

| Table | Purpose |
|-------|---------|
| `units` | Scout units (troops, packs) |
| `scouts` | Individual scouts, linked to a unit |
| `program_areas` | Categories (Aquatics, Eco-STEM, etc.) with colors |
| `time_blocks` | Schedule slots (A, B, AB, C, D, CD, ABCD) |
| `activities` | Merit badges / sessions with capacity and prerequisites |
| `scout_preferences` | Ranked activity choices per scout |
| `assignments` | Final schedule assignments (scout → activity, with lock flag) |

### Resetting the Database

To completely wipe and reinitialize:

```bash
docker compose down -v
docker compose up --build
```

The `-v` flag removes the MySQL data volume, so `init.sql` runs again on next startup.

---

## Seed Script

The included `seed_test_data.sh` creates sample data for testing:

- **150 scouts** across 5 units (Troop 9123, 123, 9999, 999, 42)
- **3–6 random preferences** per scout
- **~40% of scouts** have "Fill Schedule" enabled

### Running the Seed Script

```bash
bash seed_test_data.sh
```

The script connects to MySQL at `localhost:3306` and:
1. Clears all existing assignments, preferences, scouts, and units
2. Inserts the test units and scouts
3. Assigns random activity preferences

### Customizing the Seed Script

Edit `seed_test_data.sh` to change unit names, scout counts, or preference distributions. The script uses `mysql` CLI — make sure the MySQL container is running and port 3306 is exposed.

---

## Updating Activities

Activity data is loaded from `init.sql` on first database initialization. To update activities for a new camp year:

1. Edit the `INSERT INTO activities` section of `init.sql`
2. Update program areas and time blocks if the camp schedule changes
3. Reset the database:
   ```bash
   docker compose down -v
   docker compose up --build
   ```

Alternatively, connect to MySQL directly and run INSERT/UPDATE statements without resetting.

### Activity Fields

| Field | Description |
|-------|-------------|
| `name` | Activity / merit badge name |
| `program_area_id` | Reference to program_areas table |
| `time_block_id` | Reference to time_blocks table |
| `half_week` | `first` (Mon/Tue), `second` (Wed/Thu), or `full` (Mon–Thu) |
| `capacity` | Maximum enrollees |
| `has_prerequisites` | `TRUE` if prior completion is required (excluded from fill schedule) |

---

## Backup and Restore

### Backup

```bash
docker compose exec db mysqldump -uroot -pcampscheduler camp_scheduler > backup.sql
```

### Restore

```bash
docker compose exec -T db mysql -uroot -pcampscheduler camp_scheduler < backup.sql
```

### Backup Just Assignments

To export only the schedule (useful for sharing with camp staff):

```bash
docker compose exec db mysqldump -uroot -pcampscheduler camp_scheduler assignments > assignments_backup.sql
```

---

## Troubleshooting

### App won't start / "connection refused"

The app retries the database connection 30 times. If it still fails:

```bash
# Check if the database container is healthy
docker compose ps

# View database logs
docker compose logs db

# View app logs
docker compose logs app
```

Common causes:
- Port 3306 already in use by a local MySQL installation
- Docker not running or out of disk space

### "Table already exists" on startup

This is normal — `init.sql` uses `CREATE TABLE IF NOT EXISTS` and the warnings are harmless. Data is preserved across restarts as long as you don't use `docker compose down -v`.

### Scheduler produces unexpected results

- Verify scouts have saved preferences (check the Preferences page)
- Check activity capacity — full activities can't accept more scouts
- Multi-block activities (AB, CD, ABCD) block the component single blocks (A, B, C, D) for that half-week
- The scheduler processes most-constrained scouts first (fewest preferences)

### Port already in use

```bash
# Find what's using port 8080
lsof -i :8080

# Or change the port in docker-compose.yml
```

### Resetting a single scout's schedule

From the **Schedule** page, filter by unit, then delete individual assignments. Or from the Scout's individual **Schedule** view, identify which assignments to remove.

### Database connection from external tools

The MySQL database is exposed on `localhost:3306`:

- **Host**: localhost
- **Port**: 3306
- **User**: root
- **Password**: campscheduler
- **Database**: camp_scheduler
