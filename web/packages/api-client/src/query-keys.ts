// Centralized query key factory for TanStack Query cache management.

export const queryKeys = {
  // HR
  teachers: {
    all: ["teachers"] as const,
    list: (filters?: object) => ["teachers", "list", filters] as const,
    detail: (id: string) => ["teachers", id] as const,
    availability: (id: string) => ["teachers", id, "availability"] as const,
  },
  departments: {
    all: ["departments"] as const,
    list: () => ["departments", "list"] as const,
    detail: (id: string) => ["departments", id] as const,
  },
  // Subject
  subjects: {
    all: ["subjects"] as const,
    list: (filters?: object) => ["subjects", "list", filters] as const,
    detail: (id: string) => ["subjects", id] as const,
    prerequisites: (id: string) => ["subjects", id, "prerequisites"] as const,
    prerequisiteChain: (id: string) => ["subjects", id, "prerequisite-chain"] as const,
  },
  categories: {
    all: ["categories"] as const,
    list: () => ["categories", "list"] as const,
  },
  // Room
  rooms: {
    all: ["rooms"] as const,
    list: (filters?: object) => ["rooms", "list", filters] as const,
    detail: (id: string) => ["rooms", id] as const,
    availability: (id: string) => ["rooms", id, "availability"] as const,
  },
  // Timetable
  semesters: {
    all: ["semesters"] as const,
    list: () => ["semesters", "list"] as const,
    detail: (id: string) => ["semesters", id] as const,
    subjects: (id: string) => ["semesters", id, "subjects"] as const,
    schedule: (id: string) => ["semesters", id, "schedule"] as const,
  },
  // Agent
  conversations: {
    all: ["conversations"] as const,
    list: () => ["conversations", "list"] as const,
    detail: (id: string) => ["conversations", id] as const,
  },
  suggestions: {
    all: ["suggestions"] as const,
  },
} as const;
