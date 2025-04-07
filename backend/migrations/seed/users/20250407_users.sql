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

COPY public.users (id, email, first_name, last_name, created_at, updated_at, deleted_at) FROM stdin;
090644de-1158-402e-a103-949b089d8cf9	admin@calcutta.com	Calcutta	Admin	2025-04-07 19:58:24.269075+00	2025-04-07 19:58:24.269075+00	\N
\.


--
-- PostgreSQL database dump complete
--

