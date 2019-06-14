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
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: zachtaylor; Tablespace: 
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO zachtaylor;

--
-- Name: users; Type: TABLE; Schema: public; Owner: zachtaylor; Tablespace: 
--

CREATE TABLE public.users (
    id character varying(255) NOT NULL,
    name character varying(255) NOT NULL,
    real_name character varying(255) NOT NULL,
    avatar character varying(255) NOT NULL,
    deleted boolean NOT NULL,
    status character varying(255) DEFAULT ''::character varying NOT NULL,
    display_name character varying(255) DEFAULT ''::character varying NOT NULL,
    created_at timestamp with time zone DEFAULT '2019-01-01 12:00:00-06'::timestamp with time zone NOT NULL,
    deleted_at timestamp with time zone
);


ALTER TABLE public.users OWNER TO zachtaylor;

--
-- Name: schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: zachtaylor; Tablespace: 
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: unique_id_on_users; Type: CONSTRAINT; Schema: public; Owner: zachtaylor; Tablespace: 
--

ALTER TABLE ONLY public.users
    ADD CONSTRAINT unique_id_on_users UNIQUE (id);


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

