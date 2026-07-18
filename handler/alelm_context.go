package handler

import "github.com/gin-gonic/gin"

type AlelmContext struct {
	*gin.Context
}

func (c *AlelmContext) UserID() string {
	v, ok := c.Get("user_id")
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func (c *AlelmContext) UserIDOK() (string, bool) {
	v, ok := c.Get("user_id")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func (c *AlelmContext) Username() string {
	v, ok := c.Get("username")
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func (c *AlelmContext) UsernameOK() (string, bool) {
	v, ok := c.Get("username")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}

func (c *AlelmContext) IsSuperAdmin() bool {
	v, ok := c.Get("is_super_admin")
	if !ok {
		return false
	}
	b, ok := v.(bool)
	if !ok {
		return false
	}
	return b
}

func (c *AlelmContext) IsSuperAdminOK() (bool, bool) {
	v, ok := c.Get("is_super_admin")
	if !ok {
		return false, false
	}
	b, ok := v.(bool)
	if !ok {
		return false, false
	}
	return b, true
}

func (c *AlelmContext) TenantID() string {
	v, ok := c.Get("tenant_id")
	if !ok {
		return ""
	}
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

func (c *AlelmContext) TenantIDOK() (string, bool) {
	v, ok := c.Get("tenant_id")
	if !ok {
		return "", false
	}
	s, ok := v.(string)
	if !ok {
		return "", false
	}
	return s, true
}
