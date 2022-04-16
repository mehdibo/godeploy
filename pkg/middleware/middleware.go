package middleware

import (
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
)

func RequestLog(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		if err := next(c); err != nil {
			c.Error(err)
		}

		method := c.Request().Method
		uri := c.Request().RequestURI
		status := c.Response().Status

		log.Infof("%s %s %d", method, uri, status)
		return nil
	}
}
