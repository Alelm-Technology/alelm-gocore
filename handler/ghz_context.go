package handler

import "github.com/gin-gonic/gin"

type GhzContext struct {
	*gin.Context
}

func (c *GhzContext) UserID() string {
	v, _ := c.Get("user_id")
	s, _ := v.(string)
	return s
}

func (c *GhzContext) Username() string {
	v, _ := c.Get("username")
	s, _ := v.(string)
	return s
}

func (c *GhzContext) IsSuperAdmin() bool {
	v, _ := c.Get("is_super_admin")
	b, _ := v.(bool)
	return b
}
