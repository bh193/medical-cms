package handlers

import (
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"

	"medical-cms/config"
	"medical-cms/models"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
)

func GoogleLogin(c *gin.Context) {
    url := config.GoogleOauthConfig.AuthCodeURL(config.OauthStateString)
    c.Redirect(http.StatusTemporaryRedirect, url)
}

// google callback
func GoogleCallback(c *gin.Context) {
    state := c.Query("state")
    if state != config.OauthStateString {
        c.String(http.StatusUnauthorized, "Invalid oauth state")
        return
    }

    code := c.Query("code")
    token, err := config.GoogleOauthConfig.Exchange(c, code)
    if err != nil {
        c.String(http.StatusBadRequest, "Code exchange failed: %s", err.Error())
        return
    }

    response, err := http.Get("https://www.googleapis.com/oauth2/v2/userinfo?access_token=" + token.AccessToken)
    if err != nil {
        c.String(http.StatusBadRequest, "Failed getting user info: %s", err.Error())
        return
    }
    defer response.Body.Close()

    contents, err := io.ReadAll(response.Body)
    if err != nil {
        c.String(http.StatusInternalServerError, "Failed reading response body: %s", err.Error())
        return
    }
    log.Printf("Google個人資訊: %s", string(contents))

    var googleUser struct {
        Email         string `json:"email"`
        Name          string `json:"name"`
        Picture       string `json:"picture"`
    }

    if err := json.Unmarshal(contents, &googleUser); err != nil {
        c.String(http.StatusInternalServerError, "Failed parsing user info: %s", err.Error())
        return
    }
     
    // 檢查使用者是否已經存在於資料庫
    user, err := GetUserByEmail(googleUser.Email)
    log.Printf("找到使用者了 - User: %+v, Error: %v", user, err)

    if err != nil {
        if err == sql.ErrNoRows {
            log.Printf("找不到使用者: %s", googleUser.Email)
            c.JSON(http.StatusUnauthorized, gin.H{"error": "使用者未註冊"})
        } else {
            log.Printf("資料庫錯誤: %s", err.Error())
            c.String(http.StatusInternalServerError, "資料庫錯誤: %s", err.Error())
        }
        return
    }

    // 更新使用者的Google資訊
    user.Name = googleUser.Name
    user.Picture = &googleUser.Picture
    if err := UpdateUser(user); err != nil {
        c.String(http.StatusInternalServerError, "使用者更新資訊失敗: %s", err.Error())
        return
    }

    // 產生token
    jwtToken, err := GenerateToken(user)
    if err != nil {
        c.String(http.StatusInternalServerError, "token產生失敗: %s", err.Error())
        return
    }

    // 設置 cookie
    c.SetCookie("jwt_token", jwtToken, 3600*24, "/", "", false, true)

    c.JSON(http.StatusOK, gin.H{"message": "Login successful", "user": user.Name})
    // c.JSON(http.StatusOK, gin.H{"token": jwtToken, "user": user})
}

// 產生token
func GenerateToken(user *models.User) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "id":    user.Id,
        "email": user.Email,
        "exp":   time.Now().Add(time.Hour * 24).Unix(), // Token 有效期為 24 小時
    })

    return token.SignedString([]byte("your_secret_key"))
}

// 登入驗證中間函式
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 從 cookie 中獲取 token
        token, err := c.Cookie("jwt_token")
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization cookie is missing"})
            c.Abort()
            return
        }

        // 驗證 token
        claims, err := ValidateToken(token)
        if err != nil {
            c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid or expired token"})
            c.Abort()
            return
        }

        // 將使用者信息添加到上下文中
        c.Set("userID", claims.Id)
        c.Next()
    }
}

// 驗證token
func ValidateToken(tokenString string) (*jwt.StandardClaims, error) {
    token, err := jwt.ParseWithClaims(tokenString, &jwt.StandardClaims{}, func(token *jwt.Token) (interface{}, error) {
        return []byte("your_secret_key"), nil
    })

    if err != nil {
        return nil, err
    }

    if claims, ok := token.Claims.(*jwt.StandardClaims); ok && token.Valid {
        return claims, nil
    }

    return nil, errors.New("invalid token")
}