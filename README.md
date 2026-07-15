# เกมสะสมคะแนน Nextzy — Points Collection Game

เว็บแอปเกมสะสมคะแนน: ผู้เล่นกด **สุ่มคะแนน** (300 / 500 / 1,000 / 3,000) สะสมจนถึงเพดาน **10,000** แล้วปลดล็อกรางวัลตาม checkpoint (**5,000 / 7,500 / 10,000**) พร้อมดูประวัติการเล่นและประวัติรางวัล

## 🔗 Live

| | URL |
|---|---|
| **เว็บไซต์ (Frontend)** | https://points-collection-game.vercel.app |
| **API (Backend)** | https://points-collection-game.onrender.com |
| **Health check** | https://points-collection-game.onrender.com/healthz |

> ⚠️ **Cold start:** Backend รันบน Render free tier ที่ spin down หลัง idle ~15 นาที → **request แรกอาจใช้เวลา ~50 วินาที** จากนั้นจะเร็วปกติ (เปิด `/healthz` ทิ้งไว้สัก 1 ครั้งเพื่อ warm ก่อนใช้งาน)

## 🧱 Tech Stack

| Layer | Stack |
|---|---|
| Frontend | Next.js 16 (App Router) · React 19 · Tailwind CSS v4 · SWR — `frontend/` |
| Backend | Go 1.26 · Fiber v2 · Gorm — `backend/` |
| Database | PostgreSQL (Neon, region Singapore) |
| Deploy | Vercel (FE) · Render (BE) · Neon (DB) |

## 📱 Features

**หน้า Home** (`/`)
- แสดงคะแนนสะสม + progress bar พร้อม 3 checkpoint (5,000 / 7,500 / 10,000)
- กดรับรางวัลเมื่อถึง checkpoint → modal "ยินดีด้วย" · รับซ้ำไม่ได้ (idempotent)
- 2 แท็บประวัติ: **ประวัติการเล่น** / **ประวัติรางวัล** (มี loading / error / empty states)
- ปุ่ม **Reset** (มี confirm dialog) ล้างข้อมูลทั้งฝั่งเว็บและ database
- ปุ่ม **ไปเล่นเกม** → หน้า Game

**หน้า Game** (`/game`)
- การ์ดคะแนน 4 ใบ (300 / 500 / 1,000 / 3,000)
- กด **สุ่มคะแนน** → การ์ดค่อยๆ หายทีละใบจนเหลือคะแนนที่ได้ (elimination animation) → modal แสดงคะแนน
- คะแนนบวกเข้าคะแนนสะสม · เล่นซ้ำได้ไม่จำกัด · ปุ่มกลับหน้าหลัก

Responsive 300–500px (mobile-first) ตาม Figma

## 🏛️ Architecture

Monorepo: `frontend/` (Next.js) + `backend/` (Go) + `docker-compose.yml` (Postgres สำหรับ local)

### Backend — Clean Architecture

dependency ชี้เข้าใน (Dependency Rule) — ชั้นในไม่รู้จักชั้นนอก:

```
internal/
├── domain/       เอนทิตี (ไร้ framework tag) + checkpoint + errors + ports (interfaces)
│                 ↑ ไม่ import อะไรจากชั้นนอกเลย
├── usecase/      business logic — พึ่ง domain ports เท่านั้น (Play/Claim/Reset/Summary/History)
├── repository/   Gorm impl ของ ports + models + mapping domain↔model + TxManager + crypto rand
├── handler/      HTTP adapter (Fiber) + DTOs (json)
├── middleware/   EnsurePlayer (cookie) + CSRFGuard — พึ่ง usecase ผ่าน interface
├── server/       composition root (wire ทุกอย่าง)
├── apierr/       error envelope กลาง { error: { code, message } }
└── config/       env loading
```

**ประโยชน์:** business logic เทสได้โดยไม่ผูก framework · ตอน migrate Gin→Fiber แตะแค่ adapter ไม่แตะ usecase · gorm ถูกจำกัดอยู่แค่ `repository/`

### Database schema

```
players (id uuid PK, points int [CHECK 0..10000], created_at)
plays   (id uuid PK, player_id FK→players ON DELETE CASCADE, score int, created_at)
        index (player_id, created_at DESC)
claims  (id uuid PK, player_id FK→players ON DELETE CASCADE, checkpoint int,
         reward_name text, created_at, UNIQUE(player_id, checkpoint))
```

### API

