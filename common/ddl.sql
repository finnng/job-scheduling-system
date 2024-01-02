CREATE TABLE IF NOT EXISTS public.jobs
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

ALTER TABLE public.jobs
    OWNER TO postgres;

CREATE INDEX IF NOT EXISTS jobs_due_at_index
    ON public.jobs (due_at);

CREATE INDEX IF NOT EXISTS jobs_priority_index
    ON public.jobs (priority);

CREATE INDEX IF NOT EXISTS jobs_status_index
    ON public.jobs (status);

