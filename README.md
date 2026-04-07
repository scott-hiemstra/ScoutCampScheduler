# Summer Camp Scheduler

A web application for automatically scheduling Scouts BSA summer camp merit badge activities based on scout preferences, with buddy system support and capacity management.

Built as a proof-of-concept for the **2026 Scouts BSA Activity Schedule**.

## The Problem

[Scoutingevent.com](https://scoutingevent.com) does a great job handling summer camp registration — it's a fully functional platform that councils and camps rely on, and it works. But when it comes to merit badge scheduling, the current model puts unit leaders in a tough spot: everyone registers for activity slots **at the same time**, creating a free-for-all. Leaders are fighting each other for limited spots, refreshing pages, and scrambling to piece together workable schedules for their scouts.

For larger units this is especially painful — a leader with 30+ scouts has to manually juggle dozens of preferences, time conflicts, and capacity limits while competing with every other unit for the same openings.

There has to be a better way.

## A Better Approach

Instead of making unit leaders fight for slots in real time, this application explores a **preference-based auto-scheduling** model:

1. **Scouts submit ranked preferences** — "I want Swimming 1st, Rifle Shooting 2nd, Robotics 3rd"
2. **An algorithm assigns everyone at once** — respecting capacity limits, time conflicts, and preference priority
3. **A buddy system** keeps scouts from the same unit together at the same program area and time block
4. **Unit leaders review and adjust** — lock good assignments, re-run for gaps, or make manual overrides

This eliminates the competitive scramble, produces fairer results, and saves unit leaders hours of manual scheduling work.

> **This is a proof of concept, not a replacement for Scoutingevent.com.** It doesn't aim to replicate the full functionality of that platform — registration, payments, roster management, and everything else Scoutingevent handles well. This project focuses on one thing: demonstrating that preference-based auto-scheduling could be a better way to handle merit badge sign-ups, and proposing it as a feature for platforms like Scoutingevent.

## Features

- **Auto-Scheduling** — Constraint-satisfaction algorithm assigns scouts to activities based on ranked preferences
- **Buddy System** — Scouts from the same unit are grouped together at the same program area and time block when possible
- **Fill Schedule** — Optional per-scout setting to automatically fill empty time blocks with random activities
- **Visual Schedules** — Color-coded weekly grid view for individual scouts and full unit overviews
- **Unit Roster Report** — At-a-glance report showing where every scout is during each time block
- **Manual Overrides** — Lock assignments, manually assign/remove activities, or clear and re-run
- **Full Activity Catalog** — Pre-loaded with all program areas and activities from the 2026 schedule (Aquatics, Eco-STEM, Handicraft, Scout Scholar, Scoutcraft, COPE, Rifle & Target, TFC, Trek)

## Quick Start

### Prerequisites

- [Docker](https://www.docker.com/get-started) and Docker Compose

### Launch

```bash
git clone https://github.com/scott-hiemstra/ScoutCampScheduler.git
cd ScoutCampScheduler
docker compose up --build
```

The app will be available at **http://localhost:8080**.

On first launch, the database is automatically initialized with:
- All program areas, time blocks, and activities from the 2026 schedule
- An empty set of units/scouts ready for you to populate

### Load Sample Data

To load 150 test scouts across 5 troops with random preferences:

```bash
bash seed_test_data.sh
```

### Stop

```bash
docker compose down
```

To also clear the database (start fresh):

```bash
docker compose down -v
```

## Documentation

| Document | Description |
|----------|-------------|
| [User Guide](USER_GUIDE.md) | Step-by-step instructions for unit leaders |
| [Admin Guide](ADMIN_GUIDE.md) | Setup, configuration, and maintenance |

## Tech Stack

- **Go 1.26** — stdlib `net/http` with method-based routing (no web framework)
- **MySQL 8.0** — persistent storage via Docker volume
- **Docker Compose** — single-command deployment
- **Only 2 dependencies** — `go-sql-driver/mysql` and its indirect dependency

## Project Structure

```
ScoutCampScheduler/
├── cmd/server/main.go          # Entry point, DB connection, HTTP server
├── internal/
│   ├── database/db.go          # MySQL data access layer
│   ├── handlers/handlers.go    # HTTP request handlers
│   ├── models/models.go        # Data structures
│   └── scheduler/scheduler.go  # Auto-scheduling algorithm
├── web/
│   ├── templates/              # HTML templates
│   └── static/style.css        # Stylesheet
├── init.sql                    # Database schema + seed data
├── docker-compose.yml          # Docker services
├── Dockerfile                  # Multi-stage Go build
└── seed_test_data.sh           # Sample data loader
```

## Contributing

This is an open-source proof-of-concept. Contributions, feedback, and ideas are welcome — please open an issue or pull request.

If you're involved with Scouting America or camp administration and interested in this approach, we'd love to hear from you.

## License

MIT License — see [LICENSE](LICENSE) for details.
