"use client";

import useSWR, { mutate } from "swr";
import { api, fetcher } from "./api";
import type { Claim, HistoryResponse, Play, PlayResult, Summary } from "./types";

export const keys = {
  summary: "/me",
  plays: "/history/plays",
  claims: "/history/claims",
} as const;

export function useSummary() {
  return useSWR<Summary>(keys.summary, fetcher);
}

export function usePlayHistory() {
  return useSWR<HistoryResponse<Play>>(keys.plays, fetcher);
}

export function useClaimHistory() {
  return useSWR<HistoryResponse<Claim>>(keys.claims, fetcher);
}

// Revalidate every player-scoped view after a state change so Home and Game
// never drift apart.
async function revalidateAll() {
  await Promise.all([
    mutate(keys.summary),
    mutate(keys.plays),
    mutate(keys.claims),
  ]);
}

export async function playGame(): Promise<PlayResult> {
  const result = await api.post<PlayResult>("/game/play");
  await revalidateAll();
  return result;
}

export async function claimReward(checkpoint: number): Promise<Claim> {
  const claim = await api.post<Claim>("/claims", { checkpoint });
  await revalidateAll();
  return claim;
}

export async function resetGame(): Promise<void> {
  await api.post<void>("/reset");
  await revalidateAll();
}
