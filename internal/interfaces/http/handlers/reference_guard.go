package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"saas_baseon_go/internal/interfaces/http/response"
)

type deletionReference struct {
	Model interface{}
	Name  string
	Query string
	Args  []interface{}
}

type ReferenceGuard struct {
	db *gorm.DB
}

func NewReferenceGuard(db *gorm.DB) ReferenceGuard {
	return ReferenceGuard{db: db}
}

func ref(model interface{}, name, query string, args ...interface{}) deletionReference {
	return deletionReference{Model: model, Name: name, Query: query, Args: args}
}

func (g ReferenceGuard) BlockIfReferenced(c *gin.Context, resource string, refs ...deletionReference) bool {
	for _, item := range refs {
		var count int64
		if err := g.db.Model(item.Model).Where(item.Query, item.Args...).Count(&count).Error; err != nil {
			response.Error(c, 400, response.CodeBadRequest, "引用检查失败")
			return true
		}
		if count > 0 {
			response.Error(c, 400, response.CodeBadRequest, fmt.Sprintf("%s已被%s引用，不能删除", resource, item.Name))
			return true
		}
	}
	return false
}
