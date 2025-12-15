package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/campus-project-hub/api/internal/config"
	"github.com/campus-project-hub/api/internal/database"
	"github.com/campus-project-hub/api/internal/models"
	"github.com/campus-project-hub/api/internal/utils"
	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/gorm"
)

// Asset paths (relative to cmd/seeder)
const assetsDir = "./cmd/seeder/assets"
const uploadsDir = "./uploads"

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

	// Drop all tables first
	log.Println("🗑️  Dropping existing tables...")
	dropTables(db)

	// Run GORM AutoMigrate
	log.Println("📊 Running GORM AutoMigrate...")
	if err := runAutoMigrate(db); err != nil {
		log.Fatalf("Failed to run AutoMigrate: %v", err)
	}

	// Create uploads directory if not exists
	if err := os.MkdirAll(uploadsDir, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

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

func dropTables(db *gorm.DB) {
	// Drop tables in reverse order of dependencies (junction tables first)
	tables := []string{
		"project_likes",
		"comments",
		"transactions",
		"project_images",
		"projects",
		"articles",
		"reports",
		"block_records",
		"users",
		"categories",
		"schema_migrations", // Also drop migrations table for fresh start
	}

	for _, table := range tables {
		if err := db.Exec(fmt.Sprintf("DROP TABLE IF EXISTS %s CASCADE", table)).Error; err != nil {
			log.Printf("Warning: Could not drop table %s: %v", table, err)
		}
	}
	log.Println("   Tables dropped successfully")
}

func runAutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Project{},
		&models.ProjectImage{},
		&models.ProjectLike{},
		&models.Article{},
		&models.Comment{},
		&models.Transaction{},
		&models.Report{},
		&models.BlockRecord{},
	)
}

// getAPIBaseURL returns the API base URL for constructing full image URLs
func getAPIBaseURL() string {
	cfg := config.GetConfig()
	if cfg.App.BaseURL != "" {
		return cfg.App.BaseURL
	}
	// Default to localhost with configured port
	return fmt.Sprintf("http://localhost:%d", cfg.App.Port)
}

// copyAssetToUploads copies a file from assets to uploads with a new UUID name
func copyAssetToUploads(assetName string) (string, error) {
	srcPath := filepath.Join(assetsDir, assetName)
	ext := filepath.Ext(assetName)
	newFilename := fmt.Sprintf("%s%s", uuid.New().String(), ext)
	dstPath := filepath.Join(uploadsDir, newFilename)

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(dstPath)
	if err != nil {
		return "", fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return "", fmt.Errorf("failed to copy file: %w", err)
	}

	// Return the full URL path that will be used by frontend
	return fmt.Sprintf("%s/uploads/%s", getAPIBaseURL(), newFilename), nil
}

