"use client";

import { useState } from "react";
import { usePlayHistory, useClaimHistory } from "@/lib/hooks";
import { LoadingState, ErrorState, EmptyState } from "@/components/ui/States";

type Tab = "plays" | "rewards";

function formatDate(iso: string) {
  return new Date(iso).toLocaleString("th-TH", {
    day: "2-digit",
    month: "2-digit",
    year: "numeric",
    hour: "2-digit",
    minute: "2-digit",
  });
}

export function HistoryTabs() {
  const [tab, setTab] = useState<Tab>("plays");
  const plays = usePlayHistory();
  const claims = useClaimHistory();

  return (
    <section className="rounded-2xl bg-surface p-4 shadow-sm">
      <div className="mb-3 flex gap-2">
        <TabButton active={tab === "plays"} onClick={() => setTab("plays")}>
          ประวัติการเล่น
        </TabButton>
        <TabButton active={tab === "rewards"} onClick={() => setTab("rewards")}>
          ประวัติรางวัล
        </TabButton>
      </div>

      {tab === "plays" ? (
        <List
          loading={plays.isLoading}
          error={plays.error ? "โหลดประวัติการเล่นไม่สำเร็จ" : null}
          onRetry={() => plays.mutate()}
          empty={plays.data?.items.length === 0}
          emptyText="ยังไม่มีประวัติการเล่น"
        >
          {plays.data?.items.map((p) => (
            <Row
              key={p.id}
              dot="bg-danger"
              title={`ได้คะแนน ${p.score.toLocaleString()}`}
              time={formatDate(p.createdAt)}
            />
          ))}
        </List>
      ) : (
        <List
          loading={claims.isLoading}
          error={claims.error ? "โหลดประวัติรางวัลไม่สำเร็จ" : null}
          onRetry={() => claims.mutate()}
          empty={claims.data?.items.length === 0}
          emptyText="ยังไม่มีรางวัลที่ได้รับ"
        >
          {claims.data?.items.map((c) => (
            <Row
              key={c.id}
              dot="bg-win"
              title={`ได้รับ${c.rewardName}`}
              time={formatDate(c.createdAt)}
            />
          ))}
        </List>
      )}
    </section>
  );
}

function TabButton({
  active,
  onClick,
  children,
}: {
  active: boolean;
  onClick: () => void;
  children: React.ReactNode;
}) {
  return (
    <button
      onClick={onClick}
      className={`rounded-full px-4 py-1.5 text-sm font-medium transition ${
        active ? "bg-ink text-white" : "bg-zinc-100 text-muted"
      }`}
    >
      {children}
    </button>
  );
}

function List({
  loading,
  error,
  onRetry,
  empty,
  emptyText,
  children,
}: {
  loading: boolean;
  error: string | null;
  onRetry: () => void;
  empty?: boolean;
  emptyText: string;
  children: React.ReactNode;
}) {
  if (loading) return <LoadingState />;
  if (error) return <ErrorState message={error} onRetry={onRetry} />;
  if (empty) return <EmptyState>{emptyText}</EmptyState>;
  return <ul className="flex flex-col divide-y divide-zinc-100">{children}</ul>;
}

function Row({ dot, title, time }: { dot: string; title: string; time: string }) {
  return (
    <li className="flex items-center gap-3 py-3">
      <span className={`h-8 w-8 shrink-0 rounded-full ${dot}`} />
      <div className="flex flex-col">
        <span className="text-sm font-medium text-ink">{title}</span>
        <span className="text-xs text-muted">{time}</span>
      </div>
    </li>
  );
}
