interface Props {
  scores: number[];
  eliminated: number[];
  winner: number | null;
}

export function ScoreCards({ scores, eliminated, winner }: Props) {
  return (
    <div className="grid grid-cols-2 gap-3">
      {scores.map((score) => {
        const isOut = eliminated.includes(score);
        const isWinner = winner === score;
        return (
          <div
            key={score}
            className={`flex h-20 items-center justify-center rounded-2xl text-2xl font-bold transition-all duration-300 ${
              isWinner
                ? "scale-105 bg-win text-white shadow-lg"
                : isOut
                  ? "scale-90 bg-zinc-100 text-zinc-300 opacity-40"
                  : "bg-score text-white shadow-sm"
            }`}
          >
            {score.toLocaleString()}
          </div>
        );
      })}
    </div>
  );
}
