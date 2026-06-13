"use client";

import { useState, useCallback } from "react";
import { useWebSocket } from "@/hooks/useWebSocket";
import { Badge } from "@/components/ui/Badge";

interface LiveEvent {
  id: string;
  type: string;
  payload: unknown;
  receivedAt: string;
}

const MAX_EVENTS = 100;

const eventColor = (type: string): "blue" | "red" | "yellow" | "green" | "gray" => {
  if (type === "transaction.posted") return "green";
  if (type === "fraud.alert") return "red";
  if (type === "float.low") return "yellow";
  if (type === "metrics.tick") return "blue";
  return "gray";
};

export default function LivePage() {
  const [events, setEvents] = useState<LiveEvent[]>([]);
  const [connected, setConnected] = useState(false);

  const handleMessage = useCallback((msg: { type: string; payload: unknown }) => {
    if (msg.type === "pong") return;
    setEvents((prev) => [
      {
        id: crypto.randomUUID(),
        type: msg.type,
        payload: msg.payload,
        receivedAt: new Date().toISOString(),
      },
      ...prev.slice(0, MAX_EVENTS - 1),
    ]);
    if (!connected) setConnected(true);
  }, [connected]);

  useWebSocket(["transactions", "fraud_alerts", "float", "metrics"], handleMessage);

  return (
    <div className="p-6">
      <div className="flex items-center justify-between mb-6">
        <h1 className="text-xl font-semibold text-gray-900">Live Feed</h1>
        <div className="flex items-center gap-2">
          <span className={`w-2 h-2 rounded-full ${connected ? "bg-green-500" : "bg-gray-300"}`} />
          <span className="text-xs text-gray-500">{connected ? "Connected" : "Connecting..."}</span>
        </div>
      </div>

      <div className="grid grid-cols-4 gap-3 mb-6 text-xs text-gray-500">
        <div className="flex items-center gap-1"><Badge label="txn.posted" color="green" /></div>
        <div className="flex items-center gap-1"><Badge label="fraud.alert" color="red" /></div>
        <div className="flex items-center gap-1"><Badge label="float.low" color="yellow" /></div>
        <div className="flex items-center gap-1"><Badge label="metrics.tick" color="blue" /></div>
      </div>

      <div className="space-y-2 max-h-[600px] overflow-y-auto">
        {events.length === 0 && (
          <p className="text-gray-400 text-sm text-center py-16">Waiting for events...</p>
        )}
        {events.map((e) => (
          <div key={e.id} className="bg-white border border-gray-100 rounded-lg px-4 py-3 flex items-start gap-3">
            <Badge label={e.type} color={eventColor(e.type)} />
            <div className="flex-1 min-w-0">
              <pre className="text-xs text-gray-600 truncate">
                {JSON.stringify(e.payload)}
              </pre>
            </div>
            <span className="text-xs text-gray-400 whitespace-nowrap">
              {new Date(e.receivedAt).toLocaleTimeString()}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
