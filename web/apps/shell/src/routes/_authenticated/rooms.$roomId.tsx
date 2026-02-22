import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { RoomDetailPage } from "@mcs-erp/module-room";

export const roomDetailRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/rooms/$roomId",
  component: RoomDetailPage,
});
