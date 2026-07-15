"use client";

import { useState } from "react";
import Link from "next/link";
import { mutate } from "swr";
import { api, ApiError } from "@/lib/api";
import { keys, useSummary } from "@/lib/hooks";
import type { PlayResult } from "@/lib/types";
import { ScoreCards } from "@/components/game/ScoreCards";
import { ResultModal } from "@/components/game/ResultModal";
import { Button } from "@/components/ui/Button";

const SCORES = [300, 500, 1000, 3000];
const ELIMINATE_MS = 550;

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

// shuffle returns a new array — visual elimination order varies each round.
function shuffle<T>(arr: T[]): T[] {
  const a = [...arr];
  for (let i = a.length - 1; i > 0; i--) {
    const j = Math.floor(Math.random() * (i + 1));
    [a[i], a[j]] = [a[j], a[i]];
  }
  return a;
}

export default function GamePage() {
  const { data: summary } = useSummary();
  const [playing, setPlaying] = useState(false);
  const [eliminated, setEliminated] = useState<number[]>([]);
  const [winner, setWinner] = useState<number | null>(null);
  const [result, setResult] = useState<PlayResult | null>(null);
  const [error, setError] = useState<string | null>(null);

  async function play() {
    setPlaying(true);
    setError(null);
    setEliminated([]);
    setWinner(null);
    try {
      // Server decides the score first; the animation only visualizes it.
      const res = await api.post<PlayResult>("/game/play");
      const losers = shuffle(SCORES.filter((s) => s !== res.score));
      for (const s of losers) {
        await sleep(ELIMINATE_MS);
        setEliminated((prev) => [...prev, s]);
      }
      await sleep(ELIMINATE_MS);
      setWinner(res.score);
      setResult(res);
      // Refresh the accumulated total only after the reveal keeps the suspense.
      await mutate(keys.summary);
    } catch (e) {
      setError(e instanceof ApiError ? e.message : "เล่นเกมไม่สำเร็จ");
    } finally {
      setPlaying(false);
    }
  }

  function closeResult() {
    setResult(null);
    setEliminated([]);
    setWinner(null);
  }

  return (
    <main className="mx-auto flex min-h-dvh w-full max-w-[500px] flex-col gap-6 px-4 py-6">
      <header className="text-center">
        <p className="text-sm text-muted">คะแนนสะสม</p>
        <p className="text-2xl font-bold text-accent">
          {(summary?.points ?? 0).toLocaleString()}
          <span className="text-base font-medium text-muted">
            /{(summary?.maxPoints ?? 10000).toLocaleString()}
          </span>
        </p>
      </header>

      <section className="rounded-3xl bg-cream p-5">
        <ScoreCards scores={SCORES} eliminated={eliminated} winner={winner} />
      </section>

      {error && <p className="text-center text-sm text-danger">{error}</p>}

      <Button variant="danger" fullWidth loading={playing} onClick={play}>
        สุ่มคะแนน
      </Button>

      <Link href="/" className="mt-auto">
        <Button variant="ghost" fullWidth>
          กลับหน้าหลัก
        </Button>
      </Link>

      <ResultModal result={result} onClose={closeResult} />
    </main>
  );
}
