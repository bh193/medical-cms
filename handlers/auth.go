package handlers

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

// 從token獲取使用者ID的函數
var jwtSecret = []byte("your_secret_key")

// 中間件：需要特定權限
func RequirePermission(db *sql.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 從 token 獲取使用者 ID
			userId := GetUserIDFromToken(c) 
			if userId == 0 {
				return // GetUserIDFromToken 已經設置了錯誤響應
		}

			if !HasPermission(db, userId, c.Request.URL.Path, c.Request.Method) {
					c.JSON(http.StatusForbidden, gin.H{"error": "無權限"})
					c.Abort()
					return
			}
			c.Next()
	}
}

// 是否有權限
func HasPermission(db *sql.DB, userId uint, path string, method string) bool {
	// 1. 獲取使用者角色
	roles, err := GetUserRoles(db, userId)
	if err != nil {
			log.Printf("Error getting user roles: %v", err)
			return false
	}

	// 2. 獲取該路徑或方法需要的權限
	requiredPermission := GetRequiredPermission(path, method)
	log.Printf("獲取所需要的權限", requiredPermission)

	// 3. 檢查使用者的角色是否有所需權限
	for _, roleId := range roles {
			permissions, err := GetRolePermissions(db, roleId)
			log.Printf("此使用者的權限", permissions)
			if err != nil {
					log.Printf("Error getting role permissions: %v", err)
					continue
			}
			if contains(permissions, requiredPermission) {
					return true
			}
	}
	return false
}

// 從token解析userId
func GetUserIDFromToken(c *gin.Context) uint {
    // 從cookie中獲取token
    tokenString, err := c.Cookie("jwt_token")
    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "token not found"})
        return 0
    }

    token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
        if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
            return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
        }
        return jwtSecret, nil
    })

    if err != nil {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return 0
    }

    if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
        userID, ok := claims["id"].(float64) 
        if !ok {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "id not found"})
            return 0
        }
        return uint(userID)
    } else {
        c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
        return 0
    }
}

// 獲得RoleId
func GetRolePermissions(db *sql.DB, roleID int) ([]string, error) {
	query := `
			SELECT p.Name 
			FROM permissions p
			JOIN role_permissions rp ON p.Id = rp.PermissionId
			WHERE rp.RoleId = ?
	`
	rows, err := db.Query(query, roleID)
	if err != nil {
			return nil, err
	}
	defer rows.Close()

	var permissions []string
	for rows.Next() {
			var permissionName string
			if err := rows.Scan(&permissionName); err != nil {
					return nil, err
			}
			permissions = append(permissions, permissionName)
	}
	return permissions, nil
}

// 根據 HTTP 方法決定操作類型：
// GET 方法對應 "read"
// POST 方法對應 "create"
// PUT 或 PATCH 方法對應 "update"
// DELETE 方法對應 "delete"
// 將操作類型和資源名稱組合成權限字符串。例如，GET 請求 "/api/users" 會生成 "read_users" 權限。
func GetRequiredPermission(path string, method string) string {
	parts := strings.Split(path, "/")
	resource := parts[len(parts)-1]
	switch method {
	case http.MethodGet:
			return "read_" + resource
	case http.MethodPost:
			return "create_" + resource
	case http.MethodPut, http.MethodPatch:
			return "update_" + resource
	case http.MethodDelete:
			return "delete_" + resource
	default:
			return "unknown_permission"
	}
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
			if s == item {
					return true
			}
	}
	return false
}