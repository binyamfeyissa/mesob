"use client";

import { LineChart, Line, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from "recharts";
import { useQuery } from "@tanstack/react-query";
import { apiFetch } from "@/lib/api/client";

interface DataPoint {
  date: string;
  volume_minor: number;
  count: number;
}

export function VolumeChart() {
  const { data } = useQuery({
    queryKey: ["volume-chart"],
    queryFn: () => apiFetch<{ data: DataPoint[] }>("/admin/dashboard/volume"),
    refetchInterval: 60_000,
  });

  const chartData = (data?.data ?? []).map((d) => ({
    ...d,
    volume_etb: d.volume_minor / 100,
  }));

  return (
    <div className="bg-white rounded-xl shadow-sm border border-gray-100 p-5">
      <h2 className="text-sm font-semibold text-gray-700 mb-4">7-Day Transaction Volume (ETB)</h2>
      <ResponsiveContainer width="100%" height={200}>
        <LineChart data={chartData}>
          <CartesianGrid strokeDasharray="3 3" stroke="#F3F4F6" />
          <XAxis dataKey="date" tick={{ fontSize: 11 }} />
          <YAxis tick={{ fontSize: 11 }} />
          <Tooltip
            formatter={(value: number) => [`${value.toFixed(2)} ETB`, "Volume"]}
          />
          <Line
            type="monotone"
            dataKey="volume_etb"
            stroke="#1B4FDE"
            strokeWidth={2}
            dot={false}
          />
        </LineChart>
      </ResponsiveContainer>
    </div>
  );
}
