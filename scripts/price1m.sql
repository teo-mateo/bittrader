-- Table: public.price1m

-- DROP TABLE public.price1m;

CREATE TABLE public.price1m
(
    id integer NOT NULL DEFAULT nextval('price_id_seq'::regclass),
    tradepair_id integer NOT NULL,
    "time" timestamp without time zone NOT NULL,
    price numeric(15, 8) NOT NULL,
    imagined boolean NOT NULL DEFAULT false,
    CONSTRAINT price1m_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.price1m
    OWNER to postgres;

-- Index: ix_price1m_time

-- DROP INDEX public.ix_price1m_time;

CREATE INDEX ix_price1m_time
    ON public.price1m USING btree
    (time)
    WITH (FILLFACTOR=90)
    TABLESPACE pg_default;