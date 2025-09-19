## API Routes

### 🔑 Authentication & User
- **POST** `/register` → Register user baru  
- **POST** `/login` → Login (JWT token)  
- **GET** `/me` → Get profile user (butuh auth)  

### 📦 Content Models (Admin only)
- **POST** `/models` → Create new content model  
- **PUT** `/models/:id` → Update content model  
- **DELETE** `/models/:id` → Delete content model  
- **GET** `/models` → List all models  

### 📝 Entries
- **POST** `/entries` → Create entry (admin/editor)  
- **PUT** `/entries/:id` → Update entry (admin/editor)  
- **DELETE** `/entries/:id` → Delete entry (admin only)  
- **GET** `/entries/:id` → Get single entry by ID (admin/editor/viewer)  
- **GET** `/entries/:model_id` → List entries by model (admin/editor/viewer)  
- **POST** `/entries/:id/:action` → Change status (publish/unpublish/suspend) (admin only)  
- **GET** `/entries/:id/audit` → Get audit logs for entry (admin only)  

### 🌍 Public (no auth)
- **GET** `/public/:model` → List public entries by model (published only)  
- **GET** `/public/:model/:slug` → Get public entry by slug  

### 🖼 Media
- **GET** `/media/presign` → Get pre-signed upload URL (admin/editor)  
- **GET** `/media` → List all media (admin/editor/viewer)  
- **DELETE** `/media/:id` → Delete media (admin only)  
