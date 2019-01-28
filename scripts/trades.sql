-- Table: public.trades

-- DROP TABLE public.trades;

CREATE TABLE public.trades
(
    id integer NOT NULL DEFAULT nextval('trades_id_seq'::regclass),
    tradepair_id integer NOT NULL,
    price numeric(15, 8) NOT NULL,
    volume numeric(15, 8) NOT NULL,
    "time" timestamp without time zone NOT NULL,
    buysell character(1) COLLATE pg_catalog."default" NOT NULL,
    marketlimit character(1) COLLATE pg_catalog."default" NOT NULL,
    CONSTRAINT trades_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.trades
    OWNER to postgres;

-- Index: ix_trades_time

-- DROP INDEX public.ix_trades_time;

CREATE INDEX ix_trades_time
    ON public.trades USING btree
    (time)
    WITH (FILLFACTOR=90)
    TABLESPACE pg_default;

-- Index: ix_trades_tradepair_id

-- DROP INDEX public.ix_trades_tradepair_id;

CREATE INDEX ix_trades_tradepair_id
    ON public.trades USING btree
    (tradepair_id)
    TABLESPACE pg_default;