func seedUsers(db *gorm.DB) []models.User {
	// Hash password for demo users
	hashedPassword, _ := utils.HashPassword("password123")

	// Copy avatar assets to uploads
	avatarAdmin, _ := copyAssetToUploads("avatar_admin.png")
	avatarModerator, _ := copyAssetToUploads("avatar_moderator.png")
	avatarGading, _ := copyAssetToUploads("avatar_gading.png")
	avatarArfan, _ := copyAssetToUploads("avatar_arfan.png")
	avatarRasyid, _ := copyAssetToUploads("avatar_rasyid.png")
	avatarMarvel, _ := copyAssetToUploads("avatar_marvel.png")
	avatarRifqi, _ := copyAssetToUploads("avatar_rifqi.png")

	users := []models.User{
		{
			ID:           uuid.New(),
			Email:        "admin@campus-hub.com",
			PasswordHash: &hashedPassword,
			Name:         "Admin User",
			AvatarURL:    strPtr(avatarAdmin),
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
			AvatarURL:    strPtr(avatarModerator),
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
			Email:        "gading@uty.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Gading",
			AvatarURL:    strPtr(avatarGading),
			University:   strPtr("Universitas Teknologi Yogyakarta"),
			Major:        strPtr("Ilmu Komputer"),
			Bio:          strPtr("Passionate developer specializing in web and mobile development."),
			Phone:        strPtr("+62811223344"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     1250,
		},
		{
			ID:           uuid.New(),
			Email:        "arfan@uty.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Arfan",
			AvatarURL:    strPtr(avatarArfan),
			University:   strPtr("Universitas Teknologi Yogyakarta"),
			Major:        strPtr("Teknik Perangkat Lunak"),
			Bio:          strPtr("Software engineer focused on clean code and best practices."),
			Phone:        strPtr("+62822334455"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     850,
		},
		{
			ID:           uuid.New(),
			Email:        "rasyid@uty.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Rasyid",
			AvatarURL:    strPtr(avatarRasyid),
			University:   strPtr("Universitas Teknologi Yogyakarta"),
			Major:        strPtr("Sistem Informasi"),
			Bio:          strPtr("Information systems enthusiast with a passion for data analytics."),
			Phone:        strPtr("+62833445566"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     2100,
		},
		{
			ID:           uuid.New(),
			Email:        "marvel@uty.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Marvel",
			AvatarURL:    strPtr(avatarMarvel),
			University:   strPtr("Universitas Teknologi Yogyakarta"),
			Major:        strPtr("Teknik Informatika"),
			Bio:          strPtr("Full-stack developer with expertise in modern web technologies."),
			Phone:        strPtr("+62844556677"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     1800,
		},
		{
			ID:           uuid.New(),
			Email:        "rifqi@uty.ac.id",
			PasswordHash: &hashedPassword,
			Name:         "Rifqi",
			AvatarURL:    strPtr(avatarRifqi),
			University:   strPtr("Universitas Teknologi Yogyakarta"),
			Major:        strPtr("Ilmu Komputer"),
			Bio:          strPtr("Backend developer passionate about system architecture."),
			Phone:        strPtr("+62855667788"),
			Role:         models.RoleUser,
			Status:       models.StatusActive,
			TotalExp:     1500,
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
	// Copy project assets to uploads
	projectEcotrack, _ := copyAssetToUploads("project_ecotrack.jpg")
	projectStudybuddy, _ := copyAssetToUploads("project_studybuddy.jpg")
	projectSmartkampus, _ := copyAssetToUploads("project_smartkampus.jpg")
	projectFoodshare, _ := copyAssetToUploads("project_foodshare.jpg")
	projectCodereview, _ := copyAssetToUploads("project_codereview.jpg")
	projectEventhub, _ := copyAssetToUploads("project_eventhub.jpg")

	projects := []models.Project{
		{
			ID:           uuid.New(),
			UserID:       users[2].ID, // Gading
			CategoryID:   &categories[0].ID,
			Title:        "EcoTrack - Aplikasi Pengelolaan Sampah",
			Description:  strPtr("Progressive Web App untuk melacak dan mengelola pengumpulan sampah di area perkotaan. Dilengkapi pelacakan real-time, klasifikasi sampah berbasis AI, dan papan peringkat komunitas."),
			ThumbnailURL: strPtr(projectEcotrack),
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
			UserID:       users[3].ID, // Arfan
			CategoryID:   &categories[0].ID,
			Title:        "StudyBuddy - Platform Belajar Kolaboratif",
			Description:  strPtr("Platform belajar kolaboratif real-time dengan video chat, papan tulis bersama, dan asisten belajar AI. Sempurna untuk pembelajaran jarak jauh dan proyek kelompok."),
			ThumbnailURL: strPtr(projectStudybuddy),
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
			UserID:       users[4].ID, // Rasyid
			CategoryID:   &categories[1].ID,
			Title:        "SmartKampus - Navigasi Kampus",
			Description:  strPtr("Sistem navigasi indoor untuk kampus universitas menggunakan AR dan layanan lokasi real-time. Membantu mahasiswa menemukan ruang kelas, fasilitas, dan acara."),
			ThumbnailURL: strPtr(projectSmartkampus),
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
			UserID:       users[5].ID, // Marvel
			CategoryID:   &categories[0].ID,
			Title:        "FoodShare - Berbagi Makanan Kampus",
			Description:  strPtr("Menghubungkan mahasiswa untuk berbagi makanan berlebih dan mengurangi limbah. Dilengkapi daftar real-time, sistem chat, dan penilaian reputasi."),
			ThumbnailURL: strPtr(projectFoodshare),
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
			UserID:       users[6].ID, // Rifqi
			CategoryID:   &categories[2].ID,
			Title:        "CodeReview.AI - Analisis Kode Otomatis",
			Description:  strPtr("Asisten review kode berbasis AI yang memberikan feedback instan tentang kualitas kode, kerentanan keamanan, dan best practices."),
			ThumbnailURL: strPtr(projectCodereview),
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
			UserID:       users[2].ID, // Gading
			CategoryID:   &categories[0].ID,
			Title:        "EventHub - Manajer Acara Kampus",
			Description:  strPtr("Permudah pengorganisasian acara kampus dengan manajemen tiket, check-in kode QR, dan dashboard analitik untuk organisasi mahasiswa."),
			ThumbnailURL: strPtr(projectEventhub),
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

	// Copy article assets to uploads
	articlePortofolio, _ := copyAssetToUploads("article_portofolio.jpg")
	articleAI, _ := copyAssetToUploads("article_ai.jpg")
	articleWaktu, _ := copyAssetToUploads("article_waktu.jpg")

	articles := []models.Article{
		{
			ID:           uuid.New(),
			UserID:       users[2].ID, // Gading
			Title:        "Tips Membuat Portofolio yang Menarik untuk Mahasiswa IT",
			Excerpt:      strPtr("Panduan lengkap membangun portofolio developer yang impresif dan menonjol di mata rekruter."),
			Content:      strPtr("# Tips Membuat Portofolio yang Menarik\n\nSebagai mahasiswa IT, memiliki portofolio yang kuat adalah kunci untuk menonjol di pasar kerja yang kompetitif.\n\n## 1. Pilih Proyek Terbaikmu\n\nFokuslah pada 3-5 proyek terbaik yang menunjukkan kemampuan dan kreativitasmu.\n\n## 2. Dokumentasi yang Baik\n\nSetiap proyek harus memiliki README yang jelas dengan deskripsi, teknologi yang digunakan, dan cara menjalankan proyek."),
			ThumbnailURL: strPtr(articlePortofolio),
			Category:     strPtr("Karier"),
			ReadingTime:  8,
			Status:       models.ArticleStatusPublished,
			Views:        450,
			PublishedAt:  &now,
		},
		{
			ID:           uuid.New(),
			UserID:       users[3].ID, // Arfan
			Title:        "Teknologi AI yang Wajib Dipelajari di 2025",
			Excerpt:      strPtr("Eksplorasi tren AI terbaru yang akan mendominasi industri teknologi tahun ini."),
			Content:      strPtr("# Teknologi AI yang Wajib Dipelajari\n\nArtificial Intelligence terus berkembang pesat. Berikut adalah teknologi AI yang wajib dipelajari.\n\n## 1. Large Language Models (LLM)\n\nSetelah sukses ChatGPT, pemahaman tentang LLM menjadi sangat penting.\n\n## 2. Generative AI\n\nTools seperti Midjourney, Stable Diffusion membuka peluang baru dalam kreatif industri."),
			ThumbnailURL: strPtr(articleAI),
			Category:     strPtr("Teknologi"),
			ReadingTime:  6,
			Status:       models.ArticleStatusPublished,
			Views:        380,
			PublishedAt:  &now,
		},
		{
			ID:           uuid.New(),
			UserID:       users[4].ID, // Rasyid
			Title:        "Cara Efektif Mengelola Waktu sebagai Mahasiswa Developer",
			Excerpt:      strPtr("Strategi manajemen waktu untuk menyeimbangkan kuliah, proyek, dan kehidupan sosial."),
			Content:      strPtr("# Cara Efektif Mengelola Waktu\n\nMenyeimbangkan kuliah, projekt coding, dan kehidupan sosial bisa menjadi tantangan.\n\n## 1. Prioritaskan dengan Matriks Eisenhower\n\nKategorikan tugas berdasarkan urgency dan importance.\n\n## 2. Time Blocking\n\nAlokasikan waktu spesifik untuk coding, belajar, dan istirahat."),
			ThumbnailURL: strPtr(articleWaktu),
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
			UserID:    users[3].ID, // Arfan
			ProjectID: projects[0].ID,
			Content:   "Ini luar biasa! Fitur klasifikasi sampahnya sangat mengesankan. Seberapa akurat model AI-nya?",
		},
		{
			ID:        uuid.New(),
			UserID:    users[4].ID, // Rasyid
			ProjectID: projects[0].ID,
			Content:   "Suka desain UI-nya! Sangat bersih dan intuitif. Semoga bisa diimplementasikan di kota saya.",
		},
		{
			ID:        uuid.New(),
			UserID:    users[2].ID, // Gading
			ProjectID: projects[1].ID,
			Content:   "Kerja bagus untuk implementasi WebRTC-nya! Apakah ada tantangan dengan NAT traversal?",
		},
		{
			ID:        uuid.New(),
			UserID:    users[5].ID, // Marvel
			ProjectID: projects[2].ID,
			Content:   "Navigasi AR-nya super keren! Bagaimana cara menangani akurasi posisi indoor?",
		},
		{
			ID:        uuid.New(),
			UserID:    users[6].ID, // Rifqi
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
