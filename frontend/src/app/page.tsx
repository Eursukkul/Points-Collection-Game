"use client";

import { useState } from "react";
import Link from "next/link";
import { useSummary, claimReward, resetGame } from "@/lib/hooks";
import { ApiError } from "@/lib/api";
import { ScoreCard } from "@/components/home/ScoreCard";
import { RewardModal } from "@/components/home/RewardModal";
import { HistoryTabs } from "@/components/home/HistoryTabs";
import { ResetButton } from "@/components/home/ResetButton";
import { Button } from "@/components/ui/Button";
import { LoadingState, ErrorState } from "@/components/ui/States";

export default function HomePage() {
  const { data: summary, error, isLoading, mutate } = useSummary();
  const [claiming, setClaiming] = useState<number | null>(null);
  const [reward, setReward] = useState<string | null>(null);
  const [claimError, setClaimError] = useState<string | null>(null);

  async function handleClaim(checkpoint: number) {
    setClaiming(checkpoint);
    setClaimError(null);
    try {
      const claim = await claimReward(checkpoint);
      setReward(claim.rewardName);
    } catch (e) {
      // A stale UI (already claimed / not reached) resolves once summary refetches.
      setClaimError(e instanceof ApiError ? e.message : "รับรางวัลไม่สำเร็จ");
      mutate();
    } finally {
      setClaiming(null);
    }
  }

  return (
    <main className="mx-auto flex min-h-dvh w-full max-w-[500px] flex-col gap-4 px-4 py-6">
      <h1 className="text-center text-lg font-bold text-ink">เกมสะสมคะแนน Nextzy</h1>

      {isLoading && <LoadingState />}
      {error && <ErrorState message="โหลดข้อมูลไม่สำเร็จ" onRetry={() => mutate()} />}

      {summary && (
        <>
          <ScoreCard
            summary={summary}
            claimingCheckpoint={claiming}
            onClaim={handleClaim}
          />
          {claimError && (
            <p className="text-center text-sm text-danger">{claimError}</p>
          )}

          <ResetButton onReset={resetGame} />

          <HistoryTabs />

          <Link href="/game" className="mt-2">
            <Button fullWidth>ไปเล่นเกม</Button>
          </Link>
        </>
      )}

      <RewardModal rewardName={reward} onClose={() => setReward(null)} />
    </main>
  );
}
