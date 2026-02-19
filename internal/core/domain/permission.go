package domain

// Permission constants follow the pattern: module:resource:action
const (
	PermUserRead  = "core:user:read"
	PermUserWrite = "core:user:write"
	PermRoleRead  = "core:role:read"
	PermRoleWrite = "core:role:write"

	PermTeacherRead  = "hr:teacher:read"
	PermTeacherWrite = "hr:teacher:write"
	PermDeptRead     = "hr:department:read"
	PermDeptWrite    = "hr:department:write"

	PermSubjectRead  = "subject:subject:read"
	PermSubjectWrite = "subject:subject:write"

	PermRoomRead  = "room:room:read"
	PermRoomWrite = "room:room:write"

	PermTimetableRead  = "timetable:timetable:read"
	PermTimetableWrite = "timetable:timetable:write"

	PermAgentChat      = "agent:chat:use"
	PermAgentChatRead  = "agent:chat:read"
	PermAgentChatWrite = "agent:chat:write"
)

// AllPermissions returns every defined permission (used for admin role).
func AllPermissions() []string {
	return []string{
		PermUserRead, PermUserWrite,
		PermRoleRead, PermRoleWrite,
		PermTeacherRead, PermTeacherWrite,
		PermDeptRead, PermDeptWrite,
		PermSubjectRead, PermSubjectWrite,
		PermRoomRead, PermRoomWrite,
		PermTimetableRead, PermTimetableWrite,
		PermAgentChat, PermAgentChatRead, PermAgentChatWrite,
	}
}

// HasPermission checks if the permission set contains the required permission.
func HasPermission(perms []string, required string) bool {
	for _, p := range perms {
		if p == required {
			return true
		}
	}
	return false
}
