CREATE DATABASE test
    WITH OWNER postgres;
CREATE TABLE jobs
(
    id        serial
        CONSTRAINT jobs_pk
            PRIMARY KEY,
    due_at    timestamp DEFAULT NOW() NOT NULL,
    priority  integer   DEFAULT 0,
    tenant_id integer   DEFAULT 1,
    status    integer   DEFAULT 0,
    metadata  varchar(100)
);

ALTER TABLE jobs
    OWNER TO postgres;

CREATE TABLE processing_queue
(
    id        integer NOT NULL,
    due_at    timestamp DEFAULT NOW(),
    priority  integer   DEFAULT 0,
    tenant_id integer   DEFAULT 1,
    status    integer   DEFAULT 0
);

ALTER TABLE processing_queue
    OWNER TO postgres;

