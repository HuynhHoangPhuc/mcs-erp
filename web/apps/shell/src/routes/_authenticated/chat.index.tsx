import { createRoute } from "@tanstack/react-router";
import { authenticatedRoute } from "../_authenticated";
import { ChatPage } from "@mcs-erp/module-agent";

export const chatRoute = createRoute({
  getParentRoute: () => authenticatedRoute,
  path: "/chat",
  component: ChatPage,
});
