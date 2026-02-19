import { rootRoute } from "./routes/__root";
import { loginRoute } from "./routes/login";
import { authenticatedRoute } from "./routes/_authenticated";
import { dashboardRoute } from "./routes/_authenticated/index";

const routeTree = rootRoute.addChildren([
  loginRoute,
  authenticatedRoute.addChildren([dashboardRoute]),
]);

export { routeTree };
