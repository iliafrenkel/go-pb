$Env:GOPB_API_TOKEN_SECRET="5TEdWbDmxZ2ASXcMinBYwGi66vHiU9rq"
$Env:GOPB_WEB_COOKIE_AUTH_KEY="5TEdWbDmxZ2ASXcMinBYwGi66vHiU9rq"
$Env:GOPB_WEB_LOGO="bighead.svg"
$Env:GOPB_WEB_PORT=8081
Set-Location ../cmd/
go run . --api-db-conn-string="host=localhost user=iliaf password=iliaf dbname=iliaf port=5432 sslmode=disable" --web-templates=../src/web/templates/ --web-assets=../src/web/assets/
