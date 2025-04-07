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
-- Data for Name: tournaments; Type: TABLE DATA; Schema: public; Owner: calcutta
--

COPY public.tournaments (id, name, rounds, created_at, updated_at, deleted_at) FROM stdin;
5096aaaf-b9d4-4abf-b744-94f80effd5bb	NCAA Tournament 2016	7	2025-04-07 20:23:24.333307+00	2025-04-07 20:23:24.333307+00	\N
cb5fc573-3035-4d14-b858-56b6022a71f2	NCAA Tournament 2017	7	2025-04-07 20:23:24.437358+00	2025-04-07 20:23:24.437358+00	\N
b61b41af-b25e-4baf-9903-1517e10f32cf	NCAA Tournament 2018	7	2025-04-07 20:23:24.506586+00	2025-04-07 20:23:24.506586+00	\N
c03f62cb-9d9f-4196-a67b-291d5c811441	NCAA Tournament 2019	7	2025-04-07 20:23:24.567809+00	2025-04-07 20:23:24.567809+00	\N
5c1073e5-4c34-40b3-9651-44c40893f1dc	NCAA Tournament 2021	7	2025-04-07 20:23:24.636596+00	2025-04-07 20:23:24.636596+00	\N
d40ec459-bafc-4946-bcc0-26a2e4d73168	NCAA Tournament 2022	7	2025-04-07 20:23:24.700049+00	2025-04-07 20:23:24.700049+00	\N
9aac0e57-dd87-4808-9ac0-1491757c9bda	NCAA Tournament 2023	7	2025-04-07 20:23:24.751272+00	2025-04-07 20:23:24.751272+00	\N
b9e6f218-ca6f-4f3c-bc34-12b40804eec4	NCAA Tournament 2024	7	2025-04-07 20:23:24.817802+00	2025-04-07 20:23:24.817802+00	\N
\.


--
-- PostgreSQL database dump complete
--

