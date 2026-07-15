import type { ReactNode } from "react";
import { Button } from "./Button";
import { Spinner } from "./Spinner";

export function LoadingState({ label = "กำลังโหลด…" }: { label?: string }) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-10 text-muted">
      <Spinner className="h-6 w-6" />
      <span className="text-sm">{label}</span>
    </div>
  );
}

export function ErrorState({
  message = "เกิดข้อผิดพลาด",
  onRetry,
}: {
  message?: string;
  onRetry?: () => void;
}) {
  return (
    <div className="flex flex-col items-center justify-center gap-3 py-10 text-center">
      <p className="text-sm text-danger">{message}</p>
      {onRetry && (
        <Button variant="ghost" onClick={onRetry}>
          ลองใหม่อีกครั้ง
        </Button>
      )}
    </div>
  );
}

export function EmptyState({ children }: { children: ReactNode }) {
  return (
    <div className="flex flex-col items-center justify-center gap-2 py-10 text-center text-sm text-muted">
      {children}
    </div>
  );
}
