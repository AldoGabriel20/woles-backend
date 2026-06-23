package http_fiber

import (
	appchat "github.com/woles/woles-backend/internal/application/chat"
	appdocument "github.com/woles/woles-backend/internal/application/document"
	appfamily "github.com/woles/woles-backend/internal/application/family"
	appgoal "github.com/woles/woles-backend/internal/application/goal"
	appidentity "github.com/woles/woles-backend/internal/application/identity"
	appnotification "github.com/woles/woles-backend/internal/application/notification"
	appreminder "github.com/woles/woles-backend/internal/application/reminder"
	appsubscription "github.com/woles/woles-backend/internal/application/subscription"
	apptimeline "github.com/woles/woles-backend/internal/application/timeline"
	"github.com/woles/woles-backend/internal/port/outbound/database"
	"github.com/woles/woles-backend/internal/port/outbound/storage"
)

// Services bundles all application-layer services passed to HTTP handlers.
type Services struct {
	Identity     *appidentity.Service
	Reminder     *appreminder.Service
	Document     *appdocument.Service
	Subscription *appsubscription.Service
	Goal         *appgoal.Service
	Timeline     *apptimeline.Service
	Notification *appnotification.Service
	Family       *appfamily.Service
	Chat         *appchat.Service

	// Repositories exposed directly for account/billing handlers.
	Users       database.UserRepository
	UsageLimits database.UsageLimitRepository
	FileStore   storage.FileStore
}
