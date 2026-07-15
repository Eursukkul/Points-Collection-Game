import { describe, expect, it } from "vitest";
import { isClaimable, progressPercent } from "./progress";
import type { CheckpointStatus } from "./types";

describe("progressPercent", () => {
  it("returns 0 at zero points", () => {
    expect(progressPercent(0, 10000)).toBe(0);
  });

  it("scales linearly", () => {
    expect(progressPercent(5000, 10000)).toBe(50);
    expect(progressPercent(7500, 10000)).toBe(75);
  });

  it("clamps at 100 when points meet or exceed the ceiling", () => {
    expect(progressPercent(10000, 10000)).toBe(100);
    expect(progressPercent(15000, 10000)).toBe(100);
  });

  it("never goes negative and guards a zero ceiling", () => {
    expect(progressPercent(-100, 10000)).toBe(0);
    expect(progressPercent(100, 0)).toBe(0);
  });
});

describe("isClaimable", () => {
  const cp = (reached: boolean, claimed: boolean): CheckpointStatus => ({
    checkpoint: 5000,
    rewardName: "รางวัล A",
    reached,
    claimed,
  });

  it("is claimable only when reached and not yet claimed", () => {
    expect(isClaimable(cp(true, false))).toBe(true);
  });

  it("is not claimable before reaching the checkpoint", () => {
    expect(isClaimable(cp(false, false))).toBe(false);
  });

  it("is not claimable once already claimed", () => {
    expect(isClaimable(cp(true, true))).toBe(false);
  });
});
