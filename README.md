# duolingo-case-study

cd src
go run infra/database/campaign_db/migrate/migrate.go up
go run infra/database/campaign_db/seed/seed.go --campaign test_1000_usr --total 1000

go run util/load_simulator/load_simulator.go --campaign test_10K_usr --num 1

docker compose -p duolingo_case_study up -d --build
