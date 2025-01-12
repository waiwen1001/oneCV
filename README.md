Step 1: Setup PostgreSQL into your machine
Recommended : Docker
Using docker : 
Step 1: docker pull postgres
Step 2: docker run --name onecv_db -e POSTGRES_USER=onecv_user -e POSTGRES_PASSWORD=onecv_pw -e POSTGRES_DB=onecv -p 5432:5432 -d postgres
Step 3: 
(linux/macOS)
export GOOSE_DRIVER="postgres"
export GOOSE_DBSTRING="postgresql://onecv_user:onecv_pw@localhost:5432/onecv"
export GOOSE_MIGRATION_DIR="./migrations"
(window)
$env:GOOSE_DRIVER = "postgres"
$env:GOOSE_DBSTRING = "postgresql://onecv_user:onecv_pw@localhost:5432/onecv"
$env:GOOSE_MIGRATION_DIR = "./migrations"

