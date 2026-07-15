import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";
import { Coin } from "@/components/ui/Coin";

interface Props {
  rewardName: string | null;
  onClose: () => void;
}

export function RewardModal({ rewardName, onClose }: Props) {
  return (
    <Modal open={rewardName !== null} onClose={onClose}>
      <div className="flex flex-col items-center gap-3 pt-2 text-center">
        <Coin size={64} />
        <h2 className="text-lg font-bold text-ink">ยินดีด้วย</h2>
        <p className="text-sm text-muted">คุณได้รับ{rewardName}</p>
        <Button fullWidth onClick={onClose} className="mt-2">
          ตกลง
        </Button>
      </div>
    </Modal>
  );
}
