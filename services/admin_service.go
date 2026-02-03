package services

import (
	"errors"
	"strconv"
	"strings"

	"github.com/google/uuid"
	"github.com/jkaninda/goma-admin/config"
	"github.com/jkaninda/goma-admin/models"
	"github.com/jkaninda/goma-admin/store"
	"github.com/jkaninda/okapi"
	"gorm.io/gorm"
)

type AdminService struct {
	users store.UserStore
	roles store.RoleStore
}

func NewAdminService(conf *config.Config) *AdminService {
	return &AdminService{
		users: store.NewUserStore(conf.Database.DB),
		roles: store.NewRoleStore(conf.Database.DB),
	}
}

func (s *AdminService) ListUsers(c *okapi.Context) error {
	limit := parseLimit(c.Query("limit"), 50, 200)
	offset := parseOffset(c.Query("offset"))

	users, err := s.users.List(c.Request().Context(), limit, offset)
	if err != nil {
		return c.AbortInternalServerError("Failed to list users", err)
	}
	total, err := s.users.Count(c.Request().Context())
	if err != nil {
		return c.AbortInternalServerError("Failed to count users", err)
	}
	response := make([]UserResponse, 0, len(users))
	for _, user := range users {
		response = append(response, UserResponse{
			ID:    user.ID.String(),
			Email: user.Email,
			Name:  user.Name,
			Roles: extractRoleNames(&user),
		})
	}
	return c.OK(UsersResponse{Users: response, Total: total})
}

func (s *AdminService) GetUser(c *okapi.Context) error {
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return c.AbortBadRequest("Invalid user id", err)
	}
	user, err := s.users.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.AbortNotFound("User not found", err)
		}
		return c.AbortInternalServerError("Failed to load user", err)
	}
	return c.OK(UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Roles: extractRoleNames(user),
	})
}

func (s *AdminService) UpdateUserRoles(c *okapi.Context) error {
	id, err := parseUUIDParam(c.Param("id"))
	if err != nil {
		return c.AbortBadRequest("Invalid user id", err)
	}
	var req UpdateUserRolesRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}
	roleNames := normalizeRoleNames(req.Roles)
	if len(roleNames) == 0 {
		return c.AbortBadRequest("Roles are required")
	}

	roles, err := s.roles.FindByNames(c.Request().Context(), roleNames)
	if err != nil {
		return c.AbortInternalServerError("Failed to load roles", err)
	}
	if len(roles) != len(roleNames) {
		return c.AbortBadRequest("Unknown role supplied")
	}

	user, err := s.users.FindByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return c.AbortNotFound("User not found", err)
		}
		return c.AbortInternalServerError("Failed to load user", err)
	}
	if err := s.users.ReplaceRoles(c.Request().Context(), user, roles); err != nil {
		return c.AbortInternalServerError("Failed to update roles", err)
	}
	user.Roles = roles
	return c.OK(UserResponse{
		ID:    user.ID.String(),
		Email: user.Email,
		Name:  user.Name,
		Roles: extractRoleNames(user),
	})
}

func (s *AdminService) ListRoles(c *okapi.Context) error {
	limit := parseLimit(c.Query("limit"), 100, 500)
	offset := parseOffset(c.Query("offset"))
	roles, err := s.roles.List(c.Request().Context(), limit, offset)
	if err != nil {
		return c.AbortInternalServerError("Failed to list roles", err)
	}
	response := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		response = append(response, RoleResponse{ID: role.ID.String(), Name: role.Name})
	}
	return c.OK(RolesResponse{Roles: response})
}

func (s *AdminService) CreateRole(c *okapi.Context) error {
	var req RoleCreateRequest
	if err := c.Bind(&req); err != nil {
		return c.AbortBadRequest("Invalid request", err)
	}
	name := strings.ToLower(strings.TrimSpace(req.Name))
	if name == "" {
		return c.AbortBadRequest("Role name is required")
	}
	if _, err := s.roles.FindByName(c.Request().Context(), name); err == nil {
		return c.AbortConflict("Role already exists")
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		return c.AbortInternalServerError("Failed to check role", err)
	}

	role := &models.Role{Name: name}
	if err := s.roles.Create(c.Request().Context(), role); err != nil {
		return c.AbortInternalServerError("Failed to create role", err)
	}
	return c.Created(RoleResponse{ID: role.ID.String(), Name: role.Name})
}

func parseLimit(raw string, fallback, max int) int {
	if raw == "" {
		return fallback
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value <= 0 {
		return fallback
	}
	if value > max {
		return max
	}
	return value
}

func parseOffset(raw string) int {
	if raw == "" {
		return 0
	}
	value, err := strconv.Atoi(raw)
	if err != nil || value < 0 {
		return 0
	}
	return value
}

func parseUUIDParam(raw string) (uuid.UUID, error) {
	return uuid.Parse(strings.TrimSpace(raw))
}

func normalizeRoleNames(values []string) []string {
	seen := map[string]struct{}{}
	var normalized []string
	for _, value := range values {
		name := strings.ToLower(strings.TrimSpace(value))
		if name == "" {
			continue
		}
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}
		normalized = append(normalized, name)
	}
	return normalized
}
