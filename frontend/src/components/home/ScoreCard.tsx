import type { Summary } from "@/lib/types";
import { Button } from "@/components/ui/Button";
import { ProgressBar } from "./ProgressBar";

interface Props {
  summary: Summary;
  claimingCheckpoint: number | null;
  onClaim: (checkpoint: number) => void;
}

export function ScoreCard({ summary, claimingCheckpoint, onClaim }: Props) {
  const { points, maxPoints, checkpoints } = summary;

  return (
    <section className="rounded-2xl bg-surface p-5 shadow-sm">
      <div className="flex items-baseline justify-between">
        <span className="text-sm text-muted">คะแนนสะสม</span>
        <span className="text-2xl font-bold text-accent">
          {points.toLocaleString()}
          <span className="text-base font-medium text-muted">
            /{maxPoints.toLocaleString()}
          </span>
        </span>
      </div>

      <ProgressBar points={points} maxPoints={maxPoints} checkpoints={checkpoints} />

      <div className="flex flex-col gap-2">
        {checkpoints.map((cp) => {
          if (cp.claimed) {
            return (
              <div
                key={cp.checkpoint}
                className="flex items-center justify-between rounded-xl bg-zinc-100 px-4 py-2.5 text-sm"
              >
                <span className="text-muted">
                  {cp.rewardName} ({cp.checkpoint.toLocaleString()})
                </span>
                <span className="font-semibold text-win">✓ รับแล้ว</span>
              </div>
            );
          }
          if (cp.reached) {
            return (
              <Button
                key={cp.checkpoint}
                fullWidth
                loading={claimingCheckpoint === cp.checkpoint}
                disabled={claimingCheckpoint !== null}
                onClick={() => onClaim(cp.checkpoint)}
              >
                รับ{cp.rewardName} (ครบ {cp.checkpoint.toLocaleString()})
              </Button>
            );
          }
          return null;
        })}
      </div>
    </section>
  );
}
