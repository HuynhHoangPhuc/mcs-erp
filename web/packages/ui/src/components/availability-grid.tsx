import { Checkbox } from "../shadcn/checkbox";

const DAYS = ["Mon", "Tue", "Wed", "Thu", "Fri", "Sat", "Sun"];
const PERIODS = Array.from({ length: 10 }, (_, i) => i + 1);

interface AvailabilitySlot {
  day: number;
  period: number;
  is_available: boolean;
}

interface AvailabilityGridProps {
  slots: AvailabilitySlot[];
  onChange: (slots: AvailabilitySlot[]) => void;
  disabled?: boolean;
}

// 7-day x 10-period checkbox matrix for teacher/room availability.
export function AvailabilityGrid({ slots, onChange, disabled }: AvailabilityGridProps) {
  const slotMap = new Map<string, boolean>();
  for (const s of slots) {
    slotMap.set(`${s.day}-${s.period}`, s.is_available);
  }

  const toggle = (day: number, period: number) => {
    const key = `${day}-${period}`;
    const current = slotMap.get(key) ?? false;
    const updated = slots.map((s) =>
      s.day === day && s.period === period ? { ...s, is_available: !current } : s
    );
    // If slot doesn't exist yet, add it
    if (!slots.some((s) => s.day === day && s.period === period)) {
      updated.push({ day, period, is_available: true });
    }
    onChange(updated);
  };

  return (
    <div className="overflow-x-auto">
      <table className="border-collapse text-sm">
        <thead>
          <tr>
            <th className="p-2 text-left text-muted-foreground">Period</th>
            {DAYS.map((day, i) => (
              <th key={i} className="p-2 text-center font-medium">{day}</th>
            ))}
          </tr>
        </thead>
        <tbody>
          {PERIODS.map((period) => (
            <tr key={period}>
              <td className="p-2 text-muted-foreground">{period}</td>
              {DAYS.map((_, day) => (
                <td key={day} className="p-2 text-center">
                  <Checkbox
                    checked={slotMap.get(`${day}-${period}`) ?? false}
                    onCheckedChange={() => toggle(day, period)}
                    disabled={disabled}
                  />
                </td>
              ))}
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
