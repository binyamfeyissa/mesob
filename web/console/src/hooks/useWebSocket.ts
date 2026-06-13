"use client";

import { useEffect, useRef } from "react";

type WsMessage = {
  type: string;
  payload: unknown;
};

export function useWebSocket(
  channels: string[],
  onMessage: (msg: WsMessage) => void
) {
  const onMessageRef = useRef(onMessage);
  onMessageRef.current = onMessage;

  const channelsKey = channels.join(",");

  useEffect(() => {
    const gatewayUrl = process.env.NEXT_PUBLIC_GATEWAY_URL ?? "http://localhost:8000";
    const sseUrl = gatewayUrl + "/v1/admin/live";

    let es: EventSource | null = null;
    let destroyed = false;
    let reconnectTimer: ReturnType<typeof setTimeout> | null = null;

    function connect() {
      if (destroyed) return;
      es = new EventSource(sseUrl, { withCredentials: false });

      es.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          onMessageRef.current({ type: event.type || "message", payload: data });
        } catch {
          // ignore malformed frames
        }
      };

      // Named event types (e.g. "event: summary")
      const handleNamed = (event: MessageEvent) => {
        try {
          const data = JSON.parse(event.data);
          onMessageRef.current({ type: event.type, payload: data });
        } catch {
          // ignore
        }
      };

      es.addEventListener("summary", handleNamed);
      es.addEventListener("heartbeat", handleNamed);

      es.onerror = () => {
        es?.close();
        if (!destroyed) {
          reconnectTimer = setTimeout(connect, 5_000);
        }
      };
    }

    connect();

    return () => {
      destroyed = true;
      if (reconnectTimer) clearTimeout(reconnectTimer);
      es?.close();
    };
  // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [channelsKey]);
}
