-- Table: public.tradepairs

-- DROP TABLE public.tradepairs;

CREATE TABLE public.tradepairs
(
    id integer NOT NULL DEFAULT nextval('tradepairs_id_seq'::regclass),
    cfrom text COLLATE pg_catalog."default" NOT NULL,
    cto text COLLATE pg_catalog."default" NOT NULL,
    nicename text COLLATE pg_catalog."default" NOT NULL,
    krakenname text COLLATE pg_catalog."default" NOT NULL,
    krakenlast numeric NOT NULL DEFAULT 0,
    CONSTRAINT tradepairs_pkey PRIMARY KEY (id)
)
WITH (
    OIDS = FALSE
)
TABLESPACE pg_default;

ALTER TABLE public.tradepairs
    OWNER to postgres;