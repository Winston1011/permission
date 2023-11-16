package command

import (
	"permission/models/demo"

	"github.com/gin-gonic/gin"
)

func DemoJob1(ctx *gin.Context) error {
	_, err := demo.GetDemoByName(ctx, []string{"permission"})
	if err != nil {
		return err
	}

	return nil
}
