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
820d2909-c280-4986-b9c5-c55e1e13c1e5	admin@calcutta.com	Calcutta	Admin	2025-04-07 07:44:06.212807+00	2025-04-07 07:44:06.212807+00	\N
\.


--
-- PostgreSQL database dump complete
--

