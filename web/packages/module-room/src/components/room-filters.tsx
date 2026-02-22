// Filter bar for room list: building, min capacity, equipment search.
import { Input, Label } from "@mcs-erp/ui";
import type { RoomFilter } from "@mcs-erp/api-client";

interface RoomFiltersProps {
  filters: RoomFilter;
  onChange: (filters: RoomFilter) => void;
}

export function RoomFilters({ filters, onChange }: RoomFiltersProps) {
  return (
    <div className="flex flex-wrap gap-4 mb-4">
      <div className="space-y-1">
        <Label htmlFor="filter-building">Building</Label>
        <Input
          id="filter-building"
          placeholder="e.g. A"
          value={filters.building ?? ""}
          onChange={(e) => onChange({ ...filters, building: e.target.value || undefined })}
          className="w-40"
        />
      </div>
      <div className="space-y-1">
        <Label htmlFor="filter-capacity">Min Capacity</Label>
        <Input
          id="filter-capacity"
          type="number"
          placeholder="e.g. 30"
          value={filters.min_capacity ?? ""}
          onChange={(e) =>
            onChange({ ...filters, min_capacity: e.target.value ? Number(e.target.value) : undefined })
          }
          className="w-32"
        />
      </div>
      <div className="space-y-1">
        <Label htmlFor="filter-equipment">Equipment</Label>
        <Input
          id="filter-equipment"
          placeholder="e.g. projector"
          value={filters.equipment ?? ""}
          onChange={(e) => onChange({ ...filters, equipment: e.target.value || undefined })}
          className="w-40"
        />
      </div>
    </div>
  );
}
