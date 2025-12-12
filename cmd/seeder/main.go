package main

import (
	"log"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

func main() {
	log.Println("🌱 Starting database seeder...")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Connect to database
	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	// Clear existing data (optional - comment out in production)
	log.Println("Clearing existing data...")
	clearData(db)

	// Seed data
	log.Println("Seeding users...")
	users := seedUsers(db)

	log.Println("Seeding categories...")
	categories := seedCategories(db)

	log.Println("Seeding projects...")
	projects := seedProjects(db, users, categories)

	log.Println("Seeding articles...")
	seedArticles(db, users)

	log.Println("Seeding comments...")
	seedComments(db, users, projects)

	log.Println("✅ Seeding completed successfully!")
}

func clearData(db *gorm.DB) {
	// Clear in reverse order of dependencies
	db.Exec("DELETE FROM project_likes")
	db.Exec("DELETE FROM comments")
	db.Exec("DELETE FROM transactions")
	db.Exec("DELETE FROM project_images")
	db.Exec("DELETE FROM projects")
	db.Exec("DELETE FROM articles")
	db.Exec("DELETE FROM reports")
	db.Exec("DELETE FROM block_records")
	db.Exec("DELETE FROM users")
	db.Exec("DELETE FROM categories")
}

func seedUsers(db *gorm.DB) []models.User {
	// Hash password for demo users
	hashedPassword, _ := utils.HashPassword("password123")

	users := []models.User{
		{
			ID:           uuid.New(),
			Email:        "admin@campus-hub.com",
			PasswordHash: &hashedPassword,
			Name:         "Admin User",
			AvatarURL:    strPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=Admin"),
			University:   strPtr("Universitas Indonesia"),
			Major:        strPtr("Sistem Informasi"),
			Bio:          strPtr("Administrator platform Campus Project Hub."),
			Phone:        strPtr("+62812345678"),
			Role:         models.RoleAdmin,
			Status:       models.StatusActive,
			TotalExp:     5000,
		},
		{
			ID:           uuid.New(),
			Email:        "moderator@campus-hub.com",
			PasswordHash: &hashedPassword,
			Name:         "Siti Moderator",
			AvatarURL:    strPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=Siti"),
			University:   strPtr("Institut Teknologi Bandung"),
			Major:        strPtr("Teknik Informatika"),
			Bio:          strPtr("Moderator yang bertugas menjaga kualitas konten."),
			Phone:        strPtr("+62887654321"),
			Role:         models.RoleModerator,
			Status:       models.StatusActive,
			TotalExp:     2500,
		},
		{
			ID:           uuid.New(),
			Email:        "budi@ui.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Budi Santoso",
			AvatarURL:    strPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=Budi"),
			University:   strPtr("Universitas Indonesia"),
			Major:        strPtr("Ilmu Komputer"),
			Bio:          strPtr("Passionate developer specializing in web and mobile development."),
			Phone:        strPtr("+62811223344"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     1250,
		},
		{
			ID:           uuid.New(),
			Email:        "siti@itb.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Siti Nurhaliza",
			AvatarURL:    strPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=SitiN"),
			University:   strPtr("Institut Teknologi Bandung"),
			Major:        strPtr("Teknik Perangkat Lunak"),
			Bio:          strPtr("Software engineer focused on clean code and best practices."),
			Phone:        strPtr("+62822334455"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     850,
		},
		{
			ID:           uuid.New(),
			Email:        "ahmad@ugm.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Ahmad Rizki",
			AvatarURL:    strPtr("https://api.dicebear.com/7.x/avataaars/svg?seed=Ahmad"),
			University:   strPtr("Universitas Gadjah Mada"),
			Major:        strPtr("Sistem Informasi"),
			Bio:          strPtr("Information systems enthusiast with a passion for data analytics."),
			Phone:        strPtr("+62833445566"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     2100,
		},
	}

	for i := range users {
		if err := db.Create(&users[i]).Error; err != nil {
			log.Printf("Error creating user %s: %v", users[i].Email, err)
		}
	}

	return users
}

func seedCategories(db *gorm.DB) []models.Category {
	categories := []models.Category{
		{
			ID:          uuid.New(),
			Name:        "Web Development",
			Slug:        "web-development",
			Description: strPtr("Full-stack and frontend web applications"),
			Color:       strPtr("blue"),
		},
		{
			ID:          uuid.New(),
			Name:        "Mobile Apps",
			Slug:        "mobile-apps",
			Description: strPtr("iOS and Android applications"),
			Color:       strPtr("green"),
		},
		{
			ID:          uuid.New(),
			Name:        "AI & ML",
			Slug:        "ai-ml",
			Description: strPtr("Artificial Intelligence and Machine Learning projects"),
			Color:       strPtr("purple"),
		},
		{
			ID:          uuid.New(),
			Name:        "IoT",
			Slug:        "iot",
			Description: strPtr("Internet of Things and embedded systems"),
			Color:       strPtr("yellow"),
		},
		{
			ID:          uuid.New(),
			Name:        "Game Development",
			Slug:        "game-development",
			Description: strPtr("2D and 3D games"),
			Color:       strPtr("red"),
		},
	}

	for i := range categories {
		if err := db.Create(&categories[i]).Error; err != nil {
			log.Printf("Error creating category %s: %v", categories[i].Name, err)
		}
	}

	return categories
}

func seedProjects(db *gorm.DB, users []models.User, categories []models.Category) []models.Project {
	projects := []models.Project{
		{
			ID:           uuid.New(),
			UserID:       users[2].ID, // Budi
			CategoryID:   &categories[0].ID,
			Title:        "EcoTrack - Aplikasi Pengelolaan Sampah",
			Description:  strPtr("Progressive Web App untuk melacak dan mengelola pengumpulan sampah di area perkotaan. Dilengkapi pelacakan real-time, klasifikasi sampah berbasis AI, dan papan peringkat komunitas."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1532996122724-e3c354a0b15b?w=800&h=600&fit=crop"),
			TechStack:    pq.StringArray{"Next.js", "TypeScript", "Supabase", "TailwindCSS", "PostGIS"},
			GithubURL:    strPtr("https://github.com/example/ecotrack"),
			DemoURL:      strPtr("https://ecotrack-demo.vercel.app"),
			Type:         models.ProjectTypeFree,
			Price:        0,
			Status:       models.ProjectStatusPublished,
			Views:        1240,
			Likes:        89,
		},
		{
			ID:           uuid.New(),
			UserID:       users[3].ID, // Siti
			CategoryID:   &categories[0].ID,
			Title:        "StudyBuddy - Platform Belajar Kolaboratif",
			Description:  strPtr("Platform belajar kolaboratif real-time dengan video chat, papan tulis bersama, dan asisten belajar AI. Sempurna untuk pembelajaran jarak jauh dan proyek kelompok."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1522202176988-66273c2fd55f?w=800&h=600&fit=crop"),
			TechStack:    pq.StringArray{"React", "WebRTC", "Socket.io", "Node.js", "MongoDB"},
			GithubURL:    strPtr("https://github.com/example/studybuddy"),
			DemoURL:      strPtr("https://studybuddy-demo.netlify.app"),
			Type:         models.ProjectTypePaid,
			Price:        75000,
			Status:       models.ProjectStatusPublished,
			Views:        890,
			Likes:        67,
		},
		{
			ID:           uuid.New(),
			UserID:       users[4].ID, // Ahmad
			CategoryID:   &categories[1].ID,
			Title:        "SmartKampus - Navigasi Kampus",
			Description:  strPtr("Sistem navigasi indoor untuk kampus universitas menggunakan AR dan layanan lokasi real-time. Membantu mahasiswa menemukan ruang kelas, fasilitas, dan acara."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1562774053-701939374585?w=800&h=600&fit=crop"),
			TechStack:    pq.StringArray{"React Native", "AR.js", "Firebase", "Google Maps API"},
			GithubURL:    strPtr("https://github.com/example/smartkampus"),
			DemoURL:      strPtr("https://smartkampus.expo.dev"),
			Type:         models.ProjectTypeFree,
			Price:        0,
			Status:       models.ProjectStatusPublished,
			Views:        2100,
			Likes:        156,
		},
		{
			ID:           uuid.New(),
			UserID:       users[2].ID, // Budi
			CategoryID:   &categories[0].ID,
			Title:        "FoodShare - Berbagi Makanan Kampus",
			Description:  strPtr("Menghubungkan mahasiswa untuk berbagi makanan berlebih dan mengurangi limbah. Dilengkapi daftar real-time, sistem chat, dan penilaian reputasi."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1504674900247-0877df9cc836?w=800&h=600&fit=crop"),
			TechStack:    pq.StringArray{"Vue.js", "Express", "PostgreSQL", "Redis"},
			GithubURL:    strPtr("https://github.com/example/foodshare"),
			DemoURL:      strPtr("https://foodshare-campus.vercel.app"),
			Type:         models.ProjectTypePaid,
			Price:        50000,
			Status:       models.ProjectStatusPublished,
			Views:        1560,
			Likes:        112,
		},
		{
			ID:           uuid.New(),
			UserID:       users[3].ID, // Siti
			CategoryID:   &categories[2].ID,
			Title:        "CodeReview.AI - Analisis Kode Otomatis",
			Description:  strPtr("Asisten review kode berbasis AI yang memberikan feedback instan tentang kualitas kode, kerentanan keamanan, dan best practices."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1555066931-4365d14bab8c?w=800&h=600&fit=crop"),
			TechStack:    pq.StringArray{"Python", "FastAPI", "OpenAI", "Docker", "GitHub Actions"},
			GithubURL:    strPtr("https://github.com/example/codereview-ai"),
			DemoURL:      strPtr("https://codereview-ai.herokuapp.com"),
			Type:         models.ProjectTypePaid,
			Price:        150000,
			Status:       models.ProjectStatusPublished,
			Views:        3200,
			Likes:        245,
		},
		{
			ID:           uuid.New(),
			UserID:       users[4].ID, // Ahmad
			CategoryID:   &categories[0].ID,
			Title:        "EventHub - Manajer Acara Kampus",
			Description:  strPtr("Permudah pengorganisasian acara kampus dengan manajemen tiket, check-in kode QR, dan dashboard analitik untuk organisasi mahasiswa."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1540575467063-178a50c2df87?w=800&h=600&fit=crop"),
			TechStack:    pq.StringArray{"Next.js", "Prisma", "tRPC", "Stripe", "QR Code"},
			GithubURL:    strPtr("https://github.com/example/eventhub"),
			DemoURL:      strPtr("https://eventhub-campus.vercel.app"),
			Type:         models.ProjectTypePaid,
			Price:        100000,
			Status:       models.ProjectStatusPublished,
			Views:        980,
			Likes:        73,
		},
	}

	for i := range projects {
		if err := db.Create(&projects[i]).Error; err != nil {
			log.Printf("Error creating project %s: %v", projects[i].Title, err)
		}

		// Add project images
		images := []models.ProjectImage{
			{
				ID:        uuid.New(),
				ProjectID: projects[i].ID,
				ImageURL:  *projects[i].ThumbnailURL,
				SortOrder: 0,
			},
		}
		for j := range images {
			db.Create(&images[j])
		}
	}

	return projects
}

func seedArticles(db *gorm.DB, users []models.User) {
	now := time.Now()

	articles := []models.Article{
		{
			ID:           uuid.New(),
			UserID:       users[2].ID,
			Title:        "Tips Membuat Portofolio yang Menarik untuk Mahasiswa IT",
			Excerpt:      strPtr("Panduan lengkap membangun portofolio developer yang impresif dan menonjol di mata rekruter."),
			Content:      strPtr("# Tips Membuat Portofolio yang Menarik\n\nSebagai mahasiswa IT, memiliki portofolio yang kuat adalah kunci untuk menonjol di pasar kerja yang kompetitif.\n\n## 1. Pilih Proyek Terbaikmu\n\nFokuslah pada 3-5 proyek terbaik yang menunjukkan kemampuan dan kreativitasmu.\n\n## 2. Dokumentasi yang Baik\n\nSetiap proyek harus memiliki README yang jelas dengan deskripsi, teknologi yang digunakan, dan cara menjalankan proyek."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1517180102446-f3ece451e9d8?w=800&h=600&fit=crop"),
			Category:     strPtr("Karier"),
			ReadingTime:  8,
			Status:       models.ArticleStatusPublished,
			Views:        450,
			PublishedAt:  &now,
		},
		{
			ID:           uuid.New(),
			UserID:       users[3].ID,
			Title:        "Teknologi AI yang Wajib Dipelajari di 2025",
			Excerpt:      strPtr("Eksplorasi tren AI terbaru yang akan mendominasi industri teknologi tahun ini."),
			Content:      strPtr("# Teknologi AI yang Wajib Dipelajari\n\nArtificial Intelligence terus berkembang pesat. Berikut adalah teknologi AI yang wajib dipelajari.\n\n## 1. Large Language Models (LLM)\n\nSetelah sukses ChatGPT, pemahaman tentang LLM menjadi sangat penting.\n\n## 2. Generative AI\n\nTools seperti Midjourney, Stable Diffusion membuka peluang baru dalam kreatif industri."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1677442136019-21780ecad995?w=800&h=600&fit=crop"),
			Category:     strPtr("Teknologi"),
			ReadingTime:  6,
			Status:       models.ArticleStatusPublished,
			Views:        380,
			PublishedAt:  &now,
		},
		{
			ID:           uuid.New(),
			UserID:       users[4].ID,
			Title:        "Cara Efektif Mengelola Waktu sebagai Mahasiswa Developer",
			Excerpt:      strPtr("Strategi manajemen waktu untuk menyeimbangkan kuliah, proyek, dan kehidupan sosial."),
			Content:      strPtr("# Cara Efektif Mengelola Waktu\n\nMenyeimbangkan kuliah, projekt coding, dan kehidupan sosial bisa menjadi tantangan.\n\n## 1. Prioritaskan dengan Matriks Eisenhower\n\nKategorikan tugas berdasarkan urgency dan importance.\n\n## 2. Time Blocking\n\nAlokasikan waktu spesifik untuk coding, belajar, dan istirahat."),
			ThumbnailURL: strPtr("https://images.unsplash.com/photo-1484480974693-6ca0a78fb36b?w=800&h=600&fit=crop"),
			Category:     strPtr("Produktivitas"),
			ReadingTime:  5,
			Status:       models.ArticleStatusPublished,
			Views:        520,
			PublishedAt:  &now,
		},
	}

	for i := range articles {
		if err := db.Create(&articles[i]).Error; err != nil {
			log.Printf("Error creating article %s: %v", articles[i].Title, err)
		}
	}
}

func seedComments(db *gorm.DB, users []models.User, projects []models.Project) {
	comments := []models.Comment{
		{
			ID:        uuid.New(),
			UserID:    users[3].ID,
			ProjectID: projects[0].ID,
			Content:   "Ini luar biasa! Fitur klasifikasi sampahnya sangat mengesankan. Seberapa akurat model AI-nya?",
		},
		{
			ID:        uuid.New(),
			UserID:    users[4].ID,
			ProjectID: projects[0].ID,
			Content:   "Suka desain UI-nya! Sangat bersih dan intuitif. Semoga bisa diimplementasikan di kota saya.",
		},
		{
			ID:        uuid.New(),
			UserID:    users[2].ID,
			ProjectID: projects[1].ID,
			Content:   "Kerja bagus untuk implementasi WebRTC-nya! Apakah ada tantangan dengan NAT traversal?",
		},
		{
			ID:        uuid.New(),
			UserID:    users[2].ID,
			ProjectID: projects[2].ID,
			Content:   "Navigasi AR-nya super keren! Bagaimana cara menangani akurasi posisi indoor?",
		},
		{
			ID:        uuid.New(),
			UserID:    users[3].ID,
			ProjectID: projects[2].ID,
			Content:   "Ini akan sangat berguna untuk mahasiswa baru! Apakah sudah tersedia untuk diunduh?",
		},
	}

	for i := range comments {
		if err := db.Create(&comments[i]).Error; err != nil {
			log.Printf("Error creating comment: %v", err)
		}
	}
}

func strPtr(s string) *string {
	return &s
}
