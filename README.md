# Path Alchemy 2  🧬
 A repository contains Web-based implementation of BFS and DFS Algorithms in Recipe Searching in the [**Little Alchemy 2**](https://littlealchemy2.com/) game.
 
---

 <!-- CONTRIBUTOR -->
 <div align="center" id="contributor">
   <strong>
     <h3>~ SeleniumSoup4 🍲 ~</h3>
     <table align="center">
       <tr align="center">
         <td>NIM</td>
         <td>Nama</td>
         <td>GitHub</td>
       </tr>
       <tr align="center">
         <td>13523002</td>
         <td>Refki Alfarizi</td>
         <td><a href="https://github.com/l0stplains">@l0stplains</a></td>
       </tr>
       <tr align="center">
         <td>13523004</td>
         <td>Razi Rachman Widyadhana</td>
         <td><a href="https://github.com/zirachw">@zirachw</a></td>
       </tr>
       <tr align="center">
         <td>13523044</td>
         <td>Muhammad Luqman Hakim</td>
         <td><a href="https://github.com/carasiae">@carasiae</a></td>
       </tr>
     </table>
   </strong>
 </div>
 
 <div align="center">
   <h3 align="center">~ Tech Stacks ~ </h3>
 
   <p align="center">
 
 [![Next JS](https://img.shields.io/badge/Next-black?style=for-the-badge&logo=next.js&logoColor=white)][Next-url]
 [![React](https://img.shields.io/badge/react-%2320232a.svg?style=for-the-badge&logo=react&logoColor=%2361DAFB)][React-url]
 [![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)][Go-url]
 
   </p>
 </div>

---



<!-- MARKDOWN LINKS & IMAGES -->
[Next-url]: https://nextjs.org/
[React-url]: https://react.dev/
[Go-url]: https://go.dev/


---

## ️Running Locally

> **⚠️ Before you start, set up your environment variables!**

### 1. If You’re Using Docker Compose

Place a single `.env` file in the **project root**. Docker Compose will automatically load it.

**`./.env`**  
```dotenv
# Backend
PORT=8080
CORS_ALLOWED_ORIGIN=http://localhost:3000

# Frontend
NODE_ENV=production
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

> **Note:**
>
> * Docker Compose reads `./.env` by default.
> * The frontend service can reach the backend at `http://backend:8080` (the service name).
> * Frontend: [http://localhost:3000](http://localhost:3000)
> * Backend:  [http://localhost:8080](http://localhost:8080)

### Docker Compose (recommended)
```bash
# from project root
docker-compose up --build
```

> [!NOTE]  
> A docker engine instance must be running (usually via Docker Desktop)

Then:
- Frontend at http://localhost:3000
- Backend API at http://localhost:8080


### 2. Manual (no Docker)

Create **two** `.env` files—one in the `src/backend` folder and one in the `src/frontend` folder.

**`src/backend/.env`**

```dotenv
PORT=8080
CORS_ALLOWED_ORIGIN=http://localhost:3000
```

**`src/frontend/.env.local`**

```dotenv
NODE_ENV=production
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

Then:

1. **Backend**

   ```bash
   cd src/backend
   # Ensure src/backend/.env exists
   go run ./cmd/server
   ```

   *Server listens on the port defined in src/backend/.env (default 8080).*

2. **Frontend**

   ```bash
   cd src/frontend
   # Ensure src/frontend/.env.local exists
   npm install
   npm run dev
   ```

   *App opens at [http://localhost:3000](http://localhost:3000) and proxies API calls to the URL defined in NEXT\_PUBLIC\_API\_BASE\_URL.*

---

> 🎉 Now you’re ready to explore the Little Alchemy 2 recipe finder with BFS/DFS and live SSE updates!

```
```


1. **Backend**
   ```bash
   cd src/backend
   go run ./cmd/server
   ```
   > Runs on port 8080 by default.

2. **Frontend**
   ```bash
   cd src/frontend
   npm install
   npm run dev
   ```
   > Opens at http://localhost:3000, API proxy to http://localhost:8080

---