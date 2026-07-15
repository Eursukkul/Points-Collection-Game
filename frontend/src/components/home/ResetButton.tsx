"use client";

import { useState } from "react";
import { Modal } from "@/components/ui/Modal";
import { Button } from "@/components/ui/Button";

interface Props {
  onReset: () => Promise<void>;
}

export function ResetButton({ onReset }: Props) {
  const [confirming, setConfirming] = useState(false);
  const [resetting, setResetting] = useState(false);

  async function handleReset() {
    setResetting(true);
    try {
      await onReset();
      setConfirming(false);
    } finally {
      setResetting(false);
    }
  }

  return (
    <>
      <button
        onClick={() => setConfirming(true)}
        className="mx-auto block rounded-full bg-reset px-6 py-1.5 text-xs font-bold uppercase tracking-wide text-white shadow-sm transition hover:brightness-110"
      >
        Reset
      </button>

      <Modal open={confirming} onClose={() => !resetting && setConfirming(false)}>
        <div className="flex flex-col gap-4 pt-2 text-center">
          <h2 className="text-lg font-bold text-ink">รีเซตข้อมูล?</h2>
          <p className="text-sm text-muted">
            คะแนนสะสม ประวัติการเล่น และรางวัลทั้งหมดจะถูกล้าง
          </p>
          <div className="flex gap-2">
            <Button
              variant="ghost"
              fullWidth
              disabled={resetting}
              onClick={() => setConfirming(false)}
            >
              ยกเลิก
            </Button>
            <Button variant="danger" fullWidth loading={resetting} onClick={handleReset}>
              รีเซต
            </Button>
          </div>
        </div>
      </Modal>
    </>
  );
}
