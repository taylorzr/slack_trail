--
-- PostgreSQL database dump
--

SET statement_timeout = 0;
SET lock_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET client_min_messages = warning;

--
-- Name: plpgsql; Type: EXTENSION; Schema: -; Owner: 
--

CREATE EXTENSION IF NOT EXISTS plpgsql WITH SCHEMA pg_catalog;


--
-- Name: EXTENSION plpgsql; Type: COMMENT; Schema: -; Owner: 
--

COMMENT ON EXTENSION plpgsql IS 'PL/pgSQL procedural language';


SET default_tablespace = '';

SET default_with_oids = false;

--
-- Name: emojis; Type: TABLE; Schema: public; Owner: zachtaylor; Tablespace: 
--

CREATE TABLE public.emojis (
    name character varying(255) NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);


ALTER TABLE public.emojis OWNER TO zachtaylor;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: zachtaylor; Tablespace: 
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO zachtaylor;

--
-- Name: schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: zachtaylor; Tablespace: 
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: unique_name_on_emojis; Type: CONSTRAINT; Schema: public; Owner: zachtaylor; Tablespace: 
--

ALTER TABLE ONLY public.emojis
    ADD CONSTRAINT unique_name_on_emojis UNIQUE (name);


--
-- Name: SCHEMA public; Type: ACL; Schema: -; Owner: zachtaylor
--

REVOKE ALL ON SCHEMA public FROM PUBLIC;
REVOKE ALL ON SCHEMA public FROM zachtaylor;
GRANT ALL ON SCHEMA public TO zachtaylor;
GRANT ALL ON SCHEMA public TO PUBLIC;


--
-- PostgreSQL database dump complete
--

