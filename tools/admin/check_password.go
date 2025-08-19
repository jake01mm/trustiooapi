package main

import (
	"log"

	"trusioo_api/config"
	"trusioo_api/pkg/database"

	"golang.org/x/crypto/bcrypt"
)

type Admin struct {
	ID       int64   `db:"id"`
	Name     string  `db:"name"`
	Email    string  `db:"email"`
	Password string  `db:"password"`
	Phone    *string `db:"phone"`
	ImageKey string  `db:"image_key"`
	Role     string  `db:"role"`
	IsSuper  bool    `db:"is_super"`
	Status   string  `db:"status"`
}

func main() {
	log.Println("Checking admin account...")

	// 加载配置
	if err := config.LoadConfig(); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// 初始化数据库连接
	if err := database.InitDatabase(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.CloseDatabase()

	// 查询管理员账户
	var admin Admin
	err := database.DB.Get(&admin, `
		SELECT id, name, email, password, phone, image_key, role, is_super, status 
		FROM admins WHERE email = 'admin@trusioo.com'
	`)
	if err != nil {
		log.Fatalf("Failed to find admin: %v", err)
	}

	log.Printf("Admin found: %+v", admin)

	// 测试密码验证
	err = bcrypt.CompareHashAndPassword([]byte(admin.Password), []byte("admin123"))
	if err != nil {
		log.Printf("Password verification failed: %v", err)
		
		// 生成正确的密码哈希
		newHash, err := bcrypt.GenerateFromPassword([]byte("admin123"), bcrypt.DefaultCost)
		if err != nil {
			log.Fatalf("Failed to generate password hash: %v", err)
		}
		log.Printf("Correct hash for 'admin123': %s", string(newHash))
		
		// 更新密码
		_, err = database.DB.Exec("UPDATE admins SET password = $1 WHERE email = 'admin@trusioo.com'", string(newHash))
		if err != nil {
			log.Fatalf("Failed to update password: %v", err)
		}
		log.Println("Password updated successfully!")
	} else {
		log.Println("Password verification successful!")
	}
}