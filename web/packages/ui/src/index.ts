// @mcs-erp/ui — Shared UI components (shadcn/ui + Tailwind v4)

// Utilities
export { cn } from "./lib/utils";

// shadcn primitives
export { Button, buttonVariants, type ButtonProps } from "./shadcn/button";
export { Input } from "./shadcn/input";
export { Label } from "./shadcn/label";
export {
  Dialog, DialogPortal, DialogOverlay, DialogClose, DialogTrigger,
  DialogContent, DialogHeader, DialogFooter, DialogTitle, DialogDescription,
} from "./shadcn/dialog";
export {
  Table, TableHeader, TableBody, TableFooter, TableHead, TableRow, TableCell, TableCaption,
} from "./shadcn/table";
export { Badge, badgeVariants, type BadgeProps } from "./shadcn/badge";
export {
  Card, CardHeader, CardFooter, CardTitle, CardDescription, CardContent,
} from "./shadcn/card";
export {
  Select, SelectGroup, SelectValue, SelectTrigger, SelectContent,
  SelectLabel, SelectItem, SelectSeparator,
} from "./shadcn/select";
export {
  DropdownMenu, DropdownMenuTrigger, DropdownMenuContent, DropdownMenuItem,
  DropdownMenuCheckboxItem, DropdownMenuRadioItem, DropdownMenuLabel,
  DropdownMenuSeparator, DropdownMenuShortcut, DropdownMenuGroup,
  DropdownMenuSub, DropdownMenuSubContent, DropdownMenuSubTrigger,
} from "./shadcn/dropdown-menu";
export { Separator } from "./shadcn/separator";
export { Tooltip, TooltipTrigger, TooltipContent, TooltipProvider } from "./shadcn/tooltip";
export { Checkbox } from "./shadcn/checkbox";
export { ScrollArea, ScrollBar } from "./shadcn/scroll-area";
export { Tabs, TabsList, TabsTrigger, TabsContent } from "./shadcn/tabs";

// Application components
export { DataTable } from "./components/data-table";
export { AvailabilityGrid } from "./components/availability-grid";
export { FormDialog } from "./components/form-dialog";
export { ConfirmDialog } from "./components/confirm-dialog";
export { LoadingSpinner } from "./components/loading-spinner";
export { EmptyState } from "./components/empty-state";
