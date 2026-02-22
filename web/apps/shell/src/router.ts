import { rootRoute } from "./routes/__root";
import { loginRoute } from "./routes/login";
import { authenticatedRoute } from "./routes/_authenticated";
import { dashboardRoute } from "./routes/_authenticated/index";
import { teachersRoute } from "./routes/_authenticated/teachers.index";
import { teacherDetailRoute } from "./routes/_authenticated/teachers.$teacherId";
import { departmentsRoute } from "./routes/_authenticated/departments.index";
import { subjectsRoute } from "./routes/_authenticated/subjects.index";
import { subjectDetailRoute } from "./routes/_authenticated/subjects.$subjectId";
import { prerequisitesRoute } from "./routes/_authenticated/subjects.prerequisites";
import { categoriesRoute } from "./routes/_authenticated/categories.index";
import { roomsRoute } from "./routes/_authenticated/rooms.index";
import { roomDetailRoute } from "./routes/_authenticated/rooms.$roomId";
import { timetableRoute } from "./routes/_authenticated/timetable.index";
import { semesterDetailRoute } from "./routes/_authenticated/timetable.$semesterId";
import { chatRoute } from "./routes/_authenticated/chat.index";

const routeTree = rootRoute.addChildren([
  loginRoute,
  authenticatedRoute.addChildren([
    dashboardRoute,
    teachersRoute,
    teacherDetailRoute,
    departmentsRoute,
    subjectsRoute,
    subjectDetailRoute,
    prerequisitesRoute,
    categoriesRoute,
    roomsRoute,
    roomDetailRoute,
    timetableRoute,
    semesterDetailRoute,
    chatRoute,
  ]),
]);

export { routeTree };
