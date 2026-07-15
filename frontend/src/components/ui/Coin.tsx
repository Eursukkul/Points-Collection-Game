export function Coin({
  size = 40,
  active = true,
  className = "",
}: {
  size?: number;
  active?: boolean;
  className?: string;
}) {
  return (
    <span
      className={`inline-flex items-center justify-center rounded-full ring-2 ring-white ${
        active ? "bg-gold" : "bg-zinc-300"
      } ${className}`}
      style={{ width: size, height: size }}
    >
      <svg viewBox="0 0 24 24" width={size * 0.55} height={size * 0.55} fill="none">
        <path
          d="M12 3l2.5 5 5.5.8-4 3.9.9 5.5L12 21l-4.9 2.6.9-5.5-4-3.9L9.5 8 12 3z"
          fill={active ? "#fff" : "#f4f4f5"}
        />
      </svg>
    </span>
  );
}
