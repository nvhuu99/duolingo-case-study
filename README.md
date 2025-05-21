# duolingo-case-study

Setup docker:
    docker build -t duolingo-service -f ./dockerfile.dev ..
    docker swarm init
    docker stack deploy --detach=true -c docker-compose.yml duolingo_case_study

Database Migration:
    cd src
    go run infra/database/campaign_db/migrate/migrate.go up
    go run infra/database/campaign_db/seed/seed.go --campaign test_1000_usr --total 1000

Load 10 request:
    go run util/load_simulator/load_simulator.go --campaign test_100K_usr --num 1
