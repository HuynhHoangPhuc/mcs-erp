import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { RoomListPage } from "@mcs-erp/module-room";

export const roomsRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/rooms",
  component: RoomListPage,
});
