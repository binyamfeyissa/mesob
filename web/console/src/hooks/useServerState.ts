"use client";

import { useQueryClient } from "@tanstack/react-query";

export function useServerState() {
  const queryClient = useQueryClient();

  const invalidate = (queryKey: unknown[]) => {
    queryClient.invalidateQueries({ queryKey });
  };

  return { invalidate };
}
