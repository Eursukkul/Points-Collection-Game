import type { CheckpointStatus } from "./types";

/** Percentage (0–100, clamped) of points toward the ceiling. */
export function progressPercent(points: number, maxPoints: number): number {
  if (maxPoints <= 0) return 0;
  return Math.max(0, Math.min((points / maxPoints) * 100, 100));
}

/** A checkpoint can be claimed once reached but not yet claimed. */
export function isClaimable(cp: CheckpointStatus): boolean {
  return cp.reached && !cp.claimed;
}
