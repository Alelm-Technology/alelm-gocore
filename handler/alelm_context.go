package handler

import "github.com/gin-gonic/gin"

type AlelmContext struct {
	*gin.Context
}

func (c *AlelmContext) UserID() string {
	v, _ := c.Get("user_id")
	s, _ := v.(string)
	return s
}

func (c *AlelmContext) Username() string {
	v, _ := c.Get("username")
	s, _ := v.(string)
	return s
}

func (c *AlelmContext) IsSuperAdmin() bool {
	v, _ := c.Get("is_super_admin")
	b, _ := v.(bool)
	return b
}

func (c *AlelmContext) TenantID() string {
	v, _ := c.Get("tenant_id")
	s, _ := v.(string)
	return s
}
