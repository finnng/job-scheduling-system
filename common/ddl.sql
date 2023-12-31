create table public.jobs
(
    id        serial
        constraint jobs_pk
            primary key,
    due_at    timestamp default now() not null,
    priority  integer   default 0,
    tenant_id integer   default 1,
    status    integer   default 0,
    metadata  varchar(100)
);

alter table public.jobs
    owner to postgres;

create table public.processing_queue
(
    id        integer not null,
    due_at    timestamp default now(),
    priority  integer   default 0,
    tenant_id integer   default 1,
    status    integer   default 0
);

alter table public.processing_queue
    owner to postgres;

