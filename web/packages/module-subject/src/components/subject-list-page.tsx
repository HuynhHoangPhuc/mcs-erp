// Subject list page with search, category filter, pagination, and CRUD actions.
import { useState, useMemo } from "react";
import { useNavigate } from "@tanstack/react-router";
import { Button, Input, DataTable, LoadingSpinner } from "@mcs-erp/ui";
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from "@mcs-erp/ui";
import { useSubjects, useCategories, type Subject } from "@mcs-erp/api-client";
import { buildSubjectColumns } from "./subject-columns";
import { SubjectFormDialog } from "./subject-form-dialog";
import { Search } from "lucide-react";

const PAGE_SIZE = 20;

export function SubjectListPage() {
  const navigate = useNavigate();
  const [search, setSearch] = useState("");
  const [categoryId, setCategoryId] = useState<string>("all");
  const [offset, setOffset] = useState(0);
  const [formOpen, setFormOpen] = useState(false);
  const [editTarget, setEditTarget] = useState<Subject | undefined>();

  const { data: subjects, isLoading } = useSubjects({
    search: search || undefined,
    category_id: categoryId !== "all" ? categoryId : undefined,
    offset,
    limit: PAGE_SIZE,
  });

  const { data: categoriesData } = useCategories();

  const categoryMap = useMemo(() => {
    const map: Record<string, string> = {};
    for (const cat of categoriesData?.items ?? []) {
      map[cat.id] = cat.name;
    }
    return map;
  }, [categoriesData]);

  const columns = useMemo(
    () =>
      buildSubjectColumns({
        categoryMap,
        onEdit: (subject) => {
          setEditTarget(subject);
          setFormOpen(true);
        },
        onViewDetail: (subject) => {
          navigate({ to: "/subjects/$subjectId", params: { subjectId: subject.id } });
        },
      }),
    [categoryMap, navigate]
  );

  function handleAddNew() {
    setEditTarget(undefined);
    setFormOpen(true);
  }

  return (
    <div className="space-y-6 p-6">
      <div className="flex items-center justify-between">
        <div>
          <h1 className="text-2xl font-bold tracking-tight">Subjects</h1>
          <p className="text-muted-foreground text-sm">Manage academic subjects and their details.</p>
        </div>
        <Button onClick={handleAddNew}>Add Subject</Button>
      </div>

      <div className="flex items-center gap-3">
        <div className="relative flex-1 max-w-sm">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder="Search subjects..."
            value={search}
            onChange={(e) => { setSearch(e.target.value); setOffset(0); }}
            className="pl-8"
          />
        </div>
        <Select value={categoryId} onValueChange={(v) => { setCategoryId(v); setOffset(0); }}>
          <SelectTrigger className="w-48">
            <SelectValue placeholder="All categories" />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">All categories</SelectItem>
            {categoriesData?.items.map((cat) => (
              <SelectItem key={cat.id} value={cat.id}>{cat.name}</SelectItem>
            ))}
          </SelectContent>
        </Select>
      </div>

      {isLoading ? (
        <LoadingSpinner />
      ) : (
        <DataTable
          columns={columns}
          data={subjects?.items ?? []}
          total={subjects?.total ?? 0}
          offset={offset}
          limit={PAGE_SIZE}
          onPaginationChange={setOffset}
        />
      )}

      <SubjectFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        initialData={editTarget}
      />
    </div>
  );
}
