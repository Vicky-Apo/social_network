package handler

import "social-network/backend/pkg/logger"

func logBadRequest(log logger.Logger, action string, fields ...logger.Field) {
	log.Debug("bad request", append(fields, logger.F("action", action))...)
}

func logUnauthorized(log logger.Logger, action string, fields ...logger.Field) {
	log.Debug("unauthorized", append(fields, logger.F("action", action))...)
}

func logForbidden(log logger.Logger, action string, fields ...logger.Field) {
	log.Debug("forbidden", append(fields, logger.F("action", action))...)
}

func logNotFound(log logger.Logger, action string, fields ...logger.Field) {
	log.Debug("not found", append(fields, logger.F("action", action))...)
}

func logServerError(log logger.Logger, action string, err error, fields ...logger.Field) {
	log.Error("request failed", err, append(fields, logger.F("action", action))...)
}
