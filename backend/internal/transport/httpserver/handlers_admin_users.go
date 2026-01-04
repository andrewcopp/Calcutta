package httpserver

import (
	"net/http"
	"sort"
	"time"

	"github.com/gorilla/mux"
)

type adminUserListItem struct {
	ID          string    `json:"id"`
	Email       string    `json:"email"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
	Labels      []string  `json:"labels"`
	Permissions []string  `json:"permissions"`
}

type adminUsersListResponse struct {
	Items []adminUserListItem `json:"items"`
}

func (s *Server) registerAdminUsersRoutes(r *mux.Router) {
	r.HandleFunc("/api/admin/users", s.requirePermission("admin.users.read", s.adminUsersListHandler)).Methods("GET")
}

func (s *Server) adminUsersListHandler(w http.ResponseWriter, r *http.Request) {
	if s.pool == nil {
		writeError(w, r, http.StatusInternalServerError, "internal_error", "database pool not available", "")
		return
	}

	rows, err := s.pool.Query(r.Context(), `
		WITH active_grants AS (
			SELECT *
			FROM grants g
			WHERE g.revoked_at IS NULL
			  AND g.scope_type = 'global'
			  AND (g.expires_at IS NULL OR g.expires_at > NOW())
		),
		user_labels AS (
			SELECT g.user_id, l.key
			FROM active_grants g
			JOIN labels l ON g.label_id = l.id
			WHERE l.deleted_at IS NULL
		),
		user_permissions AS (
			SELECT g.user_id, p.key
			FROM active_grants g
			JOIN permissions p ON g.permission_id = p.id
			WHERE p.deleted_at IS NULL
			UNION
			SELECT g.user_id, p2.key
			FROM active_grants g
			JOIN labels l ON g.label_id = l.id AND l.deleted_at IS NULL
			JOIN label_permissions lp ON lp.label_id = l.id
			JOIN permissions p2 ON lp.permission_id = p2.id AND p2.deleted_at IS NULL
		)
		SELECT
			u.id::text,
			u.email,
			u.first_name,
			u.last_name,
			u.created_at,
			u.updated_at,
			COALESCE(array_agg(DISTINCT ul.key) FILTER (WHERE ul.key IS NOT NULL), ARRAY[]::text[]) AS labels,
			COALESCE(array_agg(DISTINCT up.key) FILTER (WHERE up.key IS NOT NULL), ARRAY[]::text[]) AS permissions
		FROM users u
		LEFT JOIN user_labels ul ON ul.user_id = u.id
		LEFT JOIN user_permissions up ON up.user_id = u.id
		WHERE u.deleted_at IS NULL
		GROUP BY u.id, u.email, u.first_name, u.last_name, u.created_at, u.updated_at
		ORDER BY u.created_at DESC
	`)
	if err != nil {
		writeErrorFromErr(w, r, err)
		return
	}
	defer rows.Close()

	items := make([]adminUserListItem, 0)
	for rows.Next() {
		var it adminUserListItem
		var labels []string
		var perms []string
		if err := rows.Scan(&it.ID, &it.Email, &it.FirstName, &it.LastName, &it.CreatedAt, &it.UpdatedAt, &labels, &perms); err != nil {
			writeErrorFromErr(w, r, err)
			return
		}

		sort.Strings(labels)
		sort.Strings(perms)
		it.Labels = labels
		it.Permissions = perms

		items = append(items, it)
	}
	if err := rows.Err(); err != nil {
		writeErrorFromErr(w, r, err)
		return
	}

	writeJSON(w, http.StatusOK, adminUsersListResponse{Items: items})
}
