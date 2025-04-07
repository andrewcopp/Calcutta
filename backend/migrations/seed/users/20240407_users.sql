--
-- PostgreSQL database dump
--

-- Dumped from database version 16.8
-- Dumped by pg_dump version 16.8 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: calcutta
--

INSERT INTO public.users (id, email, first_name, last_name, created_at, updated_at, deleted_at) VALUES ('85615af7-897d-4113-a345-14e693195a78', 'admin@calcutta.com', 'Calcutta', 'Admin', '2025-04-07 05:02:50.384966+00', '2025-04-07 05:02:50.384966+00', NULL);


--
-- PostgreSQL database dump complete
--

