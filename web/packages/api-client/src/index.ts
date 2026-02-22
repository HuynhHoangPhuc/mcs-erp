// API client package - fetch wrapper, types, and TanStack Query hooks.

// Core fetch wrapper
export { apiFetch, ApiError, setAccessToken, getAccessToken } from "./lib/api-client";

// Query keys
export { queryKeys } from "./query-keys";

// Types
export type * from "./types/common";
export type * from "./types/auth";
export type * from "./types/hr";
export type * from "./types/subject";
export type * from "./types/room";
export type * from "./types/timetable";
export type * from "./types/agent";

// Hooks - HR
export { useTeachers, useTeacher, useCreateTeacher, useUpdateTeacher, useTeacherAvailability, useUpdateTeacherAvailability } from "./hooks/use-teachers";
export { useDepartments, useCreateDepartment, useUpdateDepartment, useDeleteDepartment } from "./hooks/use-departments";

// Hooks - Subject
export { useSubjects, useSubject, useCreateSubject, useUpdateSubject, useSubjectPrerequisites, useSubjectPrerequisiteChain, useAddPrerequisite, useDeletePrerequisite } from "./hooks/use-subjects";
export { useCategories, useCreateCategory, useUpdateCategory, useDeleteCategory } from "./hooks/use-categories";

// Hooks - Room
export { useRooms, useRoom, useCreateRoom, useUpdateRoom, useRoomAvailability, useUpdateRoomAvailability } from "./hooks/use-rooms";

// Hooks - Timetable
export { useSemesters, useSemester, useCreateSemester, useSemesterSubjects, useSetSemesterSubjects, useAssignTeacher } from "./hooks/use-semesters";
export { useSchedule, useGenerateSchedule, useApproveSemester, useUpdateAssignment } from "./hooks/use-schedule";

// Hooks - Agent
export { useConversations, useConversation, useUpdateConversation, useDeleteConversation, useSuggestions } from "./hooks/use-conversations";
export { useChatSSE } from "./hooks/use-chat-sse";
