#!/bin/sh
export GOPB_API_TOKEN_SECRET="5TEdWbDmxZ2ASXcMinBYwGi66vHiU9rq"
export GOPB_WEB_COOKIE_AUTH_KEY="5TEdWbDmxZ2ASXcMinBYwGi66vHiU9rq"
export GOPB_WEB_LOGO="bighead.svg"
cd ../cmd
go run . --api-db-conn-string=test.db --web-templates=../src/web/templates/ --web-assets=../src/web/assets/