Base `/api/v1` — ทุก endpoint ระบุผู้เล่นด้วย httpOnly cookie `player_id` (server สร้างให้อัตโนมัติ)

| Method | Path | คำอธิบาย |
|---|---|---|
| GET | `/me` | คะแนน + สถานะ 3 checkpoint (reached/claimed) |
| POST | `/game/play` | server สุ่มคะแนน (crypto/rand) + clamp 10,000 |
| POST | `/claims` | รับรางวัล checkpoint (body `{checkpoint}`) |
| GET | `/history/plays` | ประวัติการเล่น |
| GET | `/history/claims` | ประวัติรางวัล |
| POST | `/reset` | ล้างข้อมูลผู้เล่น |
| GET | `/healthz` | health check |

รายละเอียดเต็ม (schema + ทุก status code) อยู่ที่ `backend/openapi.yaml`

## 🔐 Security & Data Integrity

- **Server-authoritative:** การสุ่มคะแนน (crypto/rand), การ clamp เพดาน, และการตรวจ checkpoint ทำฝั่ง server ทั้งหมด — client ส่งได้แค่ intent
- **Anti-cheat identity:** player id สร้างฝั่ง server เท่านั้น (cookie ปลอมที่ไม่ตรง DB → ได้ player ใหม่)
- **Concurrency:** `SELECT ... FOR UPDATE` ล็อกแถว player ใน transaction ตอน play/claim/reset → กัน lost update / race
- **Idempotency:** รับรางวัลซ้ำถูกกันด้วย `UNIQUE(player_id, checkpoint)`
- **CSRF:** cookie เป็น SameSite=None (cross-site) → มี `CSRFGuard` ตรวจ Origin ของ request ที่เปลี่ยน state
- **Input validation** ที่ boundary + Gorm parameterized (กัน SQL injection)
- **Secrets** ผ่าน env ทั้งหมด (ไม่มีใน repo) · cookie security derive จาก scheme ของ `FRONTEND_ORIGIN`

## 🚀 Getting Started (Local)

**ต้องมี:** Docker, Go 1.26+, Node 20+

```bash
# 1. Database
docker compose up -d              # Postgres ที่ localhost:5432

# 2. Backend  (http://localhost:8080)
cd backend
cp .env.example .env              # ค่า default ตรงกับ docker-compose อยู่แล้ว
go run ./cmd/server

# 3. Frontend (http://localhost:3000)
cd frontend
cp .env.example .env.local        # NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
npm install
npm run dev
```

## 🧪 Testing

```bash
# Backend — 19 tests (ต้องมี Postgres รันอยู่)
cd backend && go test ./...

# Frontend — 7 tests
cd frontend && npm test
```

ครอบคลุม logic สำคัญ: clamp เพดานคะแนน, gain 0 ที่เพดาน, สุ่มอยู่ในเซ็ต, รับรางวัล/ต่ำกว่า threshold/รับซ้ำ, reset ล้างเฉพาะผู้เล่นตัวเอง, concurrent play (FOR UPDATE), cookie bootstrap/tamper, CSRF, checkpoint derivation

## 📌 Documented Assumptions

โจทย์ไม่ได้ระบุบางจุด — ตัดสินใจดังนี้ (อธิบายเหตุผลได้ตอนสัมภาษณ์):

1. **Identity:** ไม่มี login ในโจทย์ → ผู้เล่น = anonymous ผูกกับ httpOnly cookie ที่ server สร้าง (ได้ server-authoritative โดยไม่ over-engineer ทำระบบ auth)
2. **เพดาน 10,000:** clamp ฝั่ง server — ที่เพดานแล้วยังกดเล่นได้แต่คะแนนที่บวกจริง = 0 (FE โชว์ MAX)
3. **รับรางวัล = milestone claim ไม่หักคะแนน** — ตีความจาก "ปลดล็อกตาม checkpoint"; กันรับซ้ำด้วย unique constraint
4. **การสุ่มเป็น server-authoritative:** FE เรียก API ได้ผลก่อน แล้วค่อยเล่น animation ให้ไป land ที่ผลนั้น (animation เป็น visual ล้วน)

## ☁️ Deployment

- **Frontend → Vercel:** root directory = `frontend`, env `NEXT_PUBLIC_API_BASE_URL` = URL ของ backend
- **Backend → Render:** Docker (root `backend/`), region Singapore, env `DATABASE_URL` + `FRONTEND_ORIGIN`
- **Database → Neon:** PostgreSQL free, region Singapore
