package services

type RoleCreateRequest struct {
	Name string `json:"name"`
}

type RoleResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type RolesResponse struct {
	Roles []RoleResponse `json:"roles"`
}

type UsersResponse struct {
	Users []UserResponse `json:"users"`
	Total int64          `json:"total"`
}

type UpdateUserRolesRequest struct {
	Roles []string `json:"roles"`
}
