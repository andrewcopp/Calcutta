SET search_path = '';

CREATE TABLE core.user_merges (
    id uuid DEFAULT public.uuid_generate_v4() NOT NULL,
    source_user_id uuid NOT NULL,
    target_user_id uuid NOT NULL,
    merged_by uuid NOT NULL,
    entries_moved int NOT NULL DEFAULT 0,
    invitations_moved int NOT NULL DEFAULT 0,
    grants_moved int NOT NULL DEFAULT 0,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT pk_user_merges PRIMARY KEY (id),
    CONSTRAINT fk_user_merges_source FOREIGN KEY (source_user_id) REFERENCES core.users(id),
    CONSTRAINT fk_user_merges_target FOREIGN KEY (target_user_id) REFERENCES core.users(id),
    CONSTRAINT fk_user_merges_merged_by FOREIGN KEY (merged_by) REFERENCES core.users(id),
    CONSTRAINT ck_user_merges_different CHECK (source_user_id != target_user_id)
);

CREATE INDEX idx_user_merges_source ON core.user_merges (source_user_id);
CREATE INDEX idx_user_merges_target ON core.user_merges (target_user_id);
