import type { CheckpointStatus } from "@/lib/types";
import { progressPercent } from "@/lib/progress";
import { Coin } from "@/components/ui/Coin";

interface Props {
  points: number;
  maxPoints: number;
  checkpoints: CheckpointStatus[];
}

export function ProgressBar({ points, maxPoints, checkpoints }: Props) {
  const pct = progressPercent(points, maxPoints);

  return (
    <div className="px-2 pb-8 pt-4">
      <div className="relative h-2.5 rounded-full bg-zinc-200">
        <div
          className="h-full rounded-full bg-accent transition-[width] duration-500"
          style={{ width: `${pct}%` }}
        />
        {checkpoints.map((cp) => {
          const left = progressPercent(cp.checkpoint, maxPoints);
          return (
            <div
              key={cp.checkpoint}
              className="absolute top-1/2 flex -translate-x-1/2 -translate-y-1/2 flex-col items-center"
              style={{ left: `${left}%` }}
            >
              <Coin size={34} active={cp.reached} />
              <span className="mt-1 text-[11px] font-medium text-muted">
                {cp.checkpoint.toLocaleString()}
              </span>
            </div>
          );
        })}
      </div>
    </div>
  );
}
