package handlers

import (
	"database/sql"
	"errors"

	"medical-cms/database"
	"medical-cms/models"
)

// 資料庫是否有此email
func GetUserByEmail(email string) (*models.User, error) {
	if database.DB == nil {
        return nil, errors.New("database connection not initialized")
    }

    user := &models.User{}
    err := database.DB.QueryRow("SELECT Id, Email, Name, Picture FROM users WHERE email = ?", email).
        Scan(&user.Id, &user.Email, &user.Name, &user.Picture)

    if err != nil {
        if err == sql.ErrNoRows {
            return nil, err // 使用者不存在
        }
        return nil, err // 其他資料庫錯誤
    }
    return user, nil
}

// 更新使用者資訊
func UpdateUser(user *models.User) error {
    _, err := database.DB.Exec("UPDATE users SET Name = ?, Picture = ?",
        user.Name, user.Picture)
    return err
}

// 獲取使用者角色
func GetUserRoles(db *sql.DB, userID uint) ([]int, error) {
	query := `
			SELECT roleId 
			FROM user_roles 
			WHERE userId = ?
	`
	rows, err := db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var roles []int
	for rows.Next() {
		var roleID int
		if err := rows.Scan(&roleID); err != nil {
			return nil, err
		}
		roles = append(roles, roleID)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// func CreateUser(c *gin.Context) {
//     var user models.User
//     if err := c.BindJSON(&user); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//         return
//     }
//     users[user.ID] = user
//     c.JSON(http.StatusCreated, user)
// }

// func GetUser(c *gin.Context) {
//     id := c.Param("id")
//     user, exists := users[id]
//     if !exists {
//         c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//         return
//     }
//     c.JSON(http.StatusOK, user)
// }

// func UpdateUser(c *gin.Context) {
//     id := c.Param("id")
//     _, exists := users[id]
//     if !exists {
//         c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//         return
//     }
//     var user models.User
//     if err := c.BindJSON(&user); err != nil {
//         c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
//         return
//     }
//     users[id] = user
//     c.JSON(http.StatusOK, user)
// }

// func DeleteUser(c *gin.Context) {
//     id := c.Param("id")
//     _, exists := users[id]
//     if !exists {
//         c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
//         return
//     }
//     delete(users, id)
//     c.Status(http.StatusNoContent)
// }