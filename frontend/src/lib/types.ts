// Response shapes mirroring the Go backend (see backend/openapi.yaml).

export interface CheckpointStatus {
  checkpoint: number;
  rewardName: string;
  reached: boolean;
  claimed: boolean;
}

export interface Summary {
  points: number;
  maxPoints: number;
  checkpoints: CheckpointStatus[];
}

export interface PlayResult {
  score: number;
  pointsAdded: number;
  totalPoints: number;
}

export interface Play {
  id: string;
  score: number;
  createdAt: string;
}

export interface Claim {
  id: string;
  checkpoint: number;
  rewardName: string;
  createdAt: string;
}

export interface HistoryResponse<T> {
  items: T[];
}

export interface ApiErrorBody {
  error: { code: string; message: string };
}
