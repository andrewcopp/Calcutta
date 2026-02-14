--
-- PostgreSQL database dump
--

\restrict wTwcnR6iXeeiyDTcUa5ghIYdYLsz4scUARumutwNxtvlm2ty14gITxC9sDEavb6

-- Dumped from database version 16.12
-- Dumped by pg_dump version 16.12

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
-- Data for Name: labels; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.labels (id, key, description, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'global_admin', 'All permissions (admin)', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.labels (id, key, description, created_at, updated_at, deleted_at) VALUES ('2623b7e3-4dcf-4378-a091-a2a33efe560c', 'tournament_operator', 'Tournament operations', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.labels (id, key, description, created_at, updated_at, deleted_at) VALUES ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', 'calcutta_owner', 'Manage a specific calcutta', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.labels (id, key, description, created_at, updated_at, deleted_at) VALUES ('fcbaa74f-823e-4015-8680-522f2e210fef', 'player', 'Participate in a calcutta', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);


--
-- Data for Name: permissions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('f8be010c-b161-4d9b-a8f2-268c08be5e43', 'tournament.game.write', 'Update tournament game results and winners', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('a9c96ef0-d7e7-4b35-9d68-7bef3b89c9c1', 'calcutta.config.write', 'Create and configure a calcutta', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('8ea2f503-a867-4f53-a6ee-a7113d60dc12', 'calcutta.invite.write', 'Invite and manage calcutta members', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb', 'calcutta.read', 'Read calcutta data', '2025-12-25 04:30:39.156889+00', '2025-12-25 04:30:39.156889+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('06f9f5af-0285-4852-9430-106aca75100c', 'admin.analytics.export', 'Export analytics snapshots', '2025-12-26 04:15:28.437239+00', '2025-12-26 04:15:28.437239+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('5a1ebaa1-368e-4c4e-a4aa-d12a3f9ebc21', 'admin.api_keys.write', 'Create and revoke API keys', '2025-12-26 04:15:28.486486+00', '2025-12-26 04:15:28.486486+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('20d84b63-ec46-4f20-8afa-9e5b9397f171', 'admin.bundles.export', 'Export bundles', '2025-12-26 04:15:28.52836+00', '2025-12-26 04:15:28.52836+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('3b29ff2e-a969-4dfe-848e-47de12a762cc', 'admin.bundles.import', 'Import bundles', '2025-12-26 04:15:28.52836+00', '2025-12-26 04:15:28.52836+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('10cca9d3-32eb-4ef8-ba34-2a398ea42069', 'admin.bundles.read', 'Read bundle upload status', '2025-12-26 04:15:28.52836+00', '2025-12-26 04:15:28.52836+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('47cd8bba-46e6-4959-8b86-72eaed4100be', 'admin.analytics.read', 'Read analytics', '2025-12-26 04:15:28.539826+00', '2025-12-26 04:15:28.539826+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('a3a1a1cc-8a4d-4813-911a-601675cd25f6', 'admin.hof.read', 'Read hall of fame', '2025-12-26 04:15:28.539826+00', '2025-12-26 04:15:28.539826+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('bcaed262-7b2f-4366-a724-f970f64d5a8f', 'analytics.entry_evaluation_requests.write', 'Submit entry evaluation requests', '2026-01-04 04:24:04.16856+00', '2026-01-04 04:24:04.16856+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('e2f46c07-a0b9-43b8-a5b2-118c32da4276', 'analytics.entry_evaluation_requests.read', 'Read entry evaluation requests', '2026-01-04 04:24:04.178227+00', '2026-01-04 04:24:04.178227+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('4d045f02-704b-475c-bb6a-5a88d742a07c', 'admin.users.read', 'Read users', '2026-01-04 05:47:25.791678+00', '2026-01-04 05:47:25.791678+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('caeb9563-7298-4ef3-b9d2-4c13529baa20', 'admin.users.write', 'Create/update users', '2026-01-04 22:27:51.951635+00', '2026-01-04 22:27:51.951635+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('fd6335f2-3954-422a-9b4e-50b48009efe7', 'analytics.suite_calcutta_evaluations.write', 'Submit suite calcutta evaluation requests', '2026-01-05 06:23:42.883084+00', '2026-01-05 06:23:42.883084+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('3880e398-f561-400a-84ad-4a41f3d7488d', 'analytics.suite_calcutta_evaluations.read', 'Read suite calcutta evaluation requests', '2026-01-05 06:23:42.883084+00', '2026-01-05 06:23:42.883084+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('04d72de1-b6ca-4111-b731-575419c5cd60', 'analytics.suite_executions.write', 'Create suite executions (batch runs across calcuttas)', '2026-01-05 17:11:39.2582+00', '2026-01-05 17:11:39.2582+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('2e8f6939-2ec4-461e-9a60-725bb376a909', 'analytics.suite_executions.read', 'Read suite executions (batch runs across calcuttas)', '2026-01-05 17:11:39.2582+00', '2026-01-05 17:11:39.2582+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('28df6d71-bf3e-44a8-b3aa-caa4a5be444c', 'analytics.suites.write', 'Create/update suites', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('eb55e249-6edd-42eb-8fc3-41b3920de2fa', 'analytics.suites.read', 'Read suites', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('1e366126-b681-4ff2-ac1a-248981752fdf', 'analytics.suite_scenarios.write', 'Create/update suite scenarios', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('e9d4d29e-2eb0-4c9f-87db-26ec5871d59f', 'analytics.suite_scenarios.read', 'Read suite scenarios', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('9819c1be-5287-413d-ad24-3f25664d1dfd', 'analytics.strategy_generation_runs.write', 'Create strategy generation runs', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('46febb06-bea5-4187-ba85-046db1ac6cd1', 'analytics.strategy_generation_runs.read', 'Read strategy generation runs', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.permissions (id, key, description, created_at, updated_at, deleted_at) VALUES ('e5d05903-3acb-4759-b3ed-ab900beffdad', 'analytics.run_jobs.read', 'Read run job progress and status', '2026-01-08 17:14:59.342502+00', '2026-01-08 17:14:59.342502+00', NULL);


--
-- Data for Name: users; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.users (id, email, first_name, last_name, created_at, updated_at, deleted_at, password_hash, status, invite_token_hash, invite_expires_at, invite_consumed_at, invited_at, last_invite_sent_at) VALUES ('4cfab92e-9231-46db-8aad-17d5e377a704', 'andrew.w.copp@gmail.com', 'Andrew', 'Copp', '2025-12-26 04:39:08.674737+00', '2025-12-26 04:39:08.674737+00', NULL, '$2a$10$Gj.cTG82saTJZu/gBUnldOpECv3S7cqxDyiUqasJkWH0qJ1756z5q', 'active', NULL, NULL, NULL, NULL, NULL);
INSERT INTO public.users (id, email, first_name, last_name, created_at, updated_at, deleted_at, password_hash, status, invite_token_hash, invite_expires_at, invite_consumed_at, invited_at, last_invite_sent_at) VALUES ('c55b1803-4838-442b-adc0-9c5cb074b99f', 'admin@calcutta.com', 'Calcutta', 'Admin', '2025-12-25 04:31:00.787973+00', '2026-02-13 19:22:51.93725+00', NULL, '$2a$10$XTovP7eJ/pfs.f/IAkJZnOdRb410QFBp5Lb/eH1bXJfA4DQVDF2v2', 'active', NULL, NULL, NULL, NULL, NULL);


--
-- Data for Name: grants; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.grants (id, user_id, scope_type, scope_id, label_id, permission_id, created_at, updated_at, expires_at, revoked_at, deleted_at) VALUES ('d8d22fb0-9298-42f4-9c07-70b44fba52f8', '4cfab92e-9231-46db-8aad-17d5e377a704', 'global', NULL, '7fd3956d-9df0-4c1b-b176-e7b8b6d01248', NULL, '2025-12-26 04:39:08.687043+00', '2025-12-26 04:39:08.687043+00', NULL, NULL, NULL);
INSERT INTO public.grants (id, user_id, scope_type, scope_id, label_id, permission_id, created_at, updated_at, expires_at, revoked_at, deleted_at) VALUES ('c2c94095-88b1-4045-b818-a278794487a3', 'c55b1803-4838-442b-adc0-9c5cb074b99f', 'global', NULL, NULL, 'eb55e249-6edd-42eb-8fc3-41b3920de2fa', '2026-02-13 19:23:28.441926+00', '2026-02-13 19:23:28.441926+00', NULL, NULL, NULL);
INSERT INTO public.grants (id, user_id, scope_type, scope_id, label_id, permission_id, created_at, updated_at, expires_at, revoked_at, deleted_at) VALUES ('006fc5b6-aac7-483a-92e2-61cb59e5a3b9', 'c55b1803-4838-442b-adc0-9c5cb074b99f', 'global', NULL, NULL, '28df6d71-bf3e-44a8-b3aa-caa4a5be444c', '2026-02-13 19:23:28.441926+00', '2026-02-13 19:23:28.441926+00', NULL, NULL, NULL);


--
-- Data for Name: label_permissions; Type: TABLE DATA; Schema: public; Owner: -
--

INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'f8be010c-b161-4d9b-a8f2-268c08be5e43', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'a9c96ef0-d7e7-4b35-9d68-7bef3b89c9c1', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '8ea2f503-a867-4f53-a6ee-a7113d60dc12', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('2623b7e3-4dcf-4378-a091-a2a33efe560c', 'f8be010c-b161-4d9b-a8f2-268c08be5e43', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('2623b7e3-4dcf-4378-a091-a2a33efe560c', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', 'a9c96ef0-d7e7-4b35-9d68-7bef3b89c9c1', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', '8ea2f503-a867-4f53-a6ee-a7113d60dc12', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('49b7d3d1-e45f-49b7-9bdc-424959d0c7ab', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('fcbaa74f-823e-4015-8680-522f2e210fef', '074eb6a9-c3ab-4f27-9a0d-fa0779bb7bcb', '2025-12-25 04:30:39.156889+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '06f9f5af-0285-4852-9430-106aca75100c', '2025-12-26 04:15:28.437239+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '5a1ebaa1-368e-4c4e-a4aa-d12a3f9ebc21', '2025-12-26 04:15:28.486486+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '20d84b63-ec46-4f20-8afa-9e5b9397f171', '2025-12-26 04:15:28.52836+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '3b29ff2e-a969-4dfe-848e-47de12a762cc', '2025-12-26 04:15:28.52836+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '10cca9d3-32eb-4ef8-ba34-2a398ea42069', '2025-12-26 04:15:28.52836+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '47cd8bba-46e6-4959-8b86-72eaed4100be', '2025-12-26 04:15:28.539826+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'a3a1a1cc-8a4d-4813-911a-601675cd25f6', '2025-12-26 04:15:28.539826+00', '2026-01-02 16:35:13.01656+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'bcaed262-7b2f-4366-a724-f970f64d5a8f', '2026-01-04 04:24:04.16856+00', '2026-01-04 04:24:04.16856+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'e2f46c07-a0b9-43b8-a5b2-118c32da4276', '2026-01-04 04:24:04.178227+00', '2026-01-04 04:24:04.178227+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '4d045f02-704b-475c-bb6a-5a88d742a07c', '2026-01-04 05:47:25.791678+00', '2026-01-04 05:47:25.791678+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'caeb9563-7298-4ef3-b9d2-4c13529baa20', '2026-01-04 22:27:51.951635+00', '2026-01-04 22:27:51.951635+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'fd6335f2-3954-422a-9b4e-50b48009efe7', '2026-01-05 06:23:42.883084+00', '2026-01-05 06:23:42.883084+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '3880e398-f561-400a-84ad-4a41f3d7488d', '2026-01-05 06:23:42.883084+00', '2026-01-05 06:23:42.883084+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '04d72de1-b6ca-4111-b731-575419c5cd60', '2026-01-05 17:11:39.2582+00', '2026-01-05 17:11:39.2582+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '2e8f6939-2ec4-461e-9a60-725bb376a909', '2026-01-05 17:11:39.2582+00', '2026-01-05 17:11:39.2582+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '28df6d71-bf3e-44a8-b3aa-caa4a5be444c', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'eb55e249-6edd-42eb-8fc3-41b3920de2fa', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '1e366126-b681-4ff2-ac1a-248981752fdf', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'e9d4d29e-2eb0-4c9f-87db-26ec5871d59f', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '9819c1be-5287-413d-ad24-3f25664d1dfd', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', '46febb06-bea5-4187-ba85-046db1ac6cd1', '2026-01-05 22:58:02.473296+00', '2026-01-05 22:58:02.473296+00', NULL);
INSERT INTO public.label_permissions (label_id, permission_id, created_at, updated_at, deleted_at) VALUES ('7fd3956d-9df0-4c1b-b176-e7b8b6d01248', 'e5d05903-3acb-4759-b3ed-ab900beffdad', '2026-01-08 17:14:59.342502+00', '2026-01-08 17:14:59.342502+00', NULL);


--
-- PostgreSQL database dump complete
--

\unrestrict wTwcnR6iXeeiyDTcUa5ghIYdYLsz4scUARumutwNxtvlm2ty14gITxC9sDEavb6

