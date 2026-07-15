import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Coin } from "@/components/ui/Coin";
import type { PlayResult } from "@/lib/types";

interface Props {
  result: PlayResult | null;
  onClose: () => void;
}

export function ResultModal({ result, onClose }: Props) {
  return (
    <Modal open={result !== null} onClose={onClose}>
      <div className="flex flex-col items-center gap-3 pt-2 text-center">
        <Coin size={64} />
        <h2 className="text-lg font-bold text-ink">ได้รับคะแนน</h2>
        <p className="text-3xl font-bold text-accent">
          {result?.score.toLocaleString()}
        </p>
        {result?.pointsAdded === 0 && (
          <p className="text-xs text-muted">คะแนนสะสมเต็มเพดานแล้ว (10,000)</p>
        )}
        <Button fullWidth onClick={onClose} className="mt-2">
          ตกลง
        </Button>
      </div>
    </Modal>
  );
}
