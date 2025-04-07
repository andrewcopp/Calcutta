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
b8c5c616-2b70-4556-afe0-54a7b13861e8	NCAA Tournament 2017	7	2025-04-07 07:43:43.599306+00	2025-04-07 07:43:43.599306+00	\N
01bf0205-55a6-4141-8ba4-b6e37fbef0ea	NCAA Tournament 2018	7	2025-04-07 07:43:43.692617+00	2025-04-07 07:43:43.692617+00	\N
f19e2861-c8de-4b9f-8e34-d6c330e41524	NCAA Tournament 2019	7	2025-04-07 07:43:43.767195+00	2025-04-07 07:43:43.767195+00	\N
7ce7dfcf-3271-4f6d-83c9-fcc3f6d2f8e0	NCAA Tournament 2021	7	2025-04-07 07:43:43.833266+00	2025-04-07 07:43:43.833266+00	\N
486bc23e-fd96-4158-a3b2-48b9ebebae17	NCAA Tournament 2022	7	2025-04-07 07:43:43.903588+00	2025-04-07 07:43:43.903588+00	\N
7c271759-94d6-4b7e-826f-16bdc4c7e45b	NCAA Tournament 2023	7	2025-04-07 07:43:43.966821+00	2025-04-07 07:43:43.966821+00	\N
44ae878f-a9c7-413e-957b-86ea84ec0b46	NCAA Tournament 2024	7	2025-04-07 07:43:44.031383+00	2025-04-07 07:43:44.031383+00	\N
8fe595d7-02cd-4a57-8455-5bfc56537b72	NCAA Tournament 2025	7	2025-04-07 07:43:44.102183+00	2025-04-07 07:43:44.102183+00	\N
\.


--
-- PostgreSQL database dump complete
--

