package store

// Storage is the storage interface used by the server. Implementations are
// required to be able to perform atomic compare-and-swap updates and either
// support timezones or standardize on UTC.
type Storage interface {
	Close() error

	// TODO(ericchiang): Let the storages set the IDs of these objects.
	CreateAuthRequest(a AuthRequest) error
	CreateClient(c Client) error
	CreateAuthCode(c AuthCode) error
	CreateRefresh(r RefreshToken) error
	CreatePassword(p Password) error
	CreateOfflineSessions(s OfflineSessions) error
	CreateConnector(c Connector) error
	CreateDeviceRequest(d DeviceRequest) error
	CreateDeviceToken(d DeviceToken) error

	// TODO(ericchiang): return (T, bool, error) so we can indicate not found
	// requests that way instead of using ErrNotFound.
	GetAuthRequest(id string) (AuthRequest, error)
	GetAuthCode(id string) (AuthCode, error)
	GetClient(id string) (Client, error)
	GetKeys() (Keys, error)
	GetRefresh(id string) (RefreshToken, error)
	GetPassword(email string) (Password, error)
	GetOfflineSessions(userID string, connID string) (OfflineSessions, error)
	GetConnector(id string) (Connector, error)
	GetDeviceRequest(userCode string) (DeviceRequest, error)
	GetDeviceToken(deviceCode string) (DeviceToken, error)

	ListClients() ([]Client, error)
	ListRefreshTokens() ([]RefreshToken, error)
	ListPasswords() ([]Password, error)
	ListConnectors() ([]Connector, error)

	// Delete methods MUST be atomic.
	DeleteAuthRequest(id string) error
	DeleteAuthCode(code string) error
	DeleteClient(id string) error
	DeleteRefresh(id string) error
	DeletePassword(email string) error
	DeleteOfflineSessions(userID string, connID string) error
	DeleteConnector(id string) error

	// Update methods take a function for updating an object then performs that update within
	// a transaction. "updater" functions may be called multiple times by a single update call.
	//
	// Because new fields may be added to resources, updaters should only modify existing
	// fields on the old object rather then creating new structs. For example:
	//
	//		updater := func(old storage.Client) (storage.Client, error) {
	//			old.Secret = newSecret
	//			return old, nil
	//		}
	//		if err := s.UpdateClient(clientID, updater); err != nil {
	//			// update failed, handle error
	//		}
	//
	UpdateClient(id string, updater func(old Client) (Client, error)) error
	UpdateKeys(updater func(old Keys) (Keys, error)) error
	UpdateAuthRequest(id string, updater func(a AuthRequest) (AuthRequest, error)) error
	UpdateRefreshToken(id string, updater func(r RefreshToken) (RefreshToken, error)) error
	UpdatePassword(email string, updater func(p Password) (Password, error)) error
	UpdateOfflineSessions(userID string, connID string, updater func(s OfflineSessions) (OfflineSessions, error)) error
	UpdateConnector(id string, updater func(c Connector) (Connector, error)) error
	UpdateDeviceToken(deviceCode string, updater func(t DeviceToken) (DeviceToken, error)) error

	// GarbageCollect deletes all expired AuthCodes,
	// AuthRequests, DeviceRequests, and DeviceTokens.
	GarbageCollect(now time.Time) (GCResult, error)
}
