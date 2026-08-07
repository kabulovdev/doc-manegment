# Frontend — Doc Management

Next.js 14 (App Router) frontend. Backend'ga `NEXT_PUBLIC_API_BASE` orqali ulanadi.

## Lokal ishlab chiqish (dev)

```bash
cd frontend
npm install        # birinchi marta
npm run dev        # http://localhost:3000
```

Dev rejimda API manzili default `http://localhost:8080/api/v1` bo'ladi (`lib/api/client.ts`).

---

## Production build va serverga deploy qilish

Loyihada tayyor buyruq bor — **repo root'dan** (Makefile turgan joydan) ishga tushiriladi:

```bash
make frontend-dist API_BASE=http://SERVER_IP:8080/api/v1
```

Bu buyruq:
1. `NEXT_PUBLIC_API_BASE` ni berilgan qiymat bilan o'rnatib `npm run build` qiladi
2. Natijani `frontend-dist/` papkaga yig'adi (`server.js` + minimal `node_modules` + static fayllar)

Keyin git'ga yuborish:

```bash
git add frontend-dist
git commit -m "rebuild frontend"
git push origin main
```

### Backend manzilini (API path) o'zgartirish

API manzili **build vaqtida kodga muhrlanadi** — ishlab turgan build'da uni o'zgartirib bo'lmaydi.
Manzil o'zgarsa (yangi IP, domen, port), **qayta build qilish shart**:

```bash
# IP bilan:
make frontend-dist API_BASE=http://45.130.148.20:8080/api/v1

# Domen bilan (nginx/https ortida bo'lsa):
make frontend-dist API_BASE=https://api.mydomain.uz/api/v1

# Faqat lokal test uchun:
make frontend-dist API_BASE=http://localhost:8080/api/v1
```

So'ng yana `git add frontend-dist && git commit && git push`.

> Eslatma: `API_BASE` bermasangiz buyruq xato beradi — bu ataylab qilingan,
> tasodifan noto'g'ri manzil bilan build qilib qo'ymaslik uchun.

### Kod o'zgarganda qayta deploy (to'liq tsikl)

```bash
# 1. Lokalda kodni o'zgartirasiz va test qilasiz (npm run dev)
# 2. Repo root'dan build:
make frontend-dist API_BASE=http://SERVER_IP:8080/api/v1
# 3. Push:
git add frontend-dist && git commit -m "rebuild frontend" && git push
# 4. Serverda:
git pull
pm2 restart docmgmt-frontend    # yoki node server.js ni qayta ishga tushirish
```

---

## Serverda ishga tushirish

Talab: **Node.js 20** (minimum 18.17).

```bash
git clone https://github.com/kabulovdev/doc-manegment.git
cd doc-manegment/frontend-dist
PORT=3001 node server.js
```

Doimiy ishlashi uchun pm2 bilan:

```bash
npm install -g pm2
cd doc-manegment/frontend-dist
PORT=3001 pm2 start server.js --name docmgmt-frontend
pm2 save && pm2 startup    # server restart bo'lsa ham avtomatik ko'tarilsin
```

### Backend sozlamalari (frontend ishlashi uchun shart)

Backend env'ida quyidagilar frontend manziliga mos bo'lishi kerak, aks holda login/cookie ishlamaydi (CORS):

```
FRONTEND_ORIGIN=http://SERVER_IP:3001    # frontend qaysi manzildan ochilsa, shu
COOKIE_SECURE=false                      # http ishlatilsa; https bo'lsa true
```

---

## Tez-tez uchraydigan muammolar

| Muammo | Sabab / Yechim |
|---|---|
| Sayt ochiladi, lekin login ishlamaydi | Build noto'g'ri `API_BASE` bilan qilingan (masalan localhost). To'g'ri manzil bilan qayta build qiling. |
| CORS xatosi (browser console'da) | Backend'dagi `FRONTEND_ORIGIN` frontend manziliga mos emas. |
| `frontend-dist` ichidagi fayllar git'ga tushmayapti | Root `.gitignore` da `!frontend-dist/` va `!frontend-dist/**` qatorlari turganini tekshiring. |
| Serverda `node server.js` xato beradi | Node versiyasini tekshiring: `node -v` — 18.17+ (tavsiya: 20). |
| Port band | `PORT=boshqa_port node server.js` bilan boshqa portda ishga tushiring. |
