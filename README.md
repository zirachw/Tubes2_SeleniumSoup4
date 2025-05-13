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

## 🔎 Preview
![2025-05-1321-43-54-ezgif com-video-to-gif-converter](https://github.com/user-attachments/assets/54bca0aa-f8a6-484e-996c-056c4c4acb01)

---

## ✨ Features

### This project contains:

1. **Search element name**
2. **Single/Multiple recipes option**
3. **Responsive website**
4. **`(Bonus)` Recipe search live update**
5. **`(Bonus)` *containerized* project & deployment**

### **Space for Improvement:** 

1. **`(Bonus)` Bidirectional search**

--- 
## Algorithm Explanation
### BFS
1. Group all elements into buckets based on their tier values.

2. Initialize memo with tier 0 elements. Add each base element as a single node and mark it as traced.

3. For each element in a given tier, try to form a new path from that element and the elements in memo. If the combination is valid and the number of paths does not exceed maxPath, add the path to the memo.

4. Once all tiers are complete, check if the target element is available in memo. If yes, retrieve the paths that leads to that target.

5. Make the target node the root of the tree.

6. Insert as many element paths into the tree as required.

### DFS
1. Verify the existence of the target element in the data set. If the element is not found, the process is stopped. If found, a root node is created as the initial representation of the solution structure.
    
2. Base case: If the target element is in tier 0, the recipe search is complete, as it is a base element.
    
3. Recursion: Count the number of unique paths of all possible component pairs. Then, find recipes for both elements from as many recipe pairs as the desired number of recipes with DFS.
    
4. Merge all valid subtrees into the root node.
---

## ️Running Locally

> **⚠️ Before you start, set up your environment variables!**

### 1. If You’re Using Docker Compose

Place a single `.env` file in the **project root**. Docker Compose will automatically load it.

**`./.env`**  
```dotenv
# Backend
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
CORS_ALLOWED_ORIGIN=http://localhost:3000
```

**`src/frontend/.env.local`**

```dotenv
NODE_ENV=production
NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
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
 ## 📃 Miscellaneous

 <div align="center">
   
 | No | Points | Ya | Tidak |
 | --- | --- | --- | --- |
 | 1 | Aplikasi dapat dijalankan. | ✔️ | |
 | 2 | Aplikasi dapat memperoleh data *recipe* melalui scraping. | ✔️ | |
 | 3 | Algoritma **Depth First Search** dan **Breadth First Search** dapat menemukan *recipe* elemen dengan benar. | ✔️ | |
 | 4 | Aplikasi dapat menampilkan visualisasi *recipe* elemen yang dicari sesuai dengan spesifikasi.| ✔️ | |
 | 5 | Aplikasi mengimplementasikan multithreading. | ✔️ | |
 | 6 | Membuat laporan sesuai dengan spesifikasi. | ✔️ | |
 | 7 | Membuat bonus video dan diunggah pada Youtube. | ✔️ | |
 | 6 | **[Bonus]** Membuat algoritma pencarian *Bidirectional*. |  | ✔️|
 | 7 | **[Bonus]** Membuat *Live Update* | ✔️ | |
 | 8 | **[Bonus]** Aplikasi di-*containerize* dengan Docker. | ✔️ | |
 | 8 | **[Bonus]** Aplikasi di-*deploy* dan dapat diakses melalui internet. | ✔️ | |

 </div>
