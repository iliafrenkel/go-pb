CREATE USER test WITH PASSWORD 'test';
CREATE DATABASE test;
GRANT ALL PRIVILEGES ON DATABASE test TO test;

DROP TABLE IF EXISTS "pastes";
DROP SEQUENCE IF EXISTS pastes_id_seq;
CREATE SEQUENCE pastes_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."pastes" (
    "id" bigint DEFAULT nextval('pastes_id_seq') NOT NULL,
    "title" text,
    "body" text,
    "expires" timestamptz,
    "delete_after_read" boolean,
    "password" text,
    "created" timestamptz,
    "syntax" text,
    "user_id" bigint,
    CONSTRAINT "pastes_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

CREATE INDEX "idx_pastes_expires" ON "public"."pastes" USING btree ("expires");

CREATE INDEX "idx_pastes_user_id" ON "public"."pastes" USING btree ("user_id");


DROP TABLE IF EXISTS "users";
DROP SEQUENCE IF EXISTS users_id_seq;
CREATE SEQUENCE users_id_seq INCREMENT 1 MINVALUE 1 MAXVALUE 9223372036854775807 CACHE 1;

CREATE TABLE "public"."users" (
    "id" bigint DEFAULT nextval('users_id_seq') NOT NULL,
    "username" text,
    "email" text,
    "password_hash" text,
    "created_at" timestamptz,
    CONSTRAINT "users_pkey" PRIMARY KEY ("id")
) WITH (oids = false);

CREATE INDEX "idx_users_email" ON "public"."users" USING btree ("email");

CREATE INDEX "idx_users_username" ON "public"."users" USING btree ("username");


ALTER TABLE ONLY "public"."pastes" ADD CONSTRAINT "fk_pastes_user" FOREIGN KEY (user_id) REFERENCES users(id) NOT DEFERRABLE;

