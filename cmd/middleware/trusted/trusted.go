package trusted

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

// TrustedIPMiddleware создает middleware для проверки IP-адреса клиента
// против доверенной подсети. Если IP не входит в доверенную подсеть,
// возвращает статус 403 Forbidden.
//
// Параметры:
//   - trustedSubnet: строка CIDR для доверенной подсети (например, "192.168.1.0/24")
//
// Возвращает:
//   - gin.HandlerFunc: middleware функция для Gin
//
// Логика:
//   - Извлекает IP-адрес из заголовка X-Real-IP
//   - Если заголовок отсутствует, использует RemoteAddr
//   - Проверяет, входит ли IP в доверенную подсеть
//   - Если подсеть пустая, запрещает доступ для всех запросов
//   - Если IP не входит в подсеть, возвращает 403 Forbidden
//
// Пример:
//
//	router.Use(trusted.TrustedIPMiddleware("192.168.1.0/24"))
func TrustedIPMiddleware(trustedSubnet string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if trustedSubnet == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		clientIP := getClientIP(c)
		if clientIP == "" {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		if !isIPInSubnet(clientIP, trustedSubnet) {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}

		c.Next()
	}
}

// getClientIP извлекает IP-адрес клиента из запроса.
// Сначала проверяет заголовок X-Real-IP, затем RemoteAddr.
//
// Параметры:
//   - c: контекст Gin
//
// Возвращает:
//   - string: IP-адрес клиента или пустую строку, если не удалось определить
func getClientIP(c *gin.Context) string {
	if realIP := c.GetHeader("X-Real-IP"); realIP != "" {
		return realIP
	}

	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		return ""
	}

	return ip
}

// isIPInSubnet проверяет, входит ли IP-адрес в указанную подсеть.
//
// Параметры:
//   - ip: IP-адрес для проверки
//   - subnet: строка CIDR подсети (например, "192.168.1.0/24")
//
// Возвращает:
//   - bool: true, если IP входит в подсеть, false в противном случае
//
// Пример:
//
//	isInSubnet := isIPInSubnet("192.168.1.100", "192.168.1.0/24")
//	// Возвращает true
func isIPInSubnet(ip, subnet string) bool {
	clientIP := net.ParseIP(ip)
	if clientIP == nil {
		return false
	}

	_, network, err := net.ParseCIDR(subnet)
	if err != nil {
		return false
	}

	return network.Contains(clientIP)
}
