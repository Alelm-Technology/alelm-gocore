package ctxutil

import "github.com/gin-gonic/gin"

func SetUserID(c *gin.Context, id string) {
	c.Set("user_id", id)
}

func GetUserID(c *gin.Context) string {
	v, _ := c.Get("user_id")
	s, _ := v.(string)
	return s
}

func SetUsername(c *gin.Context, username string) {
	c.Set("username", username)
}

func GetUsername(c *gin.Context) string {
	v, _ := c.Get("username")
	s, _ := v.(string)
	return s
}

func SetIsSuperAdmin(c *gin.Context, is bool) {
	c.Set("is_super_admin", is)
}

func GetIsSuperAdmin(c *gin.Context) bool {
	v, _ := c.Get("is_super_admin")
	b, _ := v.(bool)
	return b
}

func Set[T any](c *gin.Context, key string, val T) {
	c.Set(key, val)
}

func Get[T any](c *gin.Context, key string) (T, bool) {
	v, exists := c.Get(key)
	if !exists {
		return *new(T), false
	}
	val, ok := v.(T)
	return val, ok
}
