interface KPICardProps {
  title: string;
  value: string | number;
  suffix?: string;
  trend?: "up" | "down" | "neutral";
}

export function KPICard({ title, value, suffix }: KPICardProps) {
  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-5">
      <p className="text-sm text-gray-500 mb-1">{title}</p>
      <p className="text-2xl font-bold text-gray-900">
        {value}
        {suffix && <span className="text-sm font-normal text-gray-500 ml-1">{suffix}</span>}
      </p>
    </div>
  );
